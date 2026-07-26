import { Button, Flex, InputNumber, Space, Statistic, Typography } from 'antd';
import { SaveOutlined } from '@ant-design/icons';
import { useReceptionStore } from '../store/receptionStore';
import { lineCashAmount } from '../types';
import { useAuth } from '@/platform/auth/useAuth';

type ReceptionTotalsBarProps = {
  saving?: boolean;
  onSave: () => void;
};

function formatMoney(value: number): string {
  return Number(value ?? 0).toLocaleString('fa-IR');
}

/** نوار جمع‌ها و ذخیره چسبان پایین فضای کاری پذیرش */
export function ReceptionTotalsBar({ saving, onSave }: ReceptionTotalsBarProps) {
  const { hasPermission } = useAuth();
  const editing = useReceptionStore((s) => s.editing);
  const deleted = useReceptionStore((s) => s.deleted);
  const isNew = useReceptionStore((s) => s.isNew);
  const services = useReceptionStore((s) => s.services);
  const discount = useReceptionStore((s) => s.discount);
  const setDiscount = useReceptionStore((s) => s.setDiscount);

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

  const canSave =
    !deleted &&
    editing &&
    ((isNew && hasPermission('reception.create')) ||
      (!isNew && (hasPermission('reception.update') || hasPermission('reception.create'))));

  return (
    <Flex
      wrap="wrap"
      gap={12}
      align="center"
      justify="space-between"
      className="reception-totals-bar"
    >
      <Space size={16} wrap>
        <Statistic title="نرخ" value={formatMoney(totals.amount)} valueStyle={{ fontSize: 13 }} />
        <Statistic title="تعرفه" value={formatMoney(totals.tariff)} valueStyle={{ fontSize: 13 }} />
        <Statistic title="سازمان" value={formatMoney(totals.orgShare)} valueStyle={{ fontSize: 13 }} />
        <Statistic title="تکمیلی" value={formatMoney(totals.suppShare)} valueStyle={{ fontSize: 13 }} />
        <Statistic title="یارانه" value={formatMoney(totals.subsidy)} valueStyle={{ fontSize: 13 }} />
        <Flex vertical gap={2} style={{ minWidth: 96 }}>
          <Typography.Text type="secondary" style={{ fontSize: 12 }}>
            تخفیف
          </Typography.Text>
          <InputNumber
            size="small"
            min={0}
            disabled={!editing || deleted}
            value={discount}
            style={{ width: '100%' }}
            onChange={(v) => setDiscount(Number(v) || 0)}
          />
        </Flex>
        <Statistic
          title="صندوق"
          value={formatMoney(cashTotal)}
          valueStyle={{ fontSize: 18, fontWeight: 700, color: '#389e0d' }}
        />
      </Space>

      {canSave && (
        <Button type="primary" icon={<SaveOutlined />} loading={saving} onClick={onSave}>
          ذخیره
        </Button>
      )}
    </Flex>
  );
}
