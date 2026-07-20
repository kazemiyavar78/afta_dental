import { useEffect, useMemo, useState } from 'react';
import {
  Button,
  Col,
  Descriptions,
  Form,
  Input,
  InputNumber,
  Modal,
  Row,
  Select,
  Space,
  Tag,
  Typography,
} from 'antd';
import { ClearOutlined, DollarOutlined, SearchOutlined, SwapOutlined } from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';
import dayjs from 'dayjs';
import { DataTable } from '@/platform/components/DataTable';
import { JalaliDatePicker } from '@/platform/components/JalaliDatePicker/JalaliDatePicker';
import { PermissionGuard } from '@/platform/auth/PermissionGuard';
import { useApiQuery } from '@/platform/hooks/useApiQuery';
import { fetchWalletBalance, fetchWalletTransactions } from '@/modules/wallet/api';
import type { WalletTransaction } from '@/modules/wallet/types';
import type { Patient } from '@/modules/patients/types';
import { CashAdjustModal } from './CashAdjustModal';
import { CardToCardAdjustModal } from './CardToCardAdjustModal';

type LedgerFilters = {
  dateFrom: string;
  dateTo: string;
  action: string;
  receptionId: string;
  description: string;
};

const emptyFilters: LedgerFilters = {
  dateFrom: '',
  dateTo: '',
  action: '',
  receptionId: '',
  description: '',
};

const ACTION_OPTIONS = [
  { value: '', label: 'همه انواع' },
  { value: 'service_added', label: 'بدهی خدمات (صندوق)' },
  { value: 'service_removed', label: 'کاهش بدهی خدمات' },
  { value: 'charge', label: 'شارژ کیف پول' },
  { value: 'payment', label: 'پرداخت از کیف پول' },
  { value: 'refund', label: 'برگشت به کیف پول' },
];

/** برچسب فارسی نوع عملیات تراکنش را برمی‌گرداند. */
function actionLabel(action: string): string {
  return ACTION_OPTIONS.find((o) => o.value === action)?.label ?? action;
}

/** برچسب فارسی دسته‌بندی تراکنش را برمی‌گرداند. */
function categoryLabel(category: string): string {
  if (category === 'cash_fund') return 'صندوق';
  if (category === 'wallet') return 'کیف پول';
  return category;
}

/** برچسب فارسی روش پرداخت را برمی‌گرداند. */
function paymentMethodLabel(method?: string | null): string {
  switch (method) {
    case 'cash':
      return 'نقدی';
    case 'card_reader':
      return 'کارتخوان';
    case 'card_to_card':
      return 'کارت‌به‌کارت';
    case 'check':
      return 'چک';
    case 'wallet_credit':
      return 'اعتبار کیف پول';
    default:
      return '—';
  }
}

/** مشخص می‌کند تراکنش افزایش موجودی است یا کاهش (بدهی خدمات صندوق منفی است). */
function isCreditAction(action: string): boolean {
  return action === 'service_removed' || action === 'charge' || action === 'refund';
}

/** رنگ تگ نوع عملیات را برمی‌گرداند. */
function actionTagColor(action: string): string {
  if (action === 'charge' || action === 'service_removed') return 'green';
  if (action === 'refund') return 'blue';
  if (action === 'service_added' || action === 'payment') return 'red';
  return 'default';
}

/** مبلغ را با فرمت فارسی نمایش می‌دهد. */
function formatAmount(value: number): string {
  return `${value.toLocaleString('fa-IR')} ریال`;
}

/** تاریخ میلادی را به شمسی خوانا تبدیل می‌کند. */
function formatDateTime(value?: string | null): string {
  if (!value) return '—';
  const d = dayjs(value);
  if (!d.isValid()) return '—';
  return d.calendar('jalali').locale('fa').format('YYYY/MM/DD HH:mm');
}

type PatientWalletLedgerModalProps = {
  open: boolean;
  patient: Patient | null;
  onClose: () => void;
  /** نمایش دکمه‌های پرداخت نقدی و کارت‌به‌کارت داخل مدال */
  showPaymentActions?: boolean;
};

/** پاپ‌آپ پرونده مالی بیمار: اطلاعات، موجودی، فیلتر و جدول تراکنش‌ها */
export function PatientWalletLedgerModal({
  open,
  patient,
  onClose,
  showPaymentActions = false,
}: PatientWalletLedgerModalProps) {
  const [filters, setFilters] = useState<LedgerFilters>(emptyFilters);
  const [draft, setDraft] = useState<LedgerFilters>(emptyFilters);
  const [cashOpen, setCashOpen] = useState(false);
  const [cardOpen, setCardOpen] = useState(false);

  const fileId = patient?.id;

  const { data: balance, isLoading: balanceLoading, refetch: refetchBalance } = useApiQuery({
    queryKey: ['wallet-balance', fileId],
    queryFn: () => fetchWalletBalance(fileId!),
    enabled: open && !!fileId,
  });

  const {
    data: transactions = [],
    isLoading: txLoading,
    refetch: refetchTx,
  } = useApiQuery({
    queryKey: ['wallet-transactions', fileId],
    queryFn: () => fetchWalletTransactions(fileId!),
    enabled: open && !!fileId,
  });

  useEffect(() => {
    if (!open) return;
    setFilters(emptyFilters);
    setDraft(emptyFilters);
    if (fileId) {
      refetchBalance();
      refetchTx();
    }
  }, [open, fileId, refetchBalance, refetchTx]);

  const filtered = useMemo(() => {
    return transactions.filter((tx) => {
      const created = dayjs(tx.created_at);
      if (filters.dateFrom) {
        const from = dayjs(filters.dateFrom).startOf('day');
        if (created.isBefore(from)) return false;
      }
      if (filters.dateTo) {
        const to = dayjs(filters.dateTo).endOf('day');
        if (created.isAfter(to)) return false;
      }
      if (filters.action && tx.action !== filters.action) return false;
      if (filters.receptionId.trim()) {
        const rid = String(tx.reception_id ?? '');
        if (!rid.includes(filters.receptionId.trim())) return false;
      }
      if (filters.description.trim()) {
        const q = filters.description.trim().toLowerCase();
        if (!(tx.description ?? '').toLowerCase().includes(q)) return false;
      }
      return true;
    });
  }, [transactions, filters]);

  const columns: ColumnsType<WalletTransaction> = [
    {
      title: 'تاریخ',
      dataIndex: 'created_at',
      key: 'created_at',
      width: 140,
      render: (v: string) => formatDateTime(v),
    },
    {
      title: 'نوع',
      dataIndex: 'action',
      key: 'action',
      width: 170,
      render: (action: string) => <Tag color={actionTagColor(action)}>{actionLabel(action)}</Tag>,
    },
    {
      title: 'دسته',
      dataIndex: 'category',
      key: 'category',
      width: 90,
      render: (category: string) => categoryLabel(category),
    },
    {
      title: 'روش پرداخت',
      dataIndex: 'payment_method',
      key: 'payment_method',
      width: 110,
      render: (v: string | null | undefined) => paymentMethodLabel(v),
    },
    {
      title: 'پذیرش',
      dataIndex: 'reception_id',
      key: 'reception_id',
      width: 90,
      render: (v: number | null | undefined) => (v ? String(v) : '—'),
    },
    {
      title: 'توسط',
      dataIndex: 'performed_by_name',
      key: 'performed_by',
      width: 120,
      render: (name: string, record) => (record.performed_by === 0 ? 'سیستم' : name || '—'),
    },
    {
      title: 'پرداختی',
      dataIndex: 'amount',
      key: 'paid',
      width: 140,
      align: 'left',
      render: (amount: number, record) => {
        if (!isCreditAction(record.action)) return '—';
        return (
          <Typography.Text type="success" strong>
            {formatAmount(amount)}
          </Typography.Text>
        );
      },
    },
    {
      title: 'دریافتی',
      dataIndex: 'amount',
      key: 'received',
      width: 140,
      align: 'left',
      render: (amount: number, record) => {
        if (isCreditAction(record.action)) return '—';
        return (
          <Typography.Text type="danger" strong>
            {formatAmount(amount)}
          </Typography.Text>
        );
      },
    },
    {
      title: 'پیگیری / کارت',
      key: 'refs',
      width: 150,
      render: (_, record) => {
        const parts = [record.tracking_number, record.counterparty_card, record.bank_name].filter(Boolean);
        return parts.length ? parts.join(' | ') : ' ';
      },
    },
    {
      title: 'توضیحات',
      dataIndex: 'description',
      key: 'description',
      ellipsis: true,
      render: (v: string) => (
        <Typography.Paragraph style={{ marginBottom: 0, whiteSpace: 'pre-wrap' }} ellipsis={{ rows: 2, tooltip: v }}>
          {v || '—'}
        </Typography.Paragraph>
      ),
    },
  ];

  if (!patient) return null;

  const balanceValue = balance?.balance ?? 0;

  return (
    <Modal
      title={`پرونده مالی — ${patient.first_name} ${patient.last_name}`}
      open={open}
      onCancel={onClose}
      footer={
        <Button onClick={onClose}>بستن</Button>
      }
      destroyOnHidden
      width={1100}
      styles={{ body: { maxHeight: '75vh', overflowY: 'auto' } }}
    >
      <Descriptions
        bordered
        size="small"
        column={{ xs: 1, sm: 2, md: 3 }}
        style={{ marginBottom: 16 }}
      >
        <Descriptions.Item label="شماره پرونده">{patient.file_number}</Descriptions.Item>
        <Descriptions.Item label="کد ملی">{patient.national_code}</Descriptions.Item>
        <Descriptions.Item label="موبایل">{patient.mobile_phone_number || '—'}</Descriptions.Item>
        <Descriptions.Item label="نام">{patient.first_name}</Descriptions.Item>
        <Descriptions.Item label="نام خانوادگی">{patient.last_name}</Descriptions.Item>
        <Descriptions.Item label="موجودی کیف پول">
          <Typography.Text
            strong
            type={balanceValue > 0 ? 'success' : balanceValue < 0 ? 'danger' : undefined}
          >
            {balanceLoading ? '...' : formatAmount(balanceValue)}
          </Typography.Text>
        </Descriptions.Item>
      </Descriptions>

      {showPaymentActions && (
        <Space style={{ marginBottom: 16 }} wrap>
          <PermissionGuard permission="wallet.cash">
            <Button icon={<DollarOutlined />} onClick={() => setCashOpen(true)}>
              نقدی
            </Button>
          </PermissionGuard>
          <PermissionGuard permission="wallet.card_to_card">
            <Button icon={<SwapOutlined />} onClick={() => setCardOpen(true)}>
              کارت‌به‌کارت
            </Button>
          </PermissionGuard>
        </Space>
      )}

      <Form layout="vertical" style={{ marginBottom: 12 }}>
        <Row gutter={12}>
          <Col xs={24} sm={12} md={6}>
            <Form.Item label="از تاریخ">
              <JalaliDatePicker
                style={{ width: '100%' }}
                value={draft.dateFrom || undefined}
                onChange={(v) => setDraft((prev) => ({ ...prev, dateFrom: v || '' }))}
                allowClear
              />
            </Form.Item>
          </Col>
          <Col xs={24} sm={12} md={6}>
            <Form.Item label="تا تاریخ">
              <JalaliDatePicker
                style={{ width: '100%' }}
                value={draft.dateTo || undefined}
                onChange={(v) => setDraft((prev) => ({ ...prev, dateTo: v || '' }))}
                allowClear
              />
            </Form.Item>
          </Col>
          <Col xs={24} sm={12} md={6}>
            <Form.Item label="نوع تراکنش">
              <Select
                style={{ width: '100%' }}
                value={draft.action}
                onChange={(v) => setDraft((prev) => ({ ...prev, action: v }))}
                options={ACTION_OPTIONS}
              />
            </Form.Item>
          </Col>
          <Col xs={24} sm={12} md={6}>
            <Form.Item label="شماره پذیرش">
              <InputNumber
                style={{ width: '100%' }}
                min={1}
                value={draft.receptionId ? Number(draft.receptionId) : undefined}
                onChange={(v) => setDraft((prev) => ({ ...prev, receptionId: v ? String(v) : '' }))}
              />
            </Form.Item>
          </Col>
          <Col xs={24} sm={24} md={12}>
            <Form.Item label="توضیحات">
              <Input
                allowClear
                value={draft.description}
                onChange={(e) => setDraft((prev) => ({ ...prev, description: e.target.value }))}
                placeholder="جستجو در توضیحات..."
              />
            </Form.Item>
          </Col>
          <Col xs={24} sm={24} md={12}>
            <Form.Item label=" ">
              <Space>
                <Button type="primary" icon={<SearchOutlined />} onClick={() => setFilters(draft)}>
                  اعمال فیلتر
                </Button>
                <Button
                  icon={<ClearOutlined />}
                  onClick={() => {
                    setDraft(emptyFilters);
                    setFilters(emptyFilters);
                  }}
                >
                  پاک کردن
                </Button>
              </Space>
            </Form.Item>
          </Col>
        </Row>
      </Form>

      <Typography.Text type="secondary" style={{ display: 'block', marginBottom: 8 }}>
        {filtered.length.toLocaleString('fa-IR')} تراکنش از {transactions.length.toLocaleString('fa-IR')} مورد
      </Typography.Text>

      <DataTable
        columns={columns}
        data={filtered}
        loading={txLoading}
        rowKey="id"
        pageSize={10}
      />

      {showPaymentActions && (
        <>
          <CashAdjustModal
            open={cashOpen}
            patient={patient}
            onClose={() => {
              setCashOpen(false);
              refetchBalance();
              refetchTx();
            }}
          />
          <CardToCardAdjustModal
            open={cardOpen}
            patient={patient}
            onClose={() => {
              setCardOpen(false);
              refetchBalance();
              refetchTx();
            }}
          />
        </>
      )}
    </Modal>
  );
}
