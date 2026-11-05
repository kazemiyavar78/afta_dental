import { useMemo, useState } from 'react';
import { Button, Col, Form, Row, Select, Space, TimePicker, message } from 'antd';
import { SearchOutlined } from '@ant-design/icons';
import dayjs, { type Dayjs } from 'dayjs';
import { PageHeader } from '@/platform/components/PageHeader';
import { JalaliDatePicker } from '@/platform/components/JalaliDatePicker/JalaliDatePicker';
import { ReportViewer } from '@/platform/reports';
import { useAuth } from '@/platform/auth/useAuth';
import { useApiQuery } from '@/platform/hooks/useApiQuery';
import { useApiMutation } from '@/platform/hooks/useApiMutation';
import { fetchCreditAccounts, setCreditChargePaid } from '@/modules/wallet/api';
import { fetchCreditAccountPayments } from '../api';
import {
  buildCreditAccountPaymentsDefinition,
  buildCreditAccountPaymentsMetaRows,
} from '../definitions/creditAccountPayments';
import type { CreditAccountPaymentsFilters, CreditAccountPaymentsResponse } from '../types';

type FilterFormValues = {
  from_date?: string;
  to_date?: string;
  from_time?: Dayjs;
  to_time?: Dayjs;
  credit_account_id?: number;
};

/** صفحه گزارش پرداخت‌های حساب‌های اعتباری */
export function CreditAccountPaymentsPage() {
  const [form] = Form.useForm<FilterFormValues>();
  const [appliedFilters, setAppliedFilters] = useState<CreditAccountPaymentsFilters | null>(null);
  const [reportData, setReportData] = useState<CreditAccountPaymentsResponse | null>(null);
  const [pendingId, setPendingId] = useState<number | null>(null);
  const { hasPermission } = useAuth();
  const canMarkPaid = hasPermission('credit_charge.update');

  const { data: accounts = [], isLoading: accountsLoading } = useApiQuery({
    queryKey: ['credit-accounts'],
    queryFn: fetchCreditAccounts,
  });

  const { isLoading, isFetching } = useApiQuery({
    queryKey: ['reports', 'credit-account-payments', appliedFilters],
    queryFn: async () => {
      if (!appliedFilters) return null;
      const data = await fetchCreditAccountPayments(appliedFilters);
      setReportData(data);
      return data;
    },
    enabled: appliedFilters !== null,
  });

  const handleSearch = (values: FilterFormValues) => {
    if (!values.from_date || !values.to_date) {
      message.warning('تاریخ شروع و پایان الزامی است.');
      return;
    }
    setAppliedFilters({
      from_date: values.from_date,
      to_date: values.to_date,
      from_time: values.from_time?.format('HH:mm'),
      to_time: values.to_time?.format('HH:mm'),
      credit_account_id: values.credit_account_id,
    });
  };

  const markPaidMutation = useApiMutation({
    mutationFn: ({ id, creditPaid }: { id: number; creditPaid: boolean }) => setCreditChargePaid(id, creditPaid),
    successMessage: 'وضعیت پرداخت به‌روزرسانی شد',
    onSuccess: (tx, variables) => {
      setReportData((prev) => {
        if (!prev) return prev;
        return {
          ...prev,
          rows: prev.rows.map((row) =>
            row.id === variables.id
              ? { ...row, credit_paid: tx.credit_paid ?? variables.creditPaid, credit_paid_at: tx.credit_paid_at ?? undefined }
              : row,
          ),
        };
      });
    },
    onSettled: () => setPendingId(null),
  });

  const { mutate: markPaid } = markPaidMutation;
  const definition = useMemo(
    () =>
      buildCreditAccountPaymentsDefinition({
        canMarkPaid,
        pendingId,
        onTogglePaid: (row, paid) => {
          setPendingId(row.id);
          markPaid({ id: row.id, creditPaid: paid });
        },
      }),
    [canMarkPaid, pendingId, markPaid],
  );

  const handleReset = () => {
    form.resetFields();
    setAppliedFilters(null);
    setReportData(null);
  };

  const meta = useMemo(
    () => (reportData?.meta ? buildCreditAccountPaymentsMetaRows(reportData.meta) : undefined),
    [reportData?.meta],
  );

  return (
    <>
      <PageHeader
        title="گزارش پرداخت‌های حساب‌های اعتباری"
        subtitle="شارژ پرونده مالی از محل کسر سقف اعتبار، با فیلتر تاریخ و نام حساب"
      />

      <Form
        form={form}
        layout="vertical"
        onFinish={handleSearch}
        initialValues={{
          from_time: dayjs('00:00', 'HH:mm'),
          to_time: dayjs('23:59', 'HH:mm'),
        }}
        style={{ marginBottom: 24 }}
      >
        <Row gutter={16}>
          <Col xs={24} sm={12} md={6}>
            <Form.Item
              name="from_date"
              label="از تاریخ"
              rules={[{ required: true, message: 'تاریخ شروع الزامی است' }]}
            >
              <JalaliDatePicker style={{ width: '100%' }} placeholder="انتخاب تاریخ" />
            </Form.Item>
          </Col>
          <Col xs={24} sm={12} md={6}>
            <Form.Item
              name="to_date"
              label="تا تاریخ"
              rules={[{ required: true, message: 'تاریخ پایان الزامی است' }]}
            >
              <JalaliDatePicker style={{ width: '100%' }} placeholder="انتخاب تاریخ" />
            </Form.Item>
          </Col>
          <Col xs={24} sm={12} md={4}>
            <Form.Item name="from_time" label="ساعت شروع">
              <TimePicker style={{ width: '100%' }} format="HH:mm" placeholder="00:00" />
            </Form.Item>
          </Col>
          <Col xs={24} sm={12} md={4}>
            <Form.Item name="to_time" label="ساعت پایان">
              <TimePicker style={{ width: '100%' }} format="HH:mm" placeholder="23:59" />
            </Form.Item>
          </Col>
          <Col xs={24} sm={12} md={8}>
            <Form.Item name="credit_account_id" label="نام حساب">
              <Select
                allowClear
                showSearch
                optionFilterProp="label"
                placeholder="همه حساب‌ها"
                loading={accountsLoading}
                options={accounts.map((account) => ({
                  value: account.id,
                  label: account.is_active ? account.name : `${account.name} (غیرفعال)`,
                }))}
              />
            </Form.Item>
          </Col>
        </Row>
        <Space>
          <Button type="primary" htmlType="submit" icon={<SearchOutlined />} loading={isLoading || isFetching}>
            دریافت گزارش
          </Button>
          <Button onClick={handleReset}>پاک کردن فیلترها</Button>
        </Space>
      </Form>

      {appliedFilters && (
        <ReportViewer
          definition={definition}
          data={reportData?.rows ?? []}
          summary={reportData?.summary as Record<string, unknown> | undefined}
          meta={meta}
          loading={isLoading || isFetching}
          rowKey="id"
          emptyText="پرداختی در این بازه یافت نشد"
        />
      )}
    </>
  );
}
