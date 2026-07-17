import type { KeyboardEvent } from 'react';
import { Button, InputNumber, Typography, message } from 'antd';
import { PlusOutlined } from '@ant-design/icons';
import { useReceptionStore } from '../store/receptionStore';
import { lineCashAmount } from '../types';
import { ServiceRow } from './ServiceRow';

type ServicesTableProps = {
  onRecalculate: () => void;
};

/** قالب‌بندی عدد به فارسی */
function formatMoney(value: number): string {
  return Number(value ?? 0).toLocaleString('fa-IR');
}

/** جدول خدمات پذیرش با افزودن/حذف سطر، جمع‌ها و صندوق */
export function ServicesTable({ onRecalculate }: ServicesTableProps) {
  const editing = useReceptionStore((s) => s.editing);
  const services = useReceptionStore((s) => s.services);
  const discount = useReceptionStore((s) => s.discount);
  const setDiscount = useReceptionStore((s) => s.setDiscount);
  const hasOrganization = useReceptionStore((s) => s.hasOrganization);
  const requestInsuranceFocus = useReceptionStore((s) => s.requestInsuranceFocus);
  const addServiceLine = useReceptionStore((s) => s.addServiceLine);
  const removeServiceLine = useReceptionStore((s) => s.removeServiceLine);
  const updateServiceLine = useReceptionStore((s) => s.updateServiceLine);

  const totals = services.reduce(
    (acc, line) => {
      acc.amount += Number(line.service_amount ?? 0);
      acc.tariff += Number(line.service_tariff ?? 0);
      acc.orgShare += Number(line.service_organization_share ?? 0);
      acc.suppShare += Number(line.service_supplementary_insurance_share ?? 0);
      acc.subsidy += Number(line.service_subsidy_share ?? 0);
      acc.cashBeforeDiscount += lineCashAmount(line);
      return acc;
    },
    { amount: 0, tariff: 0, orgShare: 0, suppShare: 0, subsidy: 0, cashBeforeDiscount: 0 },
  );
  const cashTotal = Math.max(0, totals.cashBeforeDiscount - Number(discount ?? 0));

  /** قبل از ورود به خدمات، حداقل یک سازمان باید انتخاب شده باشد */
  function ensureOrganization(): boolean {
    if (hasOrganization()) return true;
    message.warning('لطفاً قبل از انتخاب خدمت، یک سازمان (پایه یا تکمیلی) انتخاب کنید.');
    requestInsuranceFocus();
    return false;
  }

  /** افزودن سطر خدمت جدید (دکمه، Ctrl+Space، یا Enter/Tab فقط روی توضیحات آخرین سطر) */
  function tryAddService() {
    if (!editing) return;
    if (!ensureOrganization()) return;
    addServiceLine();
  }

  /** میانبر Ctrl+Space در محدوده جدول → همیشه خدمت جدید (مستقل از موقعیت فوکوس) */
  function handleTableKeyDown(e: KeyboardEvent) {
    if (!editing) return;
    if (e.ctrlKey && e.code === 'Space') {
      e.preventDefault();
      e.stopPropagation();
      tryAddService();
    }
  }

  return (
    <div onKeyDown={handleTableKeyDown}>
      <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 8 }}>
        <Typography.Title level={5} style={{ margin: 0 }}>
          خدمات
        </Typography.Title>
        {editing && (
          <Button type="primary" icon={<PlusOutlined />} onClick={tryAddService}>
            افزودن خدمت
          </Button>
        )}
      </div>
      <div className="reception-services-layout">
        <div style={{ overflowX: 'auto', flex: 1, minWidth: 0 }}>
          <table style={{ width: '100%', borderCollapse: 'collapse' }} className="reception-services">
            <thead>
              <tr>
                <th className="col-service">کد خدمت</th>
                <th className="col-qty">تعداد</th>
                <th className="col-direction">جهت دندان</th>
                <th className="col-tooth">شماره دندان</th>
                <th className="col-amount">نرخ</th>
                <th className="col-tariff">تعرفه</th>
                <th className="col-org">سهم سازمان</th>
                <th className="col-supp">سهم تکمیلی</th>
                <th className="col-subsidy">یارانه</th>
                <th className="col-cash">صندوق</th>
                <th className="col-desc">توضیحات</th>
                <th />
              </tr>
            </thead>
            <tbody>
              {services.length === 0 ? (
                <tr>
                  <td colSpan={12} style={{ textAlign: 'center', padding: 16, color: '#888' }}>
                    خدمتی ثبت نشده است
                  </td>
                </tr>
              ) : (
                services.map((line, index) => {
                  const isLastRow = index === services.length - 1;
                  const nextLineKey = isLastRow ? null : services[index + 1]!.key;
                  return (
                    <ServiceRow
                      key={line.key}
                      line={line}
                      editing={editing}
                      cashAmount={lineCashAmount(line)}
                      isLastRow={isLastRow}
                      nextLineKey={nextLineKey}
                      ensureOrganization={ensureOrganization}
                      onChange={(patch) => updateServiceLine(line.key, patch)}
                      onRemove={() => removeServiceLine(line.key)}
                      onRecalculate={onRecalculate}
                      onAddNextLine={tryAddService}
                    />
                  );
                })
              )}
            </tbody>
          </table>
        </div>

        <aside className="reception-services-totals" aria-label="جمع خدمات">
          <Typography.Text strong style={{ display: 'block', marginBottom: 12 }}>
            جمع کل
          </Typography.Text>
          <dl className="reception-totals-list">
            <div className="tot-amount">
              <dt>جمع نرخ</dt>
              <dd>{formatMoney(totals.amount)}</dd>
            </div>
            <div className="tot-tariff">
              <dt>جمع تعرفه</dt>
              <dd>{formatMoney(totals.tariff)}</dd>
            </div>
            <div className="tot-org">
              <dt>جمع سهم سازمان</dt>
              <dd>{formatMoney(totals.orgShare)}</dd>
            </div>
            <div className="tot-supp">
              <dt>جمع سهم تکمیلی</dt>
              <dd>{formatMoney(totals.suppShare)}</dd>
            </div>
            <div className="tot-subsidy">
              <dt>جمع یارانه</dt>
              <dd>{formatMoney(totals.subsidy)}</dd>
            </div>
            <div>
              <dt>تخفیف</dt>
              <dd>
                <InputNumber
                  min={0}
                  disabled={!editing}
                  value={discount}
                  style={{ width: '100%' }}
                  onChange={(v) => setDiscount(Number(v) || 0)}
                />
              </dd>
            </div>
            <div className="reception-totals-cash tot-cash">
              <dt>صندوق</dt>
              <dd>{formatMoney(cashTotal)}</dd>
            </div>
          </dl>
        </aside>
      </div>
      <style>{`
        .reception-services-layout {
          display: flex;
          gap: 16px;
          align-items: flex-start;
        }
        .reception-services th, .reception-services td {
          border-bottom: 1px solid #e8e8e8;
          padding: 6px 8px;
          text-align: right;
          vertical-align: middle;
        }
        .reception-services th {
          white-space: nowrap;
          font-size: 12px;
          font-weight: 600;
        }
        .reception-services td.num {
          font-variant-numeric: tabular-nums;
          direction: ltr;
          text-align: left;
          white-space: nowrap;
        }
        .reception-services .cell-muted {
          color: #bfbfbf;
          font-size: 12px;
        }
        .reception-services tr.reception-service-negative td {
          background: #fff1f0 !important;
          color: #cf1322;
        }

        /* رنگ‌های معنایی ستون‌ها */
        .reception-services th.col-service,
        .reception-services td.col-service { background: #f5f7fa; }
        .reception-services th.col-qty { background: #e6f4ff; color: #0958d9; }
        .reception-services td.col-qty { background: #f0f7ff; }
        .reception-services th.col-direction { background: #f9f0ff; color: #531dab; }
        .reception-services td.col-direction { background: #fbf5ff; }
        .reception-services th.col-tooth { background: #fff0f6; color: #c41d7f; }
        .reception-services td.col-tooth { background: #fff7fa; }
        .reception-services th.col-amount { background: #e6fffb; color: #08979c; }
        .reception-services td.col-amount { background: #f0fffc; color: #08979c; font-weight: 600; }
        .reception-services th.col-tariff { background: #f0f5ff; color: #1d39c4; }
        .reception-services td.col-tariff { background: #f5f8ff; color: #1d39c4; }
        .reception-services th.col-org { background: #e6f7ff; color: #096dd9; }
        .reception-services td.col-org { background: #f0faff; color: #096dd9; }
        .reception-services th.col-supp { background: #f9f0ff; color: #722ed1; }
        .reception-services td.col-supp { background: #fbf5ff; color: #722ed1; }
        .reception-services th.col-subsidy { background: #fff7e6; color: #d46b08; }
        .reception-services td.col-subsidy { background: #fffaf0; color: #d46b08; }
        .reception-services th.col-cash { background: #f6ffed; color: #389e0d; }
        .reception-services td.col-cash { background: #fcfff8; color: #389e0d; font-weight: 700; }
        .reception-services th.col-desc,
        .reception-services td.col-desc { background: #fafafa; }

        .reception-services-totals {
          flex: 0 0 220px;
          border: 1px solid #e8e8e8;
          background: #fafafa;
          border-radius: 8px;
          padding: 12px 14px;
        }
        .reception-totals-list {
          margin: 0;
          display: flex;
          flex-direction: column;
          gap: 10px;
        }
        .reception-totals-list > div {
          display: flex;
          justify-content: space-between;
          align-items: center;
          gap: 8px;
        }
        .reception-totals-list dt {
          margin: 0;
          font-size: 13px;
        }
        .reception-totals-list dd {
          margin: 0;
          font-weight: 600;
          text-align: left;
          direction: ltr;
        }
        .tot-amount dt, .tot-amount dd { color: #08979c; }
        .tot-tariff dt, .tot-tariff dd { color: #1d39c4; }
        .tot-org dt, .tot-org dd { color: #096dd9; }
        .tot-supp dt, .tot-supp dd { color: #722ed1; }
        .tot-subsidy dt, .tot-subsidy dd { color: #d46b08; }
        .reception-totals-cash {
          border-top: 1px solid #e8e8e8;
          padding-top: 10px;
          margin-top: 2px;
        }
        .tot-cash dt,
        .tot-cash dd {
          color: #389e0d;
          font-size: 15px;
        }
        @media (max-width: 960px) {
          .reception-services-layout {
            flex-direction: column;
          }
          .reception-services-totals {
            flex: 1 1 auto;
            width: 100%;
          }
        }
      `}</style>
    </div>
  );
}
