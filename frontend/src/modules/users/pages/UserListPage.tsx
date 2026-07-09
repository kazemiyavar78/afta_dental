import { Button, Tag } from 'antd';
import { PlusOutlined, EditOutlined } from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import type { ColumnsType } from 'antd/es/table';
import { PageHeader } from '@/platform/components/PageHeader';
import { DataTable } from '@/platform/components/DataTable';
import { PermissionGuard } from '@/platform/auth/PermissionGuard';
import { useApiQuery } from '@/platform/hooks/useApiQuery';
import { fetchUsers } from '../api';
import type { User } from '../types';

/** صفحه لیست کاربران */
export function UserListPage() {
  const navigate = useNavigate();
  const { data: users = [], isLoading } = useApiQuery({
    queryKey: ['users'],
    queryFn: fetchUsers,
  });

  const columns: ColumnsType<User> = [
    { title: 'نام کاربری', dataIndex: 'username', key: 'username' },
    { title: 'نام', dataIndex: 'name', key: 'name' },
    { title: 'نام خانوادگی', dataIndex: 'family', key: 'family' },
    { title: 'تلفن', dataIndex: 'phone_number', key: 'phone_number' },
    { title: 'نقش', dataIndex: 'role_name', key: 'role_name' },
    {
      title: 'وضعیت',
      key: 'status',
      render: (_, record) => (
        <Tag color={record.is_active ? 'green' : 'red'}>
          {record.is_active ? 'فعال' : 'غیرفعال'}
        </Tag>
      ),
    },
    {
      title: 'عملیات',
      key: 'actions',
      render: (_, record) => (
        <PermissionGuard permission="users.update">
          <Button
            type="link"
            icon={<EditOutlined />}
            onClick={() => navigate(`/users/${record.id}/edit`)}
          >
            ویرایش
          </Button>
        </PermissionGuard>
      ),
    },
  ];

  return (
    <>
      <PageHeader
        title="مدیریت کاربران"
        extra={
          <PermissionGuard permission="users.create">
            <Button type="primary" icon={<PlusOutlined />} onClick={() => navigate('/users/new')}>
              کاربر جدید
            </Button>
          </PermissionGuard>
        }
      />
      <DataTable columns={columns} data={users} loading={isLoading} rowKey="id" />
    </>
  );
}
