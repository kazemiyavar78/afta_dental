// این ماژول نمونه است؛ باقی فیلدها/صفحات طبق همین الگو در فازهای بعدی اضافه می‌شوند.

import { Button } from 'antd';
import { PlusOutlined } from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import type { ColumnsType } from 'antd/es/table';
import { PageHeader } from '@/platform/components/PageHeader';
import { DataTable } from '@/platform/components/DataTable';
import { PermissionGuard } from '@/platform/auth/PermissionGuard';
import { useApiQuery } from '@/platform/hooks/useApiQuery';
import { fetchFunds } from '../api';
import type { Fund } from '../types';

export function FundListPage() {
  const navigate = useNavigate();
  const { data = [], isLoading } = useApiQuery({
    queryKey: ['funds'],
    queryFn: fetchFunds,
  });

  const columns: ColumnsType<Fund> = [
    { title: 'نام', dataIndex: 'name', key: 'name' },
    { title: 'موجودی', dataIndex: 'balance', key: 'balance' },
  ];

  return (
    <>
      <PageHeader
        title="صندوق‌ها"
        extra={
          <PermissionGuard permission="fund.create">
            <Button type="primary" icon={<PlusOutlined />} onClick={() => navigate('/fund/new')}>
              صندوق جدید
            </Button>
          </PermissionGuard>
        }
      />
      <DataTable columns={columns} data={data} loading={isLoading} rowKey="id" />
    </>
  );
}
