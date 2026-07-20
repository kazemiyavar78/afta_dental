import type { KeyboardEvent } from 'react';
import { Button, Card, Col, Flex, InputNumber, Row, Statistic, Typography, message } from 'antd';
import { PlusOutlined } from '@ant-design/icons';
import { useReceptionStore } from '../store/receptionStore';
import { lineCashAmount } from '../types';
import { ServiceRow } from './ServiceRow';

type ServicesTableProps = {
  onRecalculate: () => void;
};

function formatMoney(value: number): string {
  return Number(value ?? 0).toLocaleString('fa-IR');
}

/** جدول خدمات فشرده با هدر sticky و جمع ثابت در پایین */
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

  function ensureOrganization(): boolean {
    if (hasOrganization()) return true;
    message.warning('لطفاً قبل از انتخاب خدمت، یک سازمان (پایه یا تکمیلی) انتخاب کنید.');
    requestInsuranceFocus();
    return false;
  }

  function tryAddService() {
    if (!editing) return;
    if (!ensureOrganization()) return;
    addServiceLine();
  }

  function handleTableKeyDown(e: KeyboardEvent) {
    if (!editing) return;
    if (e.ctrlKey && e.code === 'Space') {
      e.preventDefault();
      e.stopPropagation();
      tryAddService();
    }
  }

  return (
    <Flex vertical gap={8} onKeyDown={handleTableKeyDown}>
      <Flex justify="space-between" align="center">
        <Typography.Text strong>خدمات</Typography.Text>
        {editing && (
          <Button type="primary" size="small" icon={<PlusOutlined />} onClick={tryAddService}>
            افزودن خدمت
          </Button>
        )}
      </Flex>

      <div className="reception-services-scroll">
        <table className="reception-services ant-table">
          <thead>
            <tr>
              <th>کد خدمت</th>
              <th>تعداد</th>
              <th>جهت</th>
              <th>دندان</th>
              <th>نرخ</th>
              <th>تعرفه</th>
              <th>سازمان</th>
              <th>تکمیلی</th>
              <th>یارانه</th>
              <th>صندوق</th>
              <th>توضیحات</th>
              <th />
            </tr>
          </thead>
          <tbody>
            {services.length === 0 ? (
              <tr>
                <td colSpan={12} style={{ textAlign: 'center', padding: 12, color: '#888' }}>
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

      <Card size="small" styles={{ body: { padding: '8px 12px' } }}>
        <Row gutter={[12, 8]} align="middle">
          <Col xs={12} sm={8} md={4}>
            <Statistic title="جمع نرخ" value={formatMoney(totals.amount)} valueStyle={{ fontSize: 14 }} />
          </Col>
          <Col xs={12} sm={8} md={4}>
            <Statistic title="جمع تعرفه" value={formatMoney(totals.tariff)} valueStyle={{ fontSize: 14 }} />
          </Col>
          <Col xs={12} sm={8} md={3}>
            <Statistic title="سازمان" value={formatMoney(totals.orgShare)} valueStyle={{ fontSize: 14 }} />
          </Col>
          <Col xs={12} sm={8} md={3}>
            <Statistic title="تکمیلی" value={formatMoney(totals.suppShare)} valueStyle={{ fontSize: 14 }} />
          </Col>
          <Col xs={12} sm={8} md={3}>
            <Statistic title="یارانه" value={formatMoney(totals.subsidy)} valueStyle={{ fontSize: 14 }} />
          </Col>
          <Col xs={12} sm={8} md={3}>
            <Flex vertical gap={2}>
              <Typography.Text type="secondary" style={{ fontSize: 12 }}>
                تخفیف
              </Typography.Text>
              <InputNumber
                size="small"
                min={0}
                disabled={!editing}
                value={discount}
                style={{ width: '100%' }}
                onChange={(v) => setDiscount(Number(v) || 0)}
              />
            </Flex>
          </Col>
          <Col xs={24} sm={8} md={4}>
            <Statistic
              title="صندوق"
              value={formatMoney(cashTotal)}
              valueStyle={{ fontSize: 16, fontWeight: 700, color: '#389e0d' }}
            />
          </Col>
        </Row>
      </Card>

      <style>{`
        .reception-services-scroll {
          max-height: min(42vh, 420px);
          overflow: auto;
          border: 1px solid #f0f0f0;
          border-radius: 6px;
        }
        .reception-services {
          width: 100%;
          border-collapse: collapse;
          table-layout: auto;
        }
        .reception-services thead th {
          position: sticky;
          top: 0;
          z-index: 2;
          background: #fafafa;
          border-bottom: 1px solid #f0f0f0;
          padding: 6px 8px;
          white-space: nowrap;
          font-size: 12px;
          font-weight: 600;
          text-align: right;
        }
        .reception-services td {
          border-bottom: 1px solid #f0f0f0;
          padding: 4px 6px;
          vertical-align: middle;
          white-space: nowrap;
        }
        .reception-services td.num {
          font-variant-numeric: tabular-nums;
          direction: ltr;
          text-align: left;
        }
        .reception-services .cell-muted {
          color: #bfbfbf;
          font-size: 12px;
        }
        .reception-services tr.reception-service-negative td {
          background: #fff1f0 !important;
          color: #cf1322;
        }
      `}</style>
    </Flex>
  );
}
