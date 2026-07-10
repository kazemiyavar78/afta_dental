import {
  Avatar,
  Button,
  Card,
  Col,
  Descriptions,
  Form,
  Input,
  Row,
  Space,
  Spin,
  Tag,
  Typography,
} from 'antd';
import {
  DeleteOutlined,
  LockOutlined,
  DesktopOutlined,
  UserOutlined,
} from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';
import dayjs from 'dayjs';
import { PageHeader } from '@/platform/components/PageHeader';
import { DataTable } from '@/platform/components/DataTable';
import { confirmDialog } from '@/platform/components/ConfirmDialog';
import { useApiQuery } from '@/platform/hooks/useApiQuery';
import { useApiMutation } from '@/platform/hooks/useApiMutation';
import { changePassword, deleteSession, fetchUserProfile } from '../api';
import type { ChangePasswordPayload, Session } from '../types';

/** قالب تاریخ شمسی برای نمایش در پروفایل */
function formatDate(value: string | null | undefined): string {
  if (!value) return '—';
  return dayjs(value).calendar('jalali').locale('fa').format('YYYY/MM/DD HH:mm');
}

type PasswordFormValues = ChangePasswordPayload & {
  confirm_password: string;
};

/** صفحه پروفایل کاربر: اطلاعات حساب، تغییر رمز و نشست‌های فعال */
export function ProfilePage() {
  const [passwordForm] = Form.useForm<PasswordFormValues>();

  const { data, isLoading, refetch } = useApiQuery({
    queryKey: ['profile'],
    queryFn: fetchUserProfile,
  });

  const changePasswordMutation = useApiMutation({
    mutationFn: changePassword,
    successMessage: 'رمز عبور با موفقیت تغییر یافت',
    onSuccess: () => passwordForm.resetFields(),
  });

  const deleteSessionMutation = useApiMutation({
    mutationFn: deleteSession,
    successMessage: 'نشست با موفقیت حذف شد',
    onSuccess: () => refetch(),
  });

  const user = data?.user;
  const sessions = data?.sessions ?? [];

  const sessionColumns: ColumnsType<Session> = [
    {
      title: 'آدرس IP',
      dataIndex: 'ip',
      key: 'ip',
    },
    {
      title: 'مرورگر / دستگاه',
      dataIndex: 'browser',
      key: 'browser',
      ellipsis: true,
    },
    {
      title: 'زمان ورود',
      dataIndex: 'creation_time',
      key: 'creation_time',
      render: (value: string) => formatDate(value),
    },
    {
      title: 'عملیات',
      key: 'actions',
      width: 120,
      render: (_, record) => (
        <Button
          type="link"
          danger
          icon={<DeleteOutlined />}
          loading={deleteSessionMutation.isPending}
          onClick={() =>
            confirmDialog({
              title: 'حذف نشست',
              content: 'آیا از حذف این نشست مطمئن هستید؟',
              okType: 'danger',
              onConfirm: () => deleteSessionMutation.mutateAsync(record.id),
            })
          }
        >
          حذف
        </Button>
      ),
    },
  ];

  if (isLoading) {
    return (
      <div style={{ display: 'flex', justifyContent: 'center', padding: 48 }}>
        <Spin size="large" tip="در حال بارگذاری پروفایل..." />
      </div>
    );
  }

  return (
    <>
      <PageHeader title="پروفایل کاربر" subtitle="مشاهده اطلاعات حساب و مدیریت نشست‌ها" />

      <Row gutter={[24, 24]}>
        <Col xs={24} lg={14}>
          <Card>
            <Space align="start" size={20} style={{ marginBottom: 24 }}>
              <Avatar size={72} icon={<UserOutlined />} style={{ backgroundColor: '#1677ff' }} />
              <div>
                <Typography.Title level={4} style={{ margin: 0 }}>
                  {user?.name} {user?.family}
                </Typography.Title>
                <Typography.Text type="secondary">@{user?.username}</Typography.Text>
                <div style={{ marginTop: 8 }}>
                  <Space wrap>
                    <Tag color="blue">{user?.role_name}</Tag>
                    <Tag color={user?.is_active ? 'green' : 'red'}>
                      {user?.is_active ? 'فعال' : 'غیرفعال'}
                    </Tag>
                    {user?.is_locked && <Tag color="orange">قفل‌شده</Tag>}
                  </Space>
                </div>
              </div>
            </Space>

            <Descriptions column={{ xs: 1, sm: 2 }} bordered size="small">
              <Descriptions.Item label="نام کاربری">{user?.username}</Descriptions.Item>
              <Descriptions.Item label="نقش">{user?.role_name}</Descriptions.Item>
              <Descriptions.Item label="نام">{user?.name}</Descriptions.Item>
              <Descriptions.Item label="نام خانوادگی">{user?.family}</Descriptions.Item>
              <Descriptions.Item label="تلفن">{user?.phone_number || '—'}</Descriptions.Item>
              <Descriptions.Item label="کد نظام پزشکی">
                {user?.medical_code || '—'}
              </Descriptions.Item>
              <Descriptions.Item label="آدرس" span={2}>
                {user?.address || '—'}
              </Descriptions.Item>
              <Descriptions.Item label="آخرین ورود" span={2}>
                {formatDate(user?.last_login_at)}
              </Descriptions.Item>
            </Descriptions>
          </Card>
        </Col>

        <Col xs={24} lg={10}>
          <Card
            title={
              <Space>
                <LockOutlined />
                تغییر رمز عبور
              </Space>
            }
          >
            <Form
              form={passwordForm}
              layout="vertical"
              onFinish={(values) =>
                changePasswordMutation.mutate({
                  old_password: values.old_password,
                  new_password: values.new_password,
                })
              }
            >
              <Form.Item
                name="old_password"
                label="رمز عبور فعلی"
                rules={[{ required: true, message: 'رمز عبور فعلی را وارد کنید' }]}
              >
                <Input.Password placeholder="رمز عبور فعلی" autoComplete="current-password" />
              </Form.Item>
              <Form.Item
                name="new_password"
                label="رمز عبور جدید"
                rules={[
                  { required: true, message: 'رمز عبور جدید را وارد کنید' },
                  { min: 6, message: 'رمز عبور باید حداقل ۶ کاراکتر باشد' },
                ]}
              >
                <Input.Password placeholder="رمز عبور جدید" autoComplete="new-password" />
              </Form.Item>
              <Form.Item
                name="confirm_password"
                label="تکرار رمز عبور جدید"
                dependencies={['new_password']}
                rules={[
                  { required: true, message: 'تکرار رمز عبور را وارد کنید' },
                  ({ getFieldValue }) => ({
                    validator(_, value) {
                      if (!value || getFieldValue('new_password') === value) {
                        return Promise.resolve();
                      }
                      return Promise.reject(new Error('رمز عبور و تکرار آن یکسان نیستند'));
                    },
                  }),
                ]}
              >
                <Input.Password placeholder="تکرار رمز عبور جدید" autoComplete="new-password" />
              </Form.Item>
              <Form.Item style={{ marginBottom: 0 }}>
                <Button
                  type="primary"
                  htmlType="submit"
                  loading={changePasswordMutation.isPending}
                  block
                >
                  ذخیره رمز جدید
                </Button>
              </Form.Item>
            </Form>
          </Card>
        </Col>

        <Col span={24}>
          <Card
            title={
              <Space>
                <DesktopOutlined />
                نشست‌های فعال
                <Tag>{sessions.length}</Tag>
              </Space>
            }
          >
            <DataTable
              columns={sessionColumns}
              data={sessions}
              loading={isLoading}
              rowKey="id"
              pageSize={5}
            />
          </Card>
        </Col>
      </Row>
    </>
  );
}
