import { useEffect, useMemo, useState } from 'react';
import {
  Button,
  Checkbox,
  Col,
  Form,
  Modal,
  Row,
  Space,
  TimePicker,
  message,
} from 'antd';
import {
  TeamOutlined,
  AppstoreOutlined,
  UnorderedListOutlined,
  BarChartOutlined,
} from '@ant-design/icons';
import dayjs, { type Dayjs } from 'dayjs';
import { PageHeader } from '@/platform/components/PageHeader';
import { JalaliDatePicker } from '@/platform/components/JalaliDatePicker/JalaliDatePicker';
import { MultiSelectWithActions } from '@/platform/components/MultiSelectWithActions';
import { ReportViewer } from '@/platform/reports';
import type { ReportDefinition } from '@/platform/reports';
import { useApiQuery } from '@/platform/hooks/useApiQuery';
import { fetchOrganizations } from '@/modules/organization/api';
import { fetchDoctors } from '@/modules/reception/api';
import { fetchServices } from '@/modules/services/api';
import {
  fetchDoctorPerformanceByOrganization,
  fetchDoctorPerformanceByService,
  fetchDoctorPerformancePatientList,
  fetchRevenueByDoctors,
} from '../api';
import {
  buildDoctorPerformanceMetaRows,
  getDoctorPerformanceDefinition,
  mapDoctorSections,
} from '../definitions/doctorPerformance';
import type {
  DoctorPerformanceBaseFilters,
  DoctorPerformanceByOrganizationFilters,
  DoctorPerformanceByOrganizationRow,
  DoctorPerformanceByServiceFilters,
  DoctorPerformanceByServiceRow,
  DoctorPerformanceDoctorSection,
  DoctorPerformanceMeta,
  DoctorPerformancePatientRow,
  DoctorPerformanceReportType,
  DoctorPerformanceSummaryRow,
  RevenueByDoctorsRow,
} from '../types';

type FilterFormValues = {
  from_date?: string;
  to_date?: string;
  from_time?: Dayjs;
  to_time?: Dayjs;
  doctor_ids?: number[];
  separate_doctors?: boolean;
};

type OrgModalValues = {
  base_organization_ids?: number[];
  supplementary_organization_ids?: number[];
  detail_supplementary?: boolean;
  only_supplementary?: boolean;
};

type ServiceModalValues = {
  service_ids?: number[];
};

type ReportStateBase = {
  separateDoctors: boolean;
  meta: DoctorPerformanceMeta;
};

type ReportState =
  | (ReportStateBase & {
      type: 'by_organization';
      rows: DoctorPerformanceByOrganizationRow[];
      summary: DoctorPerformanceSummaryRow;
      sections?: DoctorPerformanceDoctorSection<DoctorPerformanceByOrganizationRow>[];
    })
  | (ReportStateBase & {
      type: 'by_service';
      rows: DoctorPerformanceByServiceRow[];
      summary: DoctorPerformanceSummaryRow;
      sections?: DoctorPerformanceDoctorSection<DoctorPerformanceByServiceRow>[];
    })
  | (ReportStateBase & {
      type: 'patient_list';
      rows: DoctorPerformancePatientRow[];
      summary: DoctorPerformanceSummaryRow;
      sections?: DoctorPerformanceDoctorSection<DoctorPerformancePatientRow>[];
    })
  | (ReportStateBase & {
      type: 'revenue_by_doctors';
      rows: RevenueByDoctorsRow[];
      summary: DoctorPerformanceSummaryRow;
    })
  | null;

/** آیا گزارش باید به‌صورت جداگانه برای هر پزشک نمایش داده شود */
function shouldSeparateDoctors(separateDoctors: boolean, doctorIds: number[]): boolean {
  return separateDoctors && doctorIds.length > 1;
}

/** صفحه گزارش کارکرد پزشکان */
export function DoctorPerformancePage() {
  const [form] = Form.useForm<FilterFormValues>();
  const [orgForm] = Form.useForm<OrgModalValues>();
  const [serviceForm] = Form.useForm<ServiceModalValues>();

  const [report, setReport] = useState<ReportState>(null);
  const [loading, setLoading] = useState(false);
  const [orgModalOpen, setOrgModalOpen] = useState(false);
  const [serviceModalOpen, setServiceModalOpen] = useState(false);

  const doctorIds = Form.useWatch('doctor_ids', form);
  const separateDoctors = Form.useWatch('separate_doctors', form);

  const { data: organizations = [], isLoading: orgsLoading } = useApiQuery({
    queryKey: ['organizations'],
    queryFn: fetchOrganizations,
  });

  const { data: doctors = [], isLoading: doctorsLoading } = useApiQuery({
    queryKey: ['doctors'],
    queryFn: fetchDoctors,
  });

  const { data: services = [], isLoading: servicesLoading } = useApiQuery({
    queryKey: ['services'],
    queryFn: fetchServices,
  });

  const supplementaryIds = Form.useWatch('supplementary_organization_ids', orgForm);
  const onlySupplementary = Form.useWatch('only_supplementary', orgForm);

  const freeOrganization = useMemo(
    () => organizations.find((o) => o.is_free && !o.is_takmili),
    [organizations],
  );

  useEffect(() => {
    if (onlySupplementary) {
      orgForm.setFieldsValue({
        detail_supplementary: false,
        base_organization_ids: freeOrganization ? [freeOrganization.id] : [],
      });
    }
  }, [onlySupplementary, freeOrganization, orgForm]);

  useEffect(() => {
    if ((doctorIds?.length ?? 0) <= 1 && separateDoctors) {
      form.setFieldValue('separate_doctors', false);
    }
  }, [doctorIds, separateDoctors, form]);

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

  const doctorOptions = useMemo(
    () =>
      doctors
        .filter((d) => d.is_active)
        .map((d) => ({
          value: d.id,
          label: `${d.name} ${d.family}${d.medical_code ? ` (${d.medical_code})` : ''}`,
        })),
    [doctors],
  );

  const serviceOptions = useMemo(
    () =>
      services
        .filter((s) => s.is_active)
        .map((s) => ({ value: s.id, label: `${s.service_code} — ${s.name}` })),
    [services],
  );

  /** فیلترهای پایه را از فرم اصلی می‌خواند و اعتبارسنجی می‌کند */
  const readBaseFilters = (): (DoctorPerformanceBaseFilters & { separate_doctors: boolean }) | null => {
    const values = form.getFieldsValue();
    if (!values.from_date || !values.to_date) {
      message.warning('تاریخ شروع و پایان الزامی است.');
      return null;
    }
    if (!values.doctor_ids?.length) {
      message.warning('حداقل یک پزشک انتخاب کنید.');
      return null;
    }
    return {
      from_date: values.from_date,
      to_date: values.to_date,
      from_time: values.from_time?.format('HH:mm') ?? '00:00',
      to_time: values.to_time?.format('HH:mm') ?? '23:59',
      doctor_ids: values.doctor_ids,
      separate_doctors: values.separate_doctors ?? false,
    };
  };

  /** گزارش به تفکیک سازمان را دریافت می‌کند */
  const handleOrgReport = async (orgValues: OrgModalValues) => {
    const base = readBaseFilters();
    if (!base) return;

    if (!orgValues.only_supplementary && !orgValues.base_organization_ids?.length) {
      message.warning('حداقل یک سازمان پایه انتخاب کنید.');
      return;
    }
    if (orgValues.only_supplementary && !freeOrganization) {
      message.warning('سازمان آزاد تعریف نشده است.');
      return;
    }

    const filters: DoctorPerformanceByOrganizationFilters = {
      ...base,
      base_organization_ids: orgValues.only_supplementary
        ? [freeOrganization!.id]
        : orgValues.base_organization_ids,
      supplementary_organization_ids: orgValues.supplementary_organization_ids,
      detail_supplementary: orgValues.detail_supplementary ?? false,
      only_supplementary: orgValues.only_supplementary ?? false,
    };

    setLoading(true);
    try {
      const data = await fetchDoctorPerformanceByOrganization(filters);
      setReport({
        type: 'by_organization',
        separateDoctors: shouldSeparateDoctors(base.separate_doctors, base.doctor_ids!),
        sections: data.sections,
        rows: data.rows,
        summary: data.summary,
        meta: data.meta,
      });
      setOrgModalOpen(false);
    } catch (err) {
      message.error(err instanceof Error ? err.message : 'خطا در دریافت گزارش');
    } finally {
      setLoading(false);
    }
  };

  /** گزارش به تفکیک خدمت را دریافت می‌کند */
  const handleServiceReport = async (serviceValues: ServiceModalValues) => {
    const base = readBaseFilters();
    if (!base) return;

    if (!serviceValues.service_ids?.length) {
      message.warning('حداقل یک خدمت انتخاب کنید.');
      return;
    }

    const filters: DoctorPerformanceByServiceFilters = {
      ...base,
      service_ids: serviceValues.service_ids,
    };

    setLoading(true);
    try {
      const data = await fetchDoctorPerformanceByService(filters);
      setReport({
        type: 'by_service',
        separateDoctors: shouldSeparateDoctors(base.separate_doctors, base.doctor_ids!),
        sections: data.sections,
        rows: data.rows,
        summary: data.summary,
        meta: data.meta,
      });
      setServiceModalOpen(false);
    } catch (err) {
      message.error(err instanceof Error ? err.message : 'خطا در دریافت گزارش');
    } finally {
      setLoading(false);
    }
  };

  /** لیست بیماران را دریافت می‌کند */
  const handlePatientListReport = async () => {
    const base = readBaseFilters();
    if (!base) return;

    setLoading(true);
    try {
      const data = await fetchDoctorPerformancePatientList(base);
      setReport({
        type: 'patient_list',
        separateDoctors: shouldSeparateDoctors(base.separate_doctors, base.doctor_ids!),
        sections: data.sections,
        rows: data.rows,
        summary: data.summary,
        meta: data.meta,
      });
    } catch (err) {
      message.error(err instanceof Error ? err.message : 'خطا در دریافت گزارش');
    } finally {
      setLoading(false);
    }
  };

  /** گزارش درآمد به تفکیک پزشکان را دریافت می‌کند */
  const handleRevenueByDoctorsReport = async () => {
    const base = readBaseFilters();
    if (!base) return;

    setLoading(true);
    try {
      const data = await fetchRevenueByDoctors(base);
      setReport({
        type: 'revenue_by_doctors',
        separateDoctors: false,
        rows: data.rows,
        summary: data.summary,
        meta: data.meta,
      });
    } catch (err) {
      message.error(err instanceof Error ? err.message : 'خطا در دریافت گزارش');
    } finally {
      setLoading(false);
    }
  };

  const handleReset = () => {
    form.resetFields();
    orgForm.resetFields();
    serviceForm.resetFields();
    setReport(null);
  };

  const reportTitleMap: Record<DoctorPerformanceReportType, string> = {
    by_organization: 'گزارش به تفکیک سازمان',
    by_service: 'گزارش به تفکیک خدمت',
    patient_list: 'لیست بیماران',
    revenue_by_doctors: 'گزارش درآمد به تفکیک پزشکان',
  };

  const definition = report ? getDoctorPerformanceDefinition(report.type) : null;
  const meta = report ? buildDoctorPerformanceMetaRows(report.meta) : undefined;
  const sections = useMemo(() => {
    if (!report?.separateDoctors || report.type === 'revenue_by_doctors') {
      return undefined;
    }
    return mapDoctorSections(
      report.sections as Array<{
        doctor_id: number;
        doctor_name: string;
        rows: object[];
        summary: object;
      }>,
    );
  }, [report]);

  const hasMultiDoctorSections =
    report != null &&
    report.type !== 'revenue_by_doctors' &&
    report.separateDoctors &&
    (report.sections?.length ?? 0) > 1;

  const rowKey = useMemo((): ((row: object) => string) => {
    if (!report) return () => '0';
    if (report.type === 'by_organization') {
      return (row) => {
        const r = row as DoctorPerformanceByOrganizationRow;
        return `${r.base_organization_id}-${r.supplementary_organization_id ?? 'none'}`;
      };
    }
    if (report.type === 'by_service') {
      return (row) => String((row as DoctorPerformanceByServiceRow).service_id);
    }
    if (report.type === 'revenue_by_doctors') {
      return (row) => String((row as RevenueByDoctorsRow).doctor_id);
    }
    return (row) => String((row as DoctorPerformancePatientRow).patient_id);
  }, [report]);

  return (
    <>
      <PageHeader
        title="گزارش کارکرد پزشکان"
        subtitle="تجمیع پذیرش‌های پزشکان در بازه زمانی انتخاب‌شده"
      />

      <Form
        form={form}
        layout="vertical"
        initialValues={{
          from_time: dayjs('00:00', 'HH:mm'),
          to_time: dayjs('23:59', 'HH:mm'),
          separate_doctors: false,
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
          <Col xs={24} sm={24} md={12}>
            <Form.Item
              name="doctor_ids"
              label="پزشکان"
              rules={[{ required: true, message: 'حداقل یک پزشک انتخاب کنید' }]}
            >
              <MultiSelectWithActions
                placeholder="انتخاب پزشک"
                options={doctorOptions}
                loading={doctorsLoading}
              />
            </Form.Item>
          </Col>
          <Col xs={24} sm={24} md={12}>
            <Form.Item name="separate_doctors" valuePropName="checked">
              <Checkbox disabled={(doctorIds?.length ?? 0) <= 1}>
                پزشکان به صورت جدا
              </Checkbox>
            </Form.Item>
          </Col>
        </Row>

        <Space wrap>
          <Button
            type="primary"
            icon={<TeamOutlined />}
            loading={loading}
            onClick={() => setOrgModalOpen(true)}
          >
            گزارش به تفکیک سازمان
          </Button>
          <Button
            type="primary"
            icon={<AppstoreOutlined />}
            loading={loading}
            onClick={() => setServiceModalOpen(true)}
          >
            گزارش به تفکیک خدمت
          </Button>
          <Button
            type="primary"
            icon={<UnorderedListOutlined />}
            loading={loading}
            onClick={handlePatientListReport}
          >
            لیست بیماران
          </Button>
          <Button
            type="primary"
            icon={<BarChartOutlined />}
            loading={loading}
            onClick={handleRevenueByDoctorsReport}
          >
            درآمد به تفکیک پزشکان
          </Button>
          <Button onClick={handleReset}>پاک کردن فیلترها</Button>
        </Space>
      </Form>

      {report && definition && (
        <>
          <div style={{ marginBottom: 8, fontWeight: 600, color: '#555' }}>
            {reportTitleMap[report.type]}
            {report.separateDoctors && ' — تفکیک پزشکان'}
          </div>

          <ReportViewer
            definition={definition as ReportDefinition<object>}
            data={(report.separateDoctors ? [] : report.rows) as object[]}
            sections={sections}
            summary={report.summary as Record<string, unknown>}
            meta={meta}
            loading={loading}
            rowKey={rowKey}
            grandTotal={
              hasMultiDoctorSections
                ? (report.summary as Record<string, unknown>)
                : undefined
            }
            grandTotalLabel="جمع کل همه پزشکان"
          />
        </>
      )}

      <Modal
        title="انتخاب سازمان برای گزارش"
        open={orgModalOpen}
        onCancel={() => setOrgModalOpen(false)}
        footer={null}
        destroyOnClose
        width={720}
      >
        <Form
          form={orgForm}
          layout="vertical"
          onFinish={handleOrgReport}
          initialValues={{
            detail_supplementary: false,
            only_supplementary: false,
          }}
        >
          <Row gutter={16}>
            <Col xs={24} sm={12}>
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
            <Col xs={24} sm={12}>
              <Form.Item name="supplementary_organization_ids" label="سازمان‌های تکمیلی">
                <MultiSelectWithActions
                  placeholder="انتخاب سازمان تکمیلی (اختیاری)"
                  options={supplementaryOptions}
                  loading={orgsLoading}
                />
              </Form.Item>
            </Col>
            <Col xs={24} sm={12}>
              <Form.Item name="detail_supplementary" valuePropName="checked">
                <Checkbox disabled={!supplementaryIds?.length || onlySupplementary}>
                  ریز بیمه‌های تکمیلی
                </Checkbox>
              </Form.Item>
            </Col>
            <Col xs={24} sm={12}>
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
            <Button type="primary" htmlType="submit" loading={loading}>
              دریافت گزارش
            </Button>
            <Button onClick={() => setOrgModalOpen(false)}>انصراف</Button>
          </Space>
        </Form>
      </Modal>

      <Modal
        title="انتخاب خدمت برای گزارش"
        open={serviceModalOpen}
        onCancel={() => setServiceModalOpen(false)}
        footer={null}
        destroyOnClose
        width={560}
      >
        <Form form={serviceForm} layout="vertical" onFinish={handleServiceReport}>
          <Form.Item
            name="service_ids"
            label="خدمات"
            rules={[{ required: true, message: 'حداقل یک خدمت انتخاب کنید' }]}
          >
            <MultiSelectWithActions
              placeholder="انتخاب خدمت"
              options={serviceOptions}
              loading={servicesLoading}
            />
          </Form.Item>
          <Space>
            <Button type="primary" htmlType="submit" loading={loading}>
              دریافت گزارش
            </Button>
            <Button onClick={() => setServiceModalOpen(false)}>انصراف</Button>
          </Space>
        </Form>
      </Modal>
    </>
  );
}
