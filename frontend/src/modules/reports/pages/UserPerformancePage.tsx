import { useMemo, useState } from 'react';
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
import { FileTextOutlined, WalletOutlined } from '@ant-design/icons';
import dayjs, { type Dayjs } from 'dayjs';
import { PageHeader } from '@/platform/components/PageHeader';
import { JalaliDatePicker } from '@/platform/components/JalaliDatePicker/JalaliDatePicker';
import { MultiSelectWithActions } from '@/platform/components/MultiSelectWithActions';
import { ReportViewer } from '@/platform/reports';
import type { ReportDefinition } from '@/platform/reports';
import { useApiQuery } from '@/platform/hooks/useApiQuery';
import { fetchUsers } from '@/modules/users/api';
import { userTypeOptions } from '@/modules/users/hooks';
import {
  fetchUserPerformanceByReception,
  fetchUserPerformanceByTransaction,
} from '../api';
import {
  buildUserPerformanceMetaRows,
  getUserPerformanceDefinition,
  mapUserSections,
} from '../definitions/userPerformance';
import type {
  UserPerformanceByReceptionResponse,
  UserPerformanceByTransactionResponse,
  UserPerformanceFilters,
  UserPerformanceMeta,
  UserPerformanceReceptionDetailRow,
  UserPerformanceReceptionRow,
  UserPerformanceReportType,
  UserPerformanceTransactionDetailRow,
  UserPerformanceTransactionRow,
} from '../types';

type FilterFormValues = {
  from_date?: string;
  to_date?: string;
  from_time?: Dayjs;
  to_time?: Dayjs;
  user_ids?: number[];
  separate_by_patient?: boolean;
};

type ReportState =
  | {
      type: 'by_reception';
      separateByPatient: boolean;
      data: UserPerformanceByReceptionResponse;
    }
  | {
      type: 'by_transaction';
      separateByPatient: boolean;
      data: UserPerformanceByTransactionResponse;
    }
  | null;

/** ترتیب نمایش انواع کاربر (مطابق مدل بک‌اند) */
const userTypeOrder = userTypeOptions.map((option) => option.value);

/** برچسب نوع کاربر را برمی‌گرداند */
function userTypeLabel(userType: string): string {
  return userTypeOptions.find((option) => option.value === userType)?.label ?? userType;
}

/** صفحه گزارش کارکرد کاربران */
export function UserPerformancePage() {
  const [form] = Form.useForm<FilterFormValues>();
  const [report, setReport] = useState<ReportState>(null);
  const [loading, setLoading] = useState(false);

  const { data: users = [], isLoading: usersLoading } = useApiQuery({
    queryKey: ['users'],
    queryFn: fetchUsers,
  });

  const userOptions = useMemo(
    () =>
      [...users]
        .sort((a, b) => {
          const rankA = userTypeOrder.indexOf(a.user_type);
          const rankB = userTypeOrder.indexOf(b.user_type);
          const orderA = rankA === -1 ? 99 : rankA;
          const orderB = rankB === -1 ? 99 : rankB;
          if (orderA !== orderB) return orderA - orderB;
          const nameA = `${a.family} ${a.name}`;
          const nameB = `${b.family} ${b.name}`;
          return nameA.localeCompare(nameB, 'fa');
        })
        .map((user) => ({
          value: user.id,
          label: `${user.name} ${user.family} (${userTypeLabel(user.user_type)})`,
        })),
    [users],
  );

  /** فیلترهای پایه را از فرم می‌خواند و اعتبارسنجی می‌کند */
  const readFilters = (): UserPerformanceFilters | null => {
    const values = form.getFieldsValue();
    if (!values.from_date || !values.to_date) {
      message.warning('تاریخ شروع و پایان الزامی است.');
      return null;
    }
    if (!values.user_ids?.length) {
      message.warning('حداقل یک کاربر انتخاب کنید.');
      return null;
    }
    return {
      from_date: values.from_date,
      to_date: values.to_date,
      from_time: values.from_time?.format('HH:mm') ?? '00:00',
      to_time: values.to_time?.format('HH:mm') ?? '23:59',
      user_ids: values.user_ids,
      separate_by_patient: values.separate_by_patient ?? false,
    };
  };

  /** گزارش کارکرد بر اساس پذیرش را دریافت می‌کند */
  const handleReceptionReport = async () => {
    const filters = readFilters();
    if (!filters) return;

    setLoading(true);
    try {
      const data = await fetchUserPerformanceByReception(filters);
      setReport({
        type: 'by_reception',
        separateByPatient: filters.separate_by_patient ?? false,
        data,
      });
    } catch (err) {
      message.error(err instanceof Error ? err.message : 'خطا در دریافت گزارش');
    } finally {
      setLoading(false);
    }
  };

  /** گزارش کارکرد بر اساس تراکنش مالی را دریافت می‌کند */
  const handleTransactionReport = async () => {
    const filters = readFilters();
    if (!filters) return;

    setLoading(true);
    try {
      const data = await fetchUserPerformanceByTransaction(filters);
      setReport({
        type: 'by_transaction',
        separateByPatient: filters.separate_by_patient ?? false,
        data,
      });
    } catch (err) {
      message.error(err instanceof Error ? err.message : 'خطا در دریافت گزارش');
    } finally {
      setLoading(false);
    }
  };

  const handleReset = () => {
    form.resetFields();
    setReport(null);
  };

  const reportTitleMap: Record<UserPerformanceReportType, string> = {
    by_reception: 'گزارش بر اساس پذیرش انجام‌شده',
    by_transaction: 'گزارش بر اساس مبلغ دریافتی / پرداختی',
  };

  const definition = report
    ? getUserPerformanceDefinition(report.type, report.separateByPatient)
    : null;

  const meta = report
    ? buildUserPerformanceMetaRows(report.data.meta as UserPerformanceMeta)
    : undefined;

  const sections = useMemo(() => {
    if (!report?.separateByPatient) return undefined;
    if (report.type === 'by_reception') {
      return mapUserSections(report.data.sections);
    }
    return mapUserSections(report.data.sections);
  }, [report]);

  const rowKey = useMemo((): ((row: object) => string) => {
    if (!report) return () => '0';
    if (report.separateByPatient) {
      if (report.type === 'by_reception') {
        return (row) => String((row as UserPerformanceReceptionDetailRow).reception_id);
      }
      return (row) => String((row as UserPerformanceTransactionDetailRow).transaction_id);
    }
    if (report.type === 'by_reception') {
      return (row) => String((row as UserPerformanceReceptionRow).user_id);
    }
    return (row) => String((row as UserPerformanceTransactionRow).user_id);
  }, [report]);

  const tableData = report?.separateByPatient ? [] : (report?.data.rows ?? []);

  return (
    <>
      <PageHeader
        title="گزارش کارکرد کاربران"
        subtitle="تجمیع پذیرش‌ها و تراکنش‌های مالی کاربران در بازه زمانی انتخاب‌شده"
      />

      <Form
        form={form}
        layout="vertical"
        initialValues={{
          from_time: dayjs('00:00', 'HH:mm'),
          to_time: dayjs('23:59', 'HH:mm'),
          separate_by_patient: false,
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
              name="user_ids"
              label="کاربران"
              rules={[{ required: true, message: 'حداقل یک کاربر انتخاب کنید' }]}
            >
              <MultiSelectWithActions
                placeholder="انتخاب کاربر"
                options={userOptions}
                loading={usersLoading}
              />
            </Form.Item>
          </Col>
          <Col xs={24} sm={24} md={12}>
            <Form.Item name="separate_by_patient" valuePropName="checked">
              <Checkbox>لیست به تفکیک بیماران / پرونده</Checkbox>
            </Form.Item>
          </Col>
        </Row>

        <Space wrap>
          <Button
            type="primary"
            icon={<FileTextOutlined />}
            loading={loading}
            onClick={handleReceptionReport}
          >
            گزارش بر اساس پذیرش
          </Button>
          <Button
            type="primary"
            icon={<WalletOutlined />}
            loading={loading}
            onClick={handleTransactionReport}
          >
            گزارش بر اساس دریافتی / پرداختی
          </Button>
          <Button onClick={handleReset}>پاک کردن فیلترها</Button>
        </Space>
      </Form>

      {report && definition && (
        <>
          <div style={{ marginBottom: 8, fontWeight: 600, color: '#555' }}>
            {reportTitleMap[report.type]}
            {report.separateByPatient && ' — تفکیک بیمار / پرونده'}
          </div>

          <ReportViewer
            definition={definition as ReportDefinition<object>}
            data={tableData as object[]}
            sections={sections}
            summary={report.data.summary as Record<string, unknown>}
            meta={meta}
            loading={loading}
            rowKey={rowKey}
            grandTotal={
              report.separateByPatient && (report.data.sections?.length ?? 0) > 1
                ? (report.data.summary as Record<string, unknown>)
                : undefined
            }
            grandTotalLabel="جمع کل همه کاربران"
          />
        </>
      )}
    </>
  );
}
