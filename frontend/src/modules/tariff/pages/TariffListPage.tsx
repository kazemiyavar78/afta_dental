// این ماژول نمونه است؛ باقی فیلدها/صفحات طبق همین الگو در فازهای بعدی اضافه می‌شوند.

import { Button } from 'antd';
import { PlusOutlined } from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import type { ColumnsType } from 'antd/es/table';
import { PageHeader } from '@/platform/components/PageHeader';
import { DataTable } from '@/platform/components/DataTable';
import { PermissionGuard } from '@/platform/auth/PermissionGuard';
import { useApiQuery } from '@/platform/hooks/useApiQuery';
import { fetchTariffs } from '../api';
import type { Tariff } from '../types';

export function TariffListPage() {
  const navigate = useNavigate();
  const { data = [], isLoading } = useApiQuery({
    queryKey: ['tariffs'],
    queryFn: fetchTariffs,
  });

  const columns: ColumnsType<Tariff> = [
    { title: 'نام', dataIndex: 'name', key: 'name' },
    { title: 'مبلغ', dataIndex: 'amount', key: 'amount' },
  ];

  return (
    <>
      <PageHeader
        title="تعرفه‌ها"
        extra={
          <PermissionGuard permission="tariff.create">
            <Button type="primary" icon={<PlusOutlined />} onClick={() => navigate('/tariff/new')}>
              تعرفه جدید
            </Button>
          </PermissionGuard>
        }
      />
      <DataTable columns={columns} data={data} loading={isLoading} rowKey="id" />
    </>
  );
}
