import { useMemo, useState } from 'react';
import {
  Alert,
  Button,
  Card,
  Form,
  InputNumber,
  Modal,
  Select,
  Space,
  Table,
  Tag,
  Typography,
} from 'antd';
import { CalculatorOutlined, EditOutlined, SaveOutlined } from '@ant-design/icons';
import { Controller, useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import type { ColumnsType } from 'antd/es/table';
import { PageHeader } from '@/platform/components/PageHeader';
import { DataTable } from '@/platform/components/DataTable';
import { PermissionGuard } from '@/platform/auth/PermissionGuard';
import { useApiQuery } from '@/platform/hooks/useApiQuery';
import { useApiMutation } from '@/platform/hooks/useApiMutation';
import { fetchOrganizations } from '@/modules/organization/api';
import { fetchServices } from '@/modules/services/api';
import type { ServiceItem } from '@/modules/services/types';
import {
  calculateTariff,
  fetchTariffsByOrganization,
  recalculateTariff,
  saveTariffs,
} from '../api';
import {
  tariffCalculateSchema,
  tariffRecalculateSchema,
  type TariffCalculateFormValues,
  type TariffRecalculateFormValues,
} from '../hooks';
import type { CalculateTariffResponse, ServiceWithPrice, Tariff } from '../types';

const emptyCalculateValues: TariffCalculateFormValues = {
  organization_id: 0,
  technical_amount: 0,
  professional_center_amount: 0,
  consumption_center_amount: 0,
};

const emptyRecalculateValues: TariffRecalculateFormValues = {
  technical_amount: 0,
  professional_center_amount: 0,
  consumption_center_amount: 0,
};

/** فرمت اعداد مبلغ برای نمایش */
function formatAmount(value: number): string {
  return new Intl.NumberFormat('fa-IR').format(value);
}

/** صفحه مدیریت تعرفه: انتخاب سازمان، تست محاسبه، ذخیره تراکنشی و ویرایش تک‌خدمت */
export function TariffListPage() {
  const [excludeServiceIds, setExcludeServiceIds] = useState<number[]>([]);
  const [calcResult, setCalcResult] = useState<CalculateTariffResponse | null>(null);
  const [editTarget, setEditTarget] = useState<Tariff | null>(null);

  const {
    control,
    handleSubmit,
    watch,
    getValues,
    formState: { errors },
  } = useForm<TariffCalculateFormValues>({
    resolver: zodResolver(tariffCalculateSchema),
    defaultValues: emptyCalculateValues,
  });

  const organizationId = watch('organization_id');

  const {
    control: editControl,
    handleSubmit: handleEditSubmit,
    reset: resetEdit,
    setError: setEditError,
    formState: { errors: editErrors },
  } = useForm<TariffRecalculateFormValues>({
    resolver: zodResolver(tariffRecalculateSchema),
    defaultValues: emptyRecalculateValues,
  });

  const { data: organizations = [] } = useApiQuery({
    queryKey: ['organizations'],
    queryFn: fetchOrganizations,
  });

  const { data: services = [], isLoading: servicesLoading } = useApiQuery({
    queryKey: ['services'],
    queryFn: fetchServices,
  });

  const {
    data: savedTariffs = [],
    isLoading: tariffsLoading,
    refetch: refetchTariffs,
  } = useApiQuery({
    queryKey: ['tariffs', organizationId],
    queryFn: () => fetchTariffsByOrganization(organizationId),
    enabled: organizationId > 0,
  });

  const hasNegativeFund = useMemo(
    () => (calcResult?.services ?? []).some((row) => row.calculate.fund_amount < 0),
    [calcResult],
  );

  const buildPayload = (values: TariffCalculateFormValues) => ({
    organization_id: values.organization_id,
    exclude_service_ids: excludeServiceIds,
    technical_amount: values.technical_amount,
    professional_center_amount: values.professional_center_amount,
    consumption_center_amount: values.consumption_center_amount,
  });

  const calculateMutation = useApiMutation({
    mutationFn: calculateTariff,
    successMessage: 'محاسبه تعرفه انجام شد',
    onSuccess: (data) => setCalcResult(data),
  });

  const saveMutation = useApiMutation({
    mutationFn: saveTariffs,
    successMessage: 'تعرفه‌ها با موفقیت ذخیره شدند',
    onSuccess: () => {
      setCalcResult(null);
      refetchTariffs();
    },
  });

  const recalculateMutation = useApiMutation({
    mutationFn: (values: TariffRecalculateFormValues) =>
      recalculateTariff(editTarget!.id, values),
    successMessage: 'تعرفه با موفقیت بازمحاسبه و ذخیره شد',
    setError: setEditError,
    onSuccess: () => {
      setEditTarget(null);
      resetEdit(emptyRecalculateValues);
      refetchTariffs();
    },
  });

  const onTest = handleSubmit((values) => {
    calculateMutation.mutate(buildPayload(values));
  });

  const onSave = handleSubmit((values) => {
    if (hasNegativeFund) return;
    saveMutation.mutate(buildPayload(values));
  });

  const openEdit = (row: Tariff) => {
    setEditTarget(row);
    const current = getValues();
    resetEdit({
      technical_amount: current.technical_amount || 0,
      professional_center_amount: current.professional_center_amount || 0,
      consumption_center_amount: current.consumption_center_amount || 0,
    });
  };

  const serviceColumns: ColumnsType<ServiceItem> = [
    { title: 'کد', dataIndex: 'service_code', key: 'service_code', width: 100 },
    { title: 'نام خدمت', dataIndex: 'name', key: 'name' },
    {
      title: 'ضریب فنی',
      dataIndex: 'technical_coefficient',
      key: 'technical_coefficient',
      width: 110,
    },
    {
      title: 'ضریب حرفه‌ای',
      dataIndex: 'professional_coefficient',
      key: 'professional_coefficient',
      width: 120,
    },
    {
      title: 'ضریب مصرفی',
      dataIndex: 'consumption_coefficient',
      key: 'consumption_coefficient',
      width: 110,
    },
  ];

  const calcColumns: ColumnsType<ServiceWithPrice> = [
    { title: 'کد', dataIndex: ['service', 'service_code'], key: 'code', width: 90 },
    { title: 'خدمت', dataIndex: ['service', 'name'], key: 'name' },
    {
      title: 'نرخ',
      key: 'total',
      render: (_, row) => formatAmount(row.calculate.total_amount),
    },
    {
      title: 'تعرفه',
      key: 'tariff',
      render: (_, row) => formatAmount(row.calculate.tariff),
    },
    {
      title: 'سهم سازمان',
      key: 'org',
      render: (_, row) => formatAmount(row.calculate.organization_share),
    },
    {
      title: 'تکمیلی',
      key: 'sup',
      render: (_, row) => formatAmount(row.calculate.supplement_amount),
    },
    {
      title: 'یارانه',
      key: 'sub',
      render: (_, row) => formatAmount(row.calculate.subsidy_amount),
    },
    {
      title: 'صندوق',
      key: 'fund',
      render: (_, row) => {
        const negative = row.calculate.fund_amount < 0;
        return (
          <Typography.Text type={negative ? 'danger' : undefined} strong={negative}>
            {formatAmount(row.calculate.fund_amount)}
          </Typography.Text>
        );
      },
    },
  ];

  const savedColumns: ColumnsType<Tariff> = [
    { title: 'کد خدمت', dataIndex: 'service_code', key: 'service_code', width: 100 },
    { title: 'نام خدمت', dataIndex: 'service_name', key: 'service_name' },
    {
      title: 'نرخ',
      dataIndex: 'amount',
      key: 'amount',
      render: (v: number) => formatAmount(v),
    },
    {
      title: 'تعرفه',
      dataIndex: 'tariff_amount',
      key: 'tariff_amount',
      render: (v: number) => formatAmount(v),
    },
    {
      title: 'سهم سازمان',
      dataIndex: 'organization_share',
      key: 'organization_share',
      render: (v: number) => formatAmount(v),
    },
    {
      title: 'تکمیلی',
      dataIndex: 'supplementary_share',
      key: 'supplementary_share',
      render: (v: number) => formatAmount(v),
    },
    {
      title: 'یارانه',
      dataIndex: 'subsidy_share',
      key: 'subsidy_share',
      render: (v: number) => formatAmount(v),
    },
    {
      title: 'صندوق',
      dataIndex: 'fund_amount',
      key: 'fund_amount',
      render: (v: number) => (
        <Typography.Text type={v < 0 ? 'danger' : undefined}>{formatAmount(v)}</Typography.Text>
      ),
    },
    {
      title: 'عملیات',
      key: 'actions',
      width: 100,
      render: (_, row) => (
        <PermissionGuard permission="tariff.update">
          <Button type="link" icon={<EditOutlined />} onClick={() => openEdit(row)}>
            ویرایش
          </Button>
        </PermissionGuard>
      ),
    },
  ];

  return (
    <>
      <PageHeader title="تعرفه سازمان" />

      <Card title="محاسبه و ذخیره تعرفه" style={{ marginBottom: 16 }}>
        <Form layout="vertical">
          <Form.Item
            label="سازمان"
            required
            validateStatus={errors.organization_id ? 'error' : undefined}
            help={errors.organization_id?.message}
          >
            <Controller
              name="organization_id"
              control={control}
              render={({ field }) => (
                <Select
                  {...field}
                  showSearch
                  optionFilterProp="label"
                  placeholder="انتخاب سازمان"
                  style={{ width: '100%', maxWidth: 420 }}
                  options={organizations.map((o) => ({
                    value: o.id,
                    label: `${o.name}${o.is_takmili ? ' (تکمیلی)' : ''}`,
                  }))}
                  onChange={(value) => {
                    field.onChange(value);
                    setCalcResult(null);
                    setExcludeServiceIds([]);
                  }}
                />
              )}
            />
          </Form.Item>

          <Space wrap size="large" style={{ width: '100%' }}>
            <Form.Item
              label="مبلغ فنی مرکز"
              required
              validateStatus={errors.technical_amount ? 'error' : undefined}
              help={errors.technical_amount?.message}
            >
              <Controller
                name="technical_amount"
                control={control}
                render={({ field }) => (
                  <InputNumber
                    {...field}
                    min={0}
                    style={{ width: 200 }}
                    onChange={(v) => field.onChange(v ?? 0)}
                  />
                )}
              />
            </Form.Item>
            <Form.Item
              label="مبلغ حرفه‌ای مرکز"
              required
              validateStatus={errors.professional_center_amount ? 'error' : undefined}
              help={errors.professional_center_amount?.message}
            >
              <Controller
                name="professional_center_amount"
                control={control}
                render={({ field }) => (
                  <InputNumber
                    {...field}
                    min={0}
                    style={{ width: 200 }}
                    onChange={(v) => field.onChange(v ?? 0)}
                  />
                )}
              />
            </Form.Item>
            <Form.Item
              label="مبلغ مصرفی مرکز"
              required
              validateStatus={errors.consumption_center_amount ? 'error' : undefined}
              help={errors.consumption_center_amount?.message}
            >
              <Controller
                name="consumption_center_amount"
                control={control}
                render={({ field }) => (
                  <InputNumber
                    {...field}
                    min={0}
                    style={{ width: 200 }}
                    onChange={(v) => field.onChange(v ?? 0)}
                  />
                )}
              />
            </Form.Item>
          </Space>

          <Typography.Paragraph type="secondary" style={{ marginBottom: 8 }}>
            خدمات انتخاب‌شده از محاسبه <Tag color="orange">حذف</Tag> می‌شوند (exclude).
          </Typography.Paragraph>

          <Table
            rowKey="id"
            size="small"
            loading={servicesLoading}
            columns={serviceColumns}
            dataSource={services}
            pagination={{ pageSize: 8, showSizeChanger: true }}
            rowSelection={{
              selectedRowKeys: excludeServiceIds,
              onChange: (keys) => setExcludeServiceIds(keys.map(Number)),
            }}
            style={{ marginBottom: 16 }}
          />

          <Space>
            <Button
              icon={<CalculatorOutlined />}
              onClick={onTest}
              loading={calculateMutation.isPending}
              disabled={!organizationId}
            >
              تست تعرفه
            </Button>
            <PermissionGuard permission="tariff.create">
              <Button
                type="primary"
                icon={<SaveOutlined />}
                onClick={onSave}
                loading={saveMutation.isPending}
                disabled={!organizationId || !calcResult || hasNegativeFund}
              >
                ذخیره
              </Button>
            </PermissionGuard>
          </Space>
        </Form>
      </Card>

      {calcResult && (
        <Card title="نتیجه تست تعرفه" style={{ marginBottom: 16 }}>
          {hasNegativeFund && (
            <Alert
              type="error"
              showIcon
              style={{ marginBottom: 12 }}
              message="صندوق یک یا چند خدمت منفی است"
              description="نتیجه تست نمایش داده می‌شود، اما تا وقتی زمانی صندوق منفی باشد امکان ذخیره وجود ندارد."
            />
          )}
          <Table
            rowKey={(row) => row.service.id}
            size="small"
            columns={calcColumns}
            dataSource={calcResult.services}
            pagination={{ pageSize: 10 }}
            rowClassName={(row) => (row.calculate.fund_amount < 0 ? 'tariff-fund-negative' : '')}
            onRow={(row) =>
              row.calculate.fund_amount < 0
                ? { style: { background: '#fff2f0' } }
                : {}
            }
          />
        </Card>
      )}

      <Card title="تعرفه‌های ذخیره‌شده سازمان">
        {!organizationId ? (
          <Typography.Text type="secondary">برای مشاهده لیست، یک سازمان انتخاب کنید.</Typography.Text>
        ) : (
          <DataTable
            columns={savedColumns}
            data={savedTariffs}
            loading={tariffsLoading}
            rowKey="id"
          />
        )}
      </Card>

      <Modal
        title={
          editTarget
            ? `بازمحاسبه تعرفه: ${editTarget.service_name}`
            : 'بازمحاسبه تعرفه'
        }
        open={!!editTarget}
        onCancel={() => {
          setEditTarget(null);
          resetEdit(emptyRecalculateValues);
        }}
        onOk={handleEditSubmit((values) => recalculateMutation.mutate(values))}
        confirmLoading={recalculateMutation.isPending}
        okText="محاسبه و ذخیره"
        cancelText="انصراف"
        destroyOnClose
      >
        <Form layout="vertical">
          <Form.Item
            label="مبلغ فنی مرکز"
            required
            validateStatus={editErrors.technical_amount ? 'error' : undefined}
            help={editErrors.technical_amount?.message}
          >
            <Controller
              name="technical_amount"
              control={editControl}
              render={({ field }) => (
                <InputNumber
                  {...field}
                  min={0}
                  style={{ width: '100%' }}
                  onChange={(v) => field.onChange(v ?? 0)}
                />
              )}
            />
          </Form.Item>
          <Form.Item
            label="مبلغ حرفه‌ای مرکز"
            required
            validateStatus={editErrors.professional_center_amount ? 'error' : undefined}
            help={editErrors.professional_center_amount?.message}
          >
            <Controller
              name="professional_center_amount"
              control={editControl}
              render={({ field }) => (
                <InputNumber
                  {...field}
                  min={0}
                  style={{ width: '100%' }}
                  onChange={(v) => field.onChange(v ?? 0)}
                />
              )}
            />
          </Form.Item>
          <Form.Item
            label="مبلغ مصرفی مرکز"
            required
            validateStatus={editErrors.consumption_center_amount ? 'error' : undefined}
            help={editErrors.consumption_center_amount?.message}
          >
            <Controller
              name="consumption_center_amount"
              control={editControl}
              render={({ field }) => (
                <InputNumber
                  {...field}
                  min={0}
                  style={{ width: '100%' }}
                  onChange={(v) => field.onChange(v ?? 0)}
                />
              )}
            />
          </Form.Item>
        </Form>
      </Modal>
    </>
  );
}
