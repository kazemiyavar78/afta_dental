// این ماژول نمونه است؛ باقی فیلدها/صفحات طبق همین الگو در فازهای بعدی اضافه می‌شوند.

import { useState } from 'react';
import { Button, Tag } from 'antd';
import { PlusOutlined, EditOutlined, DeleteOutlined, StopOutlined } from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import type { ColumnsType } from 'antd/es/table';
import { PageHeader } from '@/platform/components/PageHeader';
import { DataTable } from '@/platform/components/DataTable';
import { PermissionGuard } from '@/platform/auth/PermissionGuard';
import { confirmDialog } from '@/platform/components/ConfirmDialog';
import { useApiQuery } from '@/platform/hooks/useApiQuery';
import { useApiMutation } from '@/platform/hooks/useApiMutation';
import { deleteOrganization, fetchOrganizations } from '../api';
import { ExcludedServicesModal } from '../components/ExcludedServicesModal';
import type { Organization } from '../types';

/** صفحه لیست سازمان‌ها */
export function OrganizationListPage() {
  const navigate = useNavigate();
  const [excludedOrg, setExcludedOrg] = useState<Organization | null>(null);

  const { data = [], isLoading, refetch } = useApiQuery({
    queryKey: ['organizations'],
    queryFn: fetchOrganizations,
  });

  const deleteMutation = useApiMutation({
    mutationFn: deleteOrganization,
    successMessage: 'سازمان با موفقیت حذف شد',
    onSuccess: () => refetch(),
  });

  const columns: ColumnsType<Organization> = [
    { title: 'شناسه', dataIndex: 'id', key: 'id', width: 80 },
    { title: 'نام', dataIndex: 'name', key: 'name' },
    { title: 'بسته تعرفه', dataIndex: 'package_name', key: 'package_name' },
    { title: 'بسته مرکز', dataIndex: 'center_package_name', key: 'center_package_name' },
    {
      title: 'نوع',
      key: 'type',
      render: (_, record) => (
        <Tag color={record.is_takmili ? 'blue' : 'default'}>
          {record.is_takmili ? 'تکمیلی' : 'پایه'}
        </Tag>
      ),
    },
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
      title: 'آزاد',
      key: 'is_free',
      width: 80,
      render: (_, record) => (
        record.is_free ? <Tag color="gold">آزاد</Tag> : '—'
      ),
    },
    {
      title: 'عملیات',
      key: 'actions',
      width: 340,
      render: (_, record) => (
        <>
          <PermissionGuard permission="organization.update">
            <Button
              type="link"
              icon={<EditOutlined />}
              onClick={() => navigate(`/organization/${record.id}/edit`)}
            >
              ویرایش
            </Button>
          </PermissionGuard>
          <PermissionGuard permission="excluded_services.read">
            <Button
              type="link"
              icon={<StopOutlined />}
              onClick={() => setExcludedOrg(record)}
            >
              خدمات خارج‌شده
            </Button>
          </PermissionGuard>
          <PermissionGuard permission="organization.delete">
            <Button
              type="link"
              danger
              icon={<DeleteOutlined />}
              loading={deleteMutation.isPending}
              onClick={() =>
                confirmDialog({
                  title: 'حذف سازمان',
                  content: `آیا از حذف سازمان «${record.name}» مطمئن هستید؟`,
                  okType: 'danger',
                  onConfirm: () => deleteMutation.mutateAsync(record.id),
                })
              }
            >
              حذف
            </Button>
          </PermissionGuard>
        </>
      ),
    },
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
      <ExcludedServicesModal
        organization={excludedOrg}
        open={excludedOrg != null}
        onClose={() => setExcludedOrg(null)}
      />
    </>
  );
}
