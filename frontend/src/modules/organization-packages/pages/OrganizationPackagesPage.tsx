import { useEffect, useState } from 'react';
import {
  Button,
  Form,
  Input,
  InputNumber,
  Modal,
  Space,
} from 'antd';
import { DeleteOutlined, EditOutlined, PlusOutlined } from '@ant-design/icons';
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
  createOrganizationPackage,
  deleteOrganizationPackage,
  fetchOrganizationPackages,
  updateOrganizationPackage,
} from '../api';
import { organizationPackageSchema, type OrganizationPackageFormValues } from '../hooks';
import type { OrganizationPackage } from '../types';

const emptyFormValues: OrganizationPackageFormValues = {
  package_name: '',
  package_description: '',
  technical_coefficient: 0,
  technical_professional_coefficient: 0,
  consumption_coefficient: 0,
  subsidy_percentage: 0,
  supplementary_percentage: 0,
  organization_percentage: 0,
};

/** صفحه مدیریت بسته‌های تعرفه سازمان */
export function OrganizationPackagesPage() {
  const [modalOpen, setModalOpen] = useState(false);
  const [editing, setEditing] = useState<OrganizationPackage | null>(null);

  const { data = [], isLoading, refetch } = useApiQuery({
    queryKey: ['organization-packages'],
    queryFn: fetchOrganizationPackages,
  });

  const {
    control,
    handleSubmit,
    reset,
    setError,
    formState: { errors },
  } = useForm<OrganizationPackageFormValues>({
    resolver: zodResolver(organizationPackageSchema),
    defaultValues: emptyFormValues,
  });

  useEffect(() => {
    if (!modalOpen) return;
    if (editing) {
      reset({
        package_name: editing.package_name,
        package_description: editing.package_description,
        technical_coefficient: editing.technical_coefficient,
        technical_professional_coefficient: editing.technical_professional_coefficient,
        consumption_coefficient: editing.consumption_coefficient,
        subsidy_percentage: editing.subsidy_percentage,
        supplementary_percentage: editing.supplementary_percentage,
        organization_percentage: editing.organization_percentage,
      });
    } else {
      reset(emptyFormValues);
    }
  }, [editing, modalOpen, reset]);

  const createMutation = useApiMutation({
    mutationFn: createOrganizationPackage,
    successMessage: 'بسته تعرفه با موفقیت ایجاد شد',
    setError,
    onSuccess: () => {
      setModalOpen(false);
      refetch();
    },
  });

  const updateMutation = useApiMutation({
    mutationFn: (values: OrganizationPackageFormValues) =>
      updateOrganizationPackage(editing!.id, values),
    successMessage: 'بسته تعرفه با موفقیت به‌روزرسانی شد',
    setError,
    onSuccess: () => {
      setModalOpen(false);
      setEditing(null);
      refetch();
    },
  });

  const deleteMutation = useApiMutation({
    mutationFn: deleteOrganizationPackage,
    successMessage: 'بسته تعرفه با موفقیت حذف شد',
    onSuccess: () => refetch(),
  });

  const openCreate = () => {
    setEditing(null);
    setModalOpen(true);
  };

  const openEdit = (record: OrganizationPackage) => {
    setEditing(record);
    setModalOpen(true);
  };

  const closeModal = () => {
    setModalOpen(false);
    setEditing(null);
  };

  const isPending = createMutation.isPending || updateMutation.isPending;

  const numberField = (
    name: keyof OrganizationPackageFormValues,
    label: string,
    max?: number,
  ) => (
    <Form.Item
      label={label}
      validateStatus={errors[name] ? 'error' : ''}
      help={errors[name]?.message as string | undefined}
    >
      <Controller
        name={name}
        control={control}
        render={({ field }) => (
          <InputNumber
            min={0}
            max={max}
            style={{ width: 160 }}
            value={field.value as number}
            onChange={(v) => field.onChange(v ?? 0)}
          />
        )}
      />
    </Form.Item>
  );

  const columns: ColumnsType<OrganizationPackage> = [
    { title: 'نام بسته', dataIndex: 'package_name', key: 'package_name' },
    { title: 'توضیحات', dataIndex: 'package_description', key: 'package_description', ellipsis: true },
    { title: 'یارانه٪', dataIndex: 'subsidy_percentage', key: 'subsidy_percentage', width: 90 },
    { title: 'تکمیلی٪', dataIndex: 'supplementary_percentage', key: 'supplementary_percentage', width: 90 },
    {
      title: 'عملیات',
      key: 'actions',
      width: 200,
      render: (_, record) => (
        <>
          <PermissionGuard permission="organization_packages.update">
            <Button type="link" icon={<EditOutlined />} onClick={() => openEdit(record)}>
              ویرایش
            </Button>
          </PermissionGuard>
          <PermissionGuard permission="organization_packages.delete">
            <Button
              type="link"
              danger
              icon={<DeleteOutlined />}
              loading={deleteMutation.isPending}
              onClick={() =>
                confirmDialog({
                  title: 'حذف بسته تعرفه',
                  content: `آیا از حذف بسته «${record.package_name}» مطمئن هستید؟`,
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
        title="بسته‌های تعرفه سازمان"
        extra={
          <PermissionGuard permission="organization_packages.create">
            <Button type="primary" icon={<PlusOutlined />} onClick={openCreate}>
              بسته جدید
            </Button>
          </PermissionGuard>
        }
      />
      <DataTable columns={columns} data={data} loading={isLoading} rowKey="id" />

      <Modal
        title={editing ? `ویرایش بسته: ${editing.package_name}` : 'بسته تعرفه جدید'}
        open={modalOpen}
        onCancel={closeModal}
        footer={null}
        destroyOnHidden
        width={780}
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
            label="نام بسته"
            validateStatus={errors.package_name ? 'error' : ''}
            help={errors.package_name?.message}
          >
            <Controller name="package_name" control={control} render={({ field }) => <Input {...field} />} />
          </Form.Item>
          <Form.Item
            label="توضیحات"
            validateStatus={errors.package_description ? 'error' : ''}
            help={errors.package_description?.message}
          >
            <Controller
              name="package_description"
              control={control}
              render={({ field }) => <Input.TextArea {...field} rows={2} />}
            />
          </Form.Item>

          <Space wrap size="large" style={{ display: 'flex', marginBottom: 8 }}>
            {numberField('technical_coefficient', 'ضریب فنی بسته')}
            {numberField('technical_professional_coefficient', 'ضریب حرفه‌ای بسته')}
            {numberField('consumption_coefficient', 'ضریب مصرفی بسته')}
          </Space>

          <Space wrap size="large" style={{ display: 'flex', marginBottom: 8 }}>
            {numberField('subsidy_percentage', 'درصد یارانه', 100)}
            {numberField('supplementary_percentage', 'درصد تکمیلی', 100)}
            {numberField('organization_percentage', 'درصد سهم سازمان در بسته', 100)}
          </Space>

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
