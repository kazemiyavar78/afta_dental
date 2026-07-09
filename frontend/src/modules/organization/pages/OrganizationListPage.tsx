// این ماژول نمونه است؛ باقی فیلدها/صفحات طبق همین الگو در فازهای بعدی اضافه می‌شوند.

import { Button } from 'antd';
import { PlusOutlined } from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import type { ColumnsType } from 'antd/es/table';
import { PageHeader } from '@/platform/components/PageHeader';
import { DataTable } from '@/platform/components/DataTable';
import { PermissionGuard } from '@/platform/auth/PermissionGuard';
import { useApiQuery } from '@/platform/hooks/useApiQuery';
import { fetchOrganizations } from '../api';
import type { Organization } from '../types';

export function OrganizationListPage() {
  const navigate = useNavigate();
  const { data = [], isLoading } = useApiQuery({
    queryKey: ['organizations'],
    queryFn: fetchOrganizations,
  });

  const columns: ColumnsType<Organization> = [
    { title: 'شناسه', dataIndex: 'id', key: 'id' },
    { title: 'نام', dataIndex: 'name', key: 'name' },
  ];

  return (
    <>
      <PageHeader
        title="سازمان‌ها"
        extra={
          <PermissionGuard permission="organization.create">
            <Button type="primary" icon={<PlusOutlined />} onClick={() => navigate('/organization/new')}>
              سازمان جدید
            </Button>
          </PermissionGuard>
        }
      />
      <DataTable columns={columns} data={data} loading={isLoading} rowKey="id" />
    </>
  );
}
