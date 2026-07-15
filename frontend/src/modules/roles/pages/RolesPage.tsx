import { useEffect, useState } from 'react';
import { Button, Checkbox, Form, Input, Modal, Space, Tag, Typography } from 'antd';
import { DeleteOutlined, EditOutlined, PlusOutlined, SafetyOutlined } from '@ant-design/icons';
import { useForm, Controller } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import type { ColumnsType } from 'antd/es/table';
import { PageHeader } from '@/platform/components/PageHeader';
import { DataTable } from '@/platform/components/DataTable';
import { PermissionGuard } from '@/platform/auth/PermissionGuard';
import { confirmDialog } from '@/platform/components/ConfirmDialog';
import { useApiQuery } from '@/platform/hooks/useApiQuery';
import { useApiMutation } from '@/platform/hooks/useApiMutation';
import {
  createRole,
  deleteRole,
  fetchPermissions,
  fetchRoles,
  updateRole,
} from '../api';
import { roleSchema, type RoleFormValues } from '../hooks';
import type { RoleDetail } from '../types';

const emptyFormValues: RoleFormValues = {
  name: '',
  description: '',
  permission_ids: [],
};

/** صفحه مدیریت نقش‌ها و انتصاب مجوزها */
export function RolesPage() {
  const [modalOpen, setModalOpen] = useState(false);
  const [editing, setEditing] = useState<RoleDetail | null>(null);

  const { data: roles = [], isLoading, refetch } = useApiQuery({
    queryKey: ['roles'],
    queryFn: fetchRoles,
  });

  const { data: permissions = [], isLoading: loadingPermissions } = useApiQuery({
    queryKey: ['permissions'],
    queryFn: fetchPermissions,
  });

  const {
    control,
    handleSubmit,
    reset,
    setError,
    formState: { errors },
  } = useForm<RoleFormValues>({
    resolver: zodResolver(roleSchema),
    defaultValues: emptyFormValues,
  });

  useEffect(() => {
    if (!modalOpen) return;
    if (editing) {
      reset({
        name: editing.name,
        description: editing.description ?? '',
        permission_ids: editing.permission_ids ?? [],
      });
    } else {
      reset(emptyFormValues);
    }
  }, [editing, modalOpen, reset]);

  const createMutation = useApiMutation({
    mutationFn: createRole,
    successMessage: 'نقش با موفقیت ایجاد شد',
    setError,
    onSuccess: () => {
      setModalOpen(false);
      refetch();
    },
  });

  const updateMutation = useApiMutation({
    mutationFn: (values: RoleFormValues) => updateRole(editing!.id, values),
    successMessage: 'نقش با موفقیت به‌روزرسانی شد',
    setError,
    onSuccess: () => {
      setModalOpen(false);
      setEditing(null);
      refetch();
    },
  });

  const deleteMutation = useApiMutation({
    mutationFn: deleteRole,
    successMessage: 'نقش با موفقیت حذف شد',
    onSuccess: () => refetch(),
  });

  const openCreate = () => {
    setEditing(null);
    setModalOpen(true);
  };

  const openEdit = (record: RoleDetail) => {
    setEditing(record);
    setModalOpen(true);
  };

  const closeModal = () => {
    setModalOpen(false);
    setEditing(null);
  };

  const isPending = createMutation.isPending || updateMutation.isPending;
  const permissionOptions = permissions.map((p) => ({
    label: `${p.description || p.name} (${p.name})`,
    value: p.id,
  }));

  const columns: ColumnsType<RoleDetail> = [
    { title: 'نام', dataIndex: 'name', key: 'name', width: 160 },
    { title: 'توضیحات', dataIndex: 'description', key: 'description', ellipsis: true },
    {
      title: 'مجوزها',
      key: 'permissions',
      render: (_, record) => (
        <Typography.Text type="secondary">
          {record.permission_ids?.length ?? 0} مجوز
        </Typography.Text>
      ),
    },
    {
      title: 'یکپارچگی',
      key: 'integrity',
      width: 120,
      render: (_, record) => (
        <Tag color={record.integrity_ok ? 'green' : 'red'}>
          {record.integrity_ok ? 'سالم' : 'نقص'}
        </Tag>
      ),
    },
    {
      title: 'عملیات',
      key: 'actions',
      width: 200,
      render: (_, record) => (
        <>
          <PermissionGuard permission="roles.update">
            <Button type="link" icon={<EditOutlined />} onClick={() => openEdit(record)}>
              ویرایش
            </Button>
          </PermissionGuard>
          <PermissionGuard permission="roles.delete">
            <Button
              type="link"
              danger
              icon={<DeleteOutlined />}
              disabled={record.name === 'Admin'}
              loading={deleteMutation.isPending}
              onClick={() =>
                confirmDialog({
                  title: 'حذف نقش',
                  content: `آیا از حذف نقش «${record.name}» مطمئن هستید؟`,
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
        title="نقش‌ها و مجوزها"
        extra={
          <PermissionGuard permission="roles.create">
            <Button type="primary" icon={<PlusOutlined />} onClick={openCreate}>
              نقش جدید
            </Button>
          </PermissionGuard>
        }
      />
      <DataTable columns={columns} data={roles} loading={isLoading} rowKey="id" />

      <Modal
        title={editing ? `ویرایش نقش: ${editing.name}` : 'نقش جدید'}
        open={modalOpen}
        onCancel={closeModal}
        footer={null}
        destroyOnHidden
        width={720}
      >
        <Form
          layout="vertical"
          onFinish={handleSubmit((values) => {
            if (editing) {
              updateMutation.mutate(values);
            } else {
              createMutation.mutate(values);
            }
          })}
        >
          <Form.Item
            label="نام نقش"
            validateStatus={errors.name ? 'error' : ''}
            help={errors.name?.message}
          >
            <Controller
              name="name"
              control={control}
              render={({ field }) => (
                <Input {...field} disabled={editing?.name === 'Admin'} />
              )}
            />
          </Form.Item>
          <Form.Item
            label="توضیحات"
            validateStatus={errors.description ? 'error' : ''}
            help={errors.description?.message}
          >
            <Controller
              name="description"
              control={control}
              render={({ field }) => <Input.TextArea {...field} rows={2} />}
            />
          </Form.Item>
          <Form.Item
            label={
              <Space>
                <SafetyOutlined />
                <span>مجوزها</span>
              </Space>
            }
            validateStatus={errors.permission_ids ? 'error' : ''}
            help={errors.permission_ids?.message}
          >
            {loadingPermissions ? (
              <Typography.Text type="secondary">در حال بارگذاری مجوزها...</Typography.Text>
            ) : (
              <Controller
                name="permission_ids"
                control={control}
                render={({ field }) => (
                  <Checkbox.Group
                    style={{ display: 'flex', flexDirection: 'column', gap: 8, maxHeight: 320, overflow: 'auto' }}
                    options={permissionOptions}
                    value={field.value}
                    onChange={field.onChange}
                  />
                )}
              />
            )}
          </Form.Item>
          <Space>
            <Button type="primary" htmlType="submit" loading={isPending}>
              ذخیره
            </Button>
            <Button onClick={closeModal}>انصراف</Button>
          </Space>
        </Form>
      </Modal>
    </>
  );
}
