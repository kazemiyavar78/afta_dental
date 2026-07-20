import { useEffect, useState } from 'react';
import { Button, Form, Input, InputNumber, Modal, Space, Switch, Tag } from 'antd';
import { PlusOutlined, EditOutlined, DeleteOutlined } from '@ant-design/icons';
import { useForm, Controller } from 'react-hook-form';
import { z } from 'zod';
import { zodResolver } from '@hookform/resolvers/zod';
import type { ColumnsType } from 'antd/es/table';
import { PageHeader } from '@/platform/components/PageHeader';
import { DataTable } from '@/platform/components/DataTable';
import { PermissionGuard } from '@/platform/auth/PermissionGuard';
import { confirmDialog } from '@/platform/components/ConfirmDialog';
import { useApiQuery } from '@/platform/hooks/useApiQuery';
import { useApiMutation } from '@/platform/hooks/useApiMutation';
import {
  createSpecialCode,
  deleteSpecialCode,
  fetchSpecialCodes,
  updateSpecialCode,
} from '../api';
import type { SpecialCode, SpecialCodePayload } from '../types';

const schema = z.object({
  code: z.string().min(1, 'کد الزامی است'),
  name: z.string().min(1, 'نام الزامی است'),
  description: z.string(),
  percentage: z.number().min(0).max(100),
  is_active: z.boolean(),
});

type FormValues = z.infer<typeof schema>;

const emptyValues: FormValues = {
  code: '',
  name: '',
  description: '',
  percentage: 0,
  is_active: true,
};

/** صفحه مدیریت کدهای خاص */
export function SpecialCodesPage() {
  const [modalOpen, setModalOpen] = useState(false);
  const [editing, setEditing] = useState<SpecialCode | null>(null);

  const { data = [], isLoading, refetch } = useApiQuery({
    queryKey: ['special-codes'],
    queryFn: fetchSpecialCodes,
  });

  const { control, handleSubmit, reset, formState: { errors } } = useForm<FormValues>({
    resolver: zodResolver(schema),
    defaultValues: emptyValues,
  });

  useEffect(() => {
    if (!modalOpen) return;
    if (editing) {
      reset({
        code: editing.code,
        name: editing.name,
        description: editing.description,
        percentage: editing.percentage,
        is_active: editing.is_active,
      });
    } else {
      reset(emptyValues);
    }
  }, [editing, modalOpen, reset]);

  const createMutation = useApiMutation({
    mutationFn: createSpecialCode,
    successMessage: 'کد خاص ایجاد شد',
    onSuccess: () => {
      setModalOpen(false);
      refetch();
    },
  });

  const updateMutation = useApiMutation({
    mutationFn: (payload: SpecialCodePayload) => updateSpecialCode(editing!.id, payload),
    successMessage: 'کد خاص به‌روزرسانی شد',
    onSuccess: () => {
      setModalOpen(false);
      setEditing(null);
      refetch();
    },
  });

  const deleteMutation = useApiMutation({
    mutationFn: deleteSpecialCode,
    successMessage: 'کد خاص حذف شد',
    onSuccess: () => refetch(),
  });

  const columns: ColumnsType<SpecialCode> = [
    { title: 'کد', dataIndex: 'code', key: 'code', width: 120 },
    { title: 'نام', dataIndex: 'name', key: 'name' },
    { title: 'توضیحات', dataIndex: 'description', key: 'description', ellipsis: true },
    { title: 'درصد', dataIndex: 'percentage', key: 'percentage', width: 80 },
    {
      title: 'وضعیت',
      dataIndex: 'is_active',
      key: 'is_active',
      width: 90,
      render: (v: boolean) => <Tag color={v ? 'green' : 'default'}>{v ? 'فعال' : 'غیرفعال'}</Tag>,
    },
    {
      title: 'عملیات',
      key: 'actions',
      width: 200,
      render: (_, record) => (
        <>
          <PermissionGuard permission="special_code.update">
            <Button
              type="link"
              icon={<EditOutlined />}
              onClick={() => {
                setEditing(record);
                setModalOpen(true);
              }}
            >
              ویرایش
            </Button>
          </PermissionGuard>
          <PermissionGuard permission="special_code.delete">
            <Button
              type="link"
              danger
              icon={<DeleteOutlined />}
              onClick={() =>
                confirmDialog({
                  title: 'حذف کد خاص',
                  content: `کد «${record.code}» حذف شود؟`,
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
        title="کد خاص"
        extra={
          <PermissionGuard permission="special_code.create">
            <Button
              type="primary"
              icon={<PlusOutlined />}
              onClick={() => {
                setEditing(null);
                setModalOpen(true);
              }}
            >
              کد جدید
            </Button>
          </PermissionGuard>
        }
      />
      <DataTable columns={columns} data={data} loading={isLoading} rowKey="id" />
      <Modal
        title={editing ? 'ویرایش کد خاص' : 'ایجاد کد خاص'}
        open={modalOpen}
        onCancel={() => {
          setModalOpen(false);
          setEditing(null);
        }}
        footer={null}
        destroyOnHidden
      >
        <Form
          layout="vertical"
          onFinish={handleSubmit((values) => {
            const payload: SpecialCodePayload = {
              code: values.code,
              name: values.name,
              description: values.description ?? '',
              percentage: values.percentage,
              is_active: values.is_active,
            };
            if (editing) updateMutation.mutate(payload);
            else createMutation.mutate(payload);
          })}
        >
          <Form.Item label="کد" validateStatus={errors.code ? 'error' : ''} help={errors.code?.message}>
            <Controller name="code" control={control} render={({ field }) => <Input {...field} />} />
          </Form.Item>
          <Form.Item label="نام" validateStatus={errors.name ? 'error' : ''} help={errors.name?.message}>
            <Controller name="name" control={control} render={({ field }) => <Input {...field} />} />
          </Form.Item>
          <Form.Item label="توضیحات">
            <Controller name="description" control={control} render={({ field }) => <Input.TextArea {...field} rows={2} />} />
          </Form.Item>
          <Form.Item label="درصد" validateStatus={errors.percentage ? 'error' : ''} help={errors.percentage?.message}>
            <Controller
              name="percentage"
              control={control}
              render={({ field }) => (
                <InputNumber {...field} style={{ width: '100%' }} min={0} max={100} />
              )}
            />
          </Form.Item>
          <Form.Item label="فعال">
            <Controller
              name="is_active"
              control={control}
              render={({ field }) => <Switch checked={field.value} onChange={field.onChange} />}
            />
          </Form.Item>
          <Space>
            <Button type="primary" htmlType="submit" loading={createMutation.isPending || updateMutation.isPending}>
              ذخیره
            </Button>
            <Button onClick={() => setModalOpen(false)}>انصراف</Button>
          </Space>
        </Form>
      </Modal>
    </>
  );
}
