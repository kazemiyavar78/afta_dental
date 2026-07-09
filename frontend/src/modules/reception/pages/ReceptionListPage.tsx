// این ماژول نمونه است؛ باقی فیلدها/صفحات طبق همین الگو در فازهای بعدی اضافه می‌شوند.

import { Button } from 'antd';
import { PlusOutlined } from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import type { ColumnsType } from 'antd/es/table';
import { PageHeader } from '@/platform/components/PageHeader';
import { DataTable } from '@/platform/components/DataTable';
import { PermissionGuard } from '@/platform/auth/PermissionGuard';
import { useApiQuery } from '@/platform/hooks/useApiQuery';
import { fetchReceptions } from '../api';
import type { Reception } from '../types';

export function ReceptionListPage() {
  const navigate = useNavigate();
  const { data = [], isLoading } = useApiQuery({
    queryKey: ['receptions'],
    queryFn: fetchReceptions,
  });

  const columns: ColumnsType<Reception> = [
    { title: 'نام بیمار', dataIndex: 'patient_name', key: 'patient_name' },
    { title: 'پزشک', dataIndex: 'doctor_id', key: 'doctor_id' },
    { title: 'تاریخ', dataIndex: 'reception_date', key: 'reception_date' },
    { title: 'وضعیت', dataIndex: 'status', key: 'status' },
  ];

  return (
    <>
      <PageHeader
        title="پذیرش بیماران"
        extra={
          <PermissionGuard permission="reception.create">
            <Button type="primary" icon={<PlusOutlined />} onClick={() => navigate('/reception/new')}>
              پذیرش جدید
            </Button>
          </PermissionGuard>
        }
      />
      <DataTable columns={columns} data={data} loading={isLoading} rowKey="id" />
    </>
  );
}
