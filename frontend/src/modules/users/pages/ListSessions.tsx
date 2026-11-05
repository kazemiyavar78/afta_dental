import { Button } from 'antd';
import { ArrowRightOutlined } from '@ant-design/icons';
import { useNavigate, useParams } from 'react-router-dom';
import { PageHeader } from '@/platform/components/PageHeader';
import { useApiQuery } from '@/platform/hooks/useApiQuery';
import { fetchSessionPage, fetchUser, scopeSessionQuery } from '../api';
import { SessionTable } from '../components/SessionTable';

/**
 * صفحه نشست‌های یک کاربر. همه ستون‌های نمایشی فیلتر و مرتب‌سازی سمت سرور دارند.
 * شناسه کاربر در URL مسیر است و به‌صورت بازه عددی به سرور می‌رود، نه به‌صورت ستون.
 */
export function ListSessions() {
  const navigate = useNavigate();
  const { id } = useParams<{ id: string }>();
  const userId = Number(id);
  const enabled = Number.isFinite(userId) && userId > 0;

  const { data: user } = useApiQuery({
    queryKey: ['users', userId],
    queryFn: () => fetchUser(userId),
    enabled,
  });

  const subtitle = user
    ? `نشست‌های فعال ${user.name} ${user.family} (@${user.username})`
    : 'مشاهده و مدیریت نشست‌های فعال کاربر';

  return (
    <>
      <PageHeader
        title="نشست‌های کاربر"
        subtitle={subtitle}
        extra={
          <Button htmlType="button" icon={<ArrowRightOutlined />} onClick={() => navigate('/users')}>
            بازگشت
          </Button>
        }
      />
      {enabled && (
        <SessionTable
          queryKey={['sessions', 'user', userId]}
          fetchPage={(query) => fetchSessionPage(scopeSessionQuery(query, userId, true))}
        />
      )}
    </>
  );
}
