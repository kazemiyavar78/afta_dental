import { useState } from 'react';
import { Button, Col, DatePicker, Form, Input, Row, Space } from 'antd';
import { ReloadOutlined, SearchOutlined } from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';
import dayjs from 'dayjs';
import { PageHeader } from '@/platform/components/PageHeader';
import { DataTable } from '@/platform/components/DataTable';
import { useApiQuery } from '@/platform/hooks/useApiQuery';
import { useAuth } from '@/platform/auth/useAuth';
import { fetchLogs, fetchMyLogs } from '../api';
import type { LogFilters, SecurityLog } from '../types';

type FilterFormValues = {
  ip?: string;
  username?: string;
  from_date?: dayjs.Dayjs;
  to_date?: dayjs.Dayjs;
};

/** قالب تاریخ برای نمایش زمان رویداد */
function formatDate(value: string | null | undefined): string {
  if (!value) return '—';
  return dayjs(value).calendar('jalali').locale('fa').format('YYYY/MM/DD HH:mm');
}

/** صفحه مشاهده لاگ‌های امنیتی — ادمین با فیلتر، سایر کاربران فقط لاگ خود */
export function LogsPage() {
  const { isAdmin, hasPermission } = useAuth();
  const canReadAllLogs = isAdmin || hasPermission('logs.read');
  const [form] = Form.useForm<FilterFormValues>();
  const [appliedFilters, setAppliedFilters] = useState<LogFilters>({});

  const { data = [], isLoading, refetch, isFetching } = useApiQuery({
    queryKey: ['logs', canReadAllLogs ? 'all' : 'mine', appliedFilters],
    queryFn: () => (canReadAllLogs ? fetchLogs(appliedFilters) : fetchMyLogs()),
  });

  const columns: ColumnsType<SecurityLog> = [
    { title: 'شناسه', dataIndex: 'id', key: 'id', width: 90 },
    ...(canReadAllLogs
      ? [{ title: 'شناسه کاربر', dataIndex: 'user_id', key: 'user_id', width: 110 }]
      : []),
    { title: 'آدرس IP', dataIndex: 'ip', key: 'ip', width: 140 },
    { title: 'نوع رویداد', dataIndex: 'event_type', key: 'event_type', width: 160 },
    {
      title: 'توضیحات',
      dataIndex: 'description',
      key: 'description',
      ellipsis: true,
    },
    {
      title: 'زمان',
      dataIndex: 'created_at',
      key: 'created_at',
      width: 170,
      render: (value: string) => formatDate(value),
    },
  ];

  const handleSearch = (values: FilterFormValues) => {
    setAppliedFilters({
      ip: values.ip?.trim() || undefined,
      username: values.username?.trim() || undefined,
      from_date: values.from_date?.format('YYYY-MM-DD'),
      to_date: values.to_date?.format('YYYY-MM-DD'),
    });
  };

  const handleReset = () => {
    form.resetFields();
    setAppliedFilters({});
  };

  return (
    <>
      <PageHeader
        title="لاگ‌های امنیتی"
        subtitle={
          canReadAllLogs
            ? 'مشاهده و جستجوی همه رویدادهای امنیتی سیستم'
            : 'مشاهده رویدادهای امنیتی مربوط به حساب شما'
        }
        extra={
          <Button icon={<ReloadOutlined />} onClick={() => refetch()} loading={isFetching}>
            بروزرسانی
          </Button>
        }
      />

      {canReadAllLogs && (
        <Form
          form={form}
          layout="vertical"
          onFinish={handleSearch}
          style={{ marginBottom: 24 }}
        >
          <Row gutter={16}>
            <Col xs={24} sm={12} md={6}>
              <Form.Item name="ip" label="آدرس IP">
                <Input placeholder="مثال: 192.168.1.1" allowClear />
              </Form.Item>
            </Col>
            <Col xs={24} sm={12} md={6}>
              <Form.Item name="username" label="نام کاربری">
                <Input placeholder="نام کاربری" allowClear />
              </Form.Item>
            </Col>
            <Col xs={24} sm={12} md={6}>
              <Form.Item name="from_date" label="از تاریخ">
                <DatePicker style={{ width: '100%' }} placeholder="انتخاب تاریخ" />
              </Form.Item>
            </Col>
            <Col xs={24} sm={12} md={6}>
              <Form.Item name="to_date" label="تا تاریخ">
                <DatePicker style={{ width: '100%' }} placeholder="انتخاب تاریخ" />
              </Form.Item>
            </Col>
          </Row>
          <Space>
            <Button type="primary" htmlType="submit" icon={<SearchOutlined />}>
              جستجو
            </Button>
            <Button onClick={handleReset}>پاک کردن فیلترها</Button>
          </Space>
        </Form>
      )}

      <DataTable columns={columns} data={data} loading={isLoading} rowKey="id" />
    </>
  );
}
