import { useEffect, useMemo, useState } from 'react';
import { Button, Checkbox, Divider, Form, Input, Modal, Space, Tag, Typography } from 'antd';
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
import type { Permission, RoleDetail } from '../types';

const emptyFormValues: RoleFormValues = {
  name: '',
  description: '',
  permission_ids: [],
};

/** برچسب فارسی ماژول‌ها بر اساس پیشوند نام مجوز */
const MODULE_LABELS: Record<string, string> = {
  bank_account: 'حساب بانکی',
  fund: 'صندوق',
  logs: 'لاگ‌ها',
  organization: 'سازمان',
  organization_packages: 'بسته‌های سازمانی',
  patient: 'بیماران',
  reception: 'پذیرش',
  regulation: 'ضوابط',
  roles: 'نقش‌ها',
  security: 'امنیت',
  services: 'خدمات',
  special_code: 'کد خاص',
  tariff: 'تعرفه',
  users: 'کاربران',
  wallet: 'کیف پول',
};

/**
 * پیشوند ماژول را از نام مجوز استخراج می‌کند.
 * @param permissionName نام کامل مجوز مثل bank_account.read
 */
function permissionModule(permissionName: string): string {
  const idx = permissionName.indexOf('.');
  return idx === -1 ? permissionName : permissionName.slice(0, idx);
}

/**
 * آیا این مجوز، مجوز «خواندن» ماژول است؟
 * @param permissionName نام کامل مجوز
 */
function isModuleReadPermission(permissionName: string): boolean {
  return permissionName.endsWith('.read');
}

/**
 * گروه‌بندی مجوزها بر اساس پیشوند ماژول (مرتب‌شده بر اساس کلید).
 * @param permissions لیست مجوزهای سیستم
 */
function groupPermissionsByModule(permissions: Permission[]): { module: string; label: string; items: Permission[] }[] {
  const groups = new Map<string, Permission[]>();
  for (const p of permissions) {
    const module = permissionModule(p.name);
    const list = groups.get(module) ?? [];
    list.push(p);
    groups.set(module, list);
  }

  return [...groups.entries()]
    .sort(([a], [b]) => a.localeCompare(b))
    .map(([module, items]) => ({
      module,
      label: MODULE_LABELS[module] ?? module,
      items: [...items].sort((a, b) => {
        const aRead = isModuleReadPermission(a.name) ? 0 : 1;
        const bRead = isModuleReadPermission(b.name) ? 0 : 1;
        if (aRead !== bRead) return aRead - bRead;
        return a.name.localeCompare(b.name);
      }),
    }));
}

/**
 * آیا نقش تمام مجوزهای سیستم را دارد (ادمین)؟
 * @param role نقش
 * @param allPermissionIds شناسه تمام مجوزهای سیستم
 */
function roleHasAllPermissions(role: RoleDetail, allPermissionIds: number[]): boolean {
  if (role.name === 'Admin') return true;
  if (allPermissionIds.length === 0) return false;
  const owned = new Set(role.permission_ids ?? []);
  return allPermissionIds.every((id) => owned.has(id));
}

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
    watch,
    setValue,
    formState: { errors },
  } = useForm<RoleFormValues>({
    resolver: zodResolver(roleSchema),
    defaultValues: emptyFormValues,
  });

  const selectedIds = watch('permission_ids') ?? [];

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
  const permissionGroups = useMemo(() => groupPermissionsByModule(permissions), [permissions]);
  const allPermissionIds = useMemo(() => permissions.map((p) => p.id), [permissions]);

  const readIdByModule = useMemo(() => {
    const map = new Map<string, number>();
    for (const p of permissions) {
      if (isModuleReadPermission(p.name)) {
        map.set(permissionModule(p.name), p.id);
      }
    }
    return map;
  }, [permissions]);

  /**
   * تغییر انتخاب مجوز با قانون: بدون read، بقیه ماژول قابل انتخاب نیستند.
   * @param permission مجوز هدف
   * @param checked وضعیت جدید چک‌باکس
   * @param currentIds شناسه‌های انتخاب‌شده فعلی
   */
  const togglePermission = (permission: Permission, checked: boolean, currentIds: number[]) => {
    const module = permissionModule(permission.name);
    const readId = readIdByModule.get(module);
    const moduleIds = permissions.filter((p) => permissionModule(p.name) === module).map((p) => p.id);

    if (isModuleReadPermission(permission.name)) {
      if (!checked) {
        setValue(
          'permission_ids',
          currentIds.filter((id) => !moduleIds.includes(id)),
          { shouldDirty: true },
        );
        return;
      }
      setValue('permission_ids', [...new Set([...currentIds, permission.id])], { shouldDirty: true });
      return;
    }

    if (!checked) {
      setValue(
        'permission_ids',
        currentIds.filter((id) => id !== permission.id),
        { shouldDirty: true },
      );
      return;
    }

    if (readId != null && !currentIds.includes(readId)) {
      return;
    }
    setValue('permission_ids', [...new Set([...currentIds, permission.id])], { shouldDirty: true });
  };

  const columns: ColumnsType<RoleDetail> = [
    { title: 'نام', dataIndex: 'name', key: 'name', width: 160 },
    { title: 'توضیحات', dataIndex: 'description', key: 'description', ellipsis: true },
    {
      title: 'مجوزها',
      key: 'permissions',
      render: (_, record) => (
        <Space size={4}>
          <Typography.Text type="secondary">
            {record.permission_ids?.length ?? 0} مجوز
          </Typography.Text>
          {roleHasAllPermissions(record, allPermissionIds) && <Tag color="blue">ادمین</Tag>}
        </Space>
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
      render: (_, record) => {
        const isAdminRole = roleHasAllPermissions(record, allPermissionIds);
        return (
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
                disabled={isAdminRole}
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
        );
      },
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
        width={760}
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
                <Input
                  {...field}
                  disabled={editing != null && roleHasAllPermissions(editing, allPermissionIds)}
                />
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
              <div style={{ maxHeight: 360, overflow: 'auto', paddingInlineEnd: 4 }}>
                {permissionGroups.map((group) => {
                  const readId = readIdByModule.get(group.module);
                  const hasRead = readId == null || selectedIds.includes(readId);
                  return (
                    <div key={group.module} style={{ marginBottom: 12 }}>
                      <Divider plain style={{ margin: '8px 0' }}>
                        <Typography.Text strong>
                          {group.label}{' '}
                          <Typography.Text type="secondary" style={{ fontWeight: 400 }}>
                            ({group.module})
                          </Typography.Text>
                        </Typography.Text>
                      </Divider>
                      <Space direction="vertical" size={4} style={{ width: '100%' }}>
                        {group.items.map((p) => {
                          const isRead = isModuleReadPermission(p.name);
                          const disabled = !isRead && readId != null && !hasRead;
                          return (
                            <Checkbox
                              key={p.id}
                              checked={selectedIds.includes(p.id)}
                              disabled={disabled}
                              onChange={(e) => togglePermission(p, e.target.checked, selectedIds)}
                            >
                              {p.description || p.name}{' '}
                              <Typography.Text type="secondary">({p.name})</Typography.Text>
                            </Checkbox>
                          );
                        })}
                      </Space>
                    </div>
                  );
                })}
              </div>
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
