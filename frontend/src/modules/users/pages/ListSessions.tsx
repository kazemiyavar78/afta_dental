import { Button } from 'antd';
import { ArrowRightOutlined, DeleteOutlined } from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';
import dayjs from 'dayjs';
import { useNavigate, useParams } from 'react-router-dom';
import { PageHeader } from '@/platform/components/PageHeader';
import { DataTable } from '@/platform/components/DataTable';
import { confirmDialog } from '@/platform/components/ConfirmDialog';
import { useApiQuery } from '@/platform/hooks/useApiQuery';
import { useApiMutation } from '@/platform/hooks/useApiMutation';
import { deleteSession, fetchSessions, fetchUser } from '../api';
import type { Session } from '../types';

/** قالب تاریخ برای نمایش زمان ایجاد نشست */
function formatDate(value: string | null | undefined): string {
  if (!value) return '—';
  return dayjs(value).calendar('jalali').locale('fa').format('YYYY/MM/DD HH:mm');
}

/** صفحه لیست نشست‌های یک کاربر برای ادمین */
export function ListSessions() {
  const navigate = useNavigate();
  const { id } = useParams<{ id: string }>();
  const userId = Number(id);

  const { data: user, isLoading: loadingUser } = useApiQuery({
    queryKey: ['users', userId],
    queryFn: () => fetchUser(userId),
    enabled: Number.isFinite(userId) && userId > 0,
  });

  const {
    data: sessions = [],
    isLoading: loadingSessions,
    refetch,
  } = useApiQuery({
    queryKey: ['sessions', userId],
    queryFn: () => fetchSessions(userId),
    enabled: Number.isFinite(userId) && userId > 0,
  });

  const deleteSessionMutation = useApiMutation({
    mutationFn: deleteSession,
    successMessage: 'نشست با موفقیت حذف شد',
    onSuccess: () => refetch(),
  });

  const columns: ColumnsType<Session> = [
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

  const subtitle = user
    ? `نشست‌های فعال ${user.name} ${user.family} (@${user.username})`
    : 'مشاهده و مدیریت نشست‌های فعال کاربر';

  return (
    <>
      <PageHeader
        title="نشست‌های کاربر"
        subtitle={subtitle}
        extra={
          <Button icon={<ArrowRightOutlined />} onClick={() => navigate('/users')}>
            بازگشت
          </Button>
        }
      />
      <DataTable
        columns={columns}
        data={sessions}
        loading={loadingUser || loadingSessions}
        rowKey="id"
      />
    </>
  );
}
