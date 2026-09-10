import type { KeyboardEvent } from 'react';
import { Button, Empty, Flex, Typography, message } from 'antd';
import { PlusOutlined } from '@ant-design/icons';
import { useReceptionStore } from '../store/receptionStore';
import { lineCashAmount } from '../types';
import { ServiceRow } from './ServiceRow';

type ServicesTableProps = {
  onRecalculate: () => void;
};

/** جدول خدمات فشرده با هدر sticky؛ جمع‌ها در نوار پایین صفحه هستند */
export function ServicesTable({ onRecalculate }: ServicesTableProps) {
  const editing = useReceptionStore((s) => s.editing);
  const services = useReceptionStore((s) => s.services);
  const hasOrganization = useReceptionStore((s) => s.hasOrganization);
  const requestInsuranceFocus = useReceptionStore((s) => s.requestInsuranceFocus);
  const addServiceLine = useReceptionStore((s) => s.addServiceLine);
  const removeServiceLine = useReceptionStore((s) => s.removeServiceLine);
  const updateServiceLine = useReceptionStore((s) => s.updateServiceLine);

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
    if (e.code === 'F1') {
      e.preventDefault();
      e.stopPropagation();
      tryAddService();
      return;
    }
    if (e.ctrlKey && e.code === 'Space') {
      e.preventDefault();
      e.stopPropagation();
      tryAddService();
    }
  }

  return (
    <Flex vertical gap={8} className="reception-services-root" onKeyDown={handleTableKeyDown}>
      <Flex justify="space-between" align="center" wrap="wrap" gap={8}>
        <Flex align="center" gap={8}>
          <Typography.Text strong>خدمات</Typography.Text>
          {editing && (
            <Typography.Text type="secondary" style={{ fontSize: 12 }}>
              F1 یا Ctrl+Space افزودن سطر
            </Typography.Text>
          )}
        </Flex>
        {editing && (
          <Button type="primary" size="small" icon={<PlusOutlined />} onClick={tryAddService}>
            افزودن خدمت
          </Button>
        )}
      </Flex>

      <div className="reception-services-scroll">
        {services.length === 0 ? (
          <Empty
            image={Empty.PRESENTED_IMAGE_SIMPLE}
            description={
              <Flex vertical gap={4} align="center">
                <Typography.Text type="secondary">خدمتی ثبت نشده است</Typography.Text>
                {editing && (
                  <Typography.Text type="secondary" style={{ fontSize: 12 }}>
                    ابتدا بیمه را انتخاب کنید، سپس خدمت اضافه کنید (Ctrl+Space)
                  </Typography.Text>
                )}
              </Flex>
            }
          >
            {editing && (
              <Button type="primary" size="small" icon={<PlusOutlined />} onClick={tryAddService}>
                افزودن خدمت
              </Button>
            )}
          </Empty>
        ) : (
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
              {services.map((line, index) => {
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
              })}
            </tbody>
          </table>
        )}
      </div>

      <style>{`
        .reception-services-root {
          flex: 1;
          min-height: 0;
          height: 100%;
        }
        .reception-services-scroll {
          flex: 1;
          min-height: 160px;
          overflow: auto;
          border: 1px solid #f0f0f0;
          border-radius: 6px;
          display: flex;
          flex-direction: column;
        }
        .reception-services-scroll .ant-empty {
          margin: auto;
          padding: 24px 12px;
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
