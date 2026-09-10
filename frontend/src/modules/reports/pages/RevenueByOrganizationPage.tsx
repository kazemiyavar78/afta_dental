import { useMemo, useState, useEffect } from 'react';
import {
  Button,
  Checkbox,
  Col,
  Form,
  Row,
  Space,
  TimePicker,
  message,
} from 'antd';
import { SearchOutlined } from '@ant-design/icons';
import dayjs, { type Dayjs } from 'dayjs';
import { PageHeader } from '@/platform/components/PageHeader';
import { JalaliDatePicker } from '@/platform/components/JalaliDatePicker/JalaliDatePicker';
import { MultiSelectWithActions } from '@/platform/components/MultiSelectWithActions';
import { ReportViewer } from '@/platform/reports';
import { useApiQuery } from '@/platform/hooks/useApiQuery';
import { fetchOrganizations } from '@/modules/organization/api';
import { fetchRevenueByOrganization } from '../api';
import {
  revenueByOrganizationDefinition,
} from '../definitions/revenueByOrganization';
import { buildRevenueOrgMetaRows } from '../definitions/metaHelpers';
import type {
  RevenueByOrganizationFilters,
  RevenueByOrganizationResponse,
} from '../types';

type FilterFormValues = {
  from_date?: string;
  to_date?: string;
  from_time?: Dayjs;
  to_time?: Dayjs;
  base_organization_ids?: number[];
  supplementary_organization_ids?: number[];
  detail_supplementary?: boolean;
  only_supplementary?: boolean;
};

/** صفحه گزارش درآمد به تفکیک سازمان */
export function RevenueByOrganizationPage() {
  const [form] = Form.useForm<FilterFormValues>();
  const [appliedFilters, setAppliedFilters] = useState<RevenueByOrganizationFilters | null>(null);
  const [reportData, setReportData] = useState<RevenueByOrganizationResponse | null>(null);

  const { data: organizations = [], isLoading: orgsLoading } = useApiQuery({
    queryKey: ['organizations'],
    queryFn: fetchOrganizations,
  });

  const supplementaryIds = Form.useWatch('supplementary_organization_ids', form);
  const onlySupplementary = Form.useWatch('only_supplementary', form);

  const freeOrganization = useMemo(
    () => organizations.find((o) => o.is_free && !o.is_takmili),
    [organizations],
  );

  useEffect(() => {
    if (onlySupplementary) {
      form.setFieldsValue({
        detail_supplementary: false,
        base_organization_ids: freeOrganization ? [freeOrganization.id] : [],
      });
    }
  }, [onlySupplementary, freeOrganization, form]);

  const baseOptions = useMemo(
    () =>
      organizations
        .filter((o) => !o.is_takmili)
        .map((o) => ({ value: o.id, label: o.name })),
    [organizations],
  );

  const supplementaryOptions = useMemo(
    () =>
      organizations
        .filter((o) => o.is_takmili)
        .map((o) => ({ value: o.id, label: o.name })),
    [organizations],
  );

  const { isLoading, isFetching } = useApiQuery({
    queryKey: ['reports', 'revenue-by-organization', appliedFilters],
    queryFn: async () => {
      if (!appliedFilters) return null;
      const data = await fetchRevenueByOrganization(appliedFilters);
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
    if (!values.only_supplementary && !values.base_organization_ids?.length) {
      message.warning('حداقل یک سازمان پایه انتخاب کنید.');
      return;
    }
    if (values.only_supplementary && !freeOrganization) {
      message.warning('سازمان آزاد تعریف نشده است.');
      return;
    }

    setAppliedFilters({
      from_date: values.from_date,
      to_date: values.to_date,
      from_time: values.from_time?.format('HH:mm'),
      to_time: values.to_time?.format('HH:mm'),
      base_organization_ids: values.only_supplementary
        ? [freeOrganization!.id]
        : values.base_organization_ids,
      supplementary_organization_ids: values.supplementary_organization_ids,
      detail_supplementary: values.detail_supplementary ?? false,
      only_supplementary: values.only_supplementary ?? false,
    });
  };

  const handleReset = () => {
    form.resetFields();
    setAppliedFilters(null);
    setReportData(null);
  };

  const meta = useMemo(
    () => (reportData?.meta ? buildRevenueOrgMetaRows(reportData.meta) : undefined),
    [reportData?.meta],
  );

  const rowKey = (row: { base_organization_id: number; supplementary_organization_id?: number | null }) =>
    `${row.base_organization_id}-${row.supplementary_organization_id ?? 'none'}`;

  return (
    <>
      <PageHeader
        title="گزارش درآمد به تفکیک سازمان"
        subtitle="تجمیع درآمد پذیرش‌ها بر اساس سازمان پایه و تکمیلی"
      />

      <Form
        form={form}
        layout="vertical"
        onFinish={handleSearch}
        initialValues={{
          from_time: dayjs('00:00', 'HH:mm'),
          to_time: dayjs('23:59', 'HH:mm'),
          detail_supplementary: false,
          only_supplementary: false,
        }}
        style={{ marginBottom: 24 }}
      >
        <Row gutter={16}>
          <Col xs={24} sm={12} md={6}>
            <Form.Item
              name="from_date"
              label="تاریخ شروع"
              rules={[{ required: true, message: 'تاریخ شروع الزامی است' }]}
            >
              <JalaliDatePicker style={{ width: '100%' }} placeholder="انتخاب تاریخ" />
            </Form.Item>
          </Col>
          <Col xs={24} sm={12} md={6}>
            <Form.Item
              name="to_date"
              label="تاریخ پایان"
              rules={[{ required: true, message: 'تاریخ پایان الزامی است' }]}
            >
              <JalaliDatePicker style={{ width: '100%' }} placeholder="انتخاب تاریخ" />
            </Form.Item>
          </Col>
          <Col xs={24} sm={12} md={6}>
            <Form.Item name="from_time" label="ساعت شروع">
              <TimePicker style={{ width: '100%' }} format="HH:mm" placeholder="00:00" />
            </Form.Item>
          </Col>
          <Col xs={24} sm={12} md={6}>
            <Form.Item name="to_time" label="ساعت پایان">
              <TimePicker style={{ width: '100%' }} format="HH:mm" placeholder="23:59" />
            </Form.Item>
          </Col>
          <Col xs={24} sm={12} md={8}>
            <Form.Item
              name="base_organization_ids"
              label="سازمان‌های پایه"
              rules={[
                {
                  validator: (_, value) => {
                    if (onlySupplementary) return Promise.resolve();
                    if (value?.length) return Promise.resolve();
                    return Promise.reject(new Error('حداقل یک سازمان پایه انتخاب کنید'));
                  },
                },
              ]}
            >
              <MultiSelectWithActions
                placeholder="انتخاب سازمان پایه"
                options={baseOptions}
                loading={orgsLoading}
                disabled={onlySupplementary}
              />
            </Form.Item>
          </Col>
          <Col xs={24} sm={12} md={8}>
            <Form.Item name="supplementary_organization_ids" label="سازمان‌های تکمیلی">
              <MultiSelectWithActions
                placeholder="انتخاب سازمان تکمیلی (اختیاری)"
                options={supplementaryOptions}
                loading={orgsLoading}
              />
            </Form.Item>
          </Col>
          <Col xs={24} sm={12} md={8}>
            <Form.Item name="detail_supplementary" valuePropName="checked">
              <Checkbox disabled={!supplementaryIds?.length || onlySupplementary}>
                ریز بیمه‌های تکمیلی
              </Checkbox>
            </Form.Item>
          </Col>
          <Col xs={24} sm={12} md={8}>
            <Form.Item
              name="only_supplementary"
              valuePropName="checked"
              extra={
                onlySupplementary
                  ? freeOrganization
                    ? `سازمان پایه: ${freeOrganization.name}`
                    : 'سازمان آزاد تعریف نشده است'
                  : undefined
              }
            >
              <Checkbox>فقط سازمان تکمیلی</Checkbox>
            </Form.Item>
          </Col>
        </Row>
        <Space>
          <Button
            type="primary"
            htmlType="submit"
            icon={<SearchOutlined />}
            loading={isLoading || isFetching}
          >
            دریافت گزارش
          </Button>
          <Button onClick={handleReset}>پاک کردن فیلترها</Button>
        </Space>
      </Form>

      {appliedFilters && (
        <ReportViewer
          definition={revenueByOrganizationDefinition}
          data={reportData?.rows ?? []}
          summary={reportData?.summary as Record<string, unknown> | undefined}
          meta={meta}
          loading={isLoading || isFetching}
          rowKey={rowKey}
          emptyText="ابتدا گزارش را دریافت کنید یا فیلترها را تغییر دهید"
        />
      )}
    </>
  );
}
