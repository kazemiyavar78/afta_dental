import { useEffect, useState } from 'react';
import { Button, Form, Input, InputNumber, Modal, Select, Space, Switch, Tag } from 'antd';
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
import { fetchServices } from '@/modules/services/api';
import {
  createRegulation,
  deleteRegulation,
  fetchRegulations,
  updateRegulation,
} from '../api';
import type { Regulation, RegulationPayload } from '../types';

const schema = z.object({
  service_ids: z.array(z.number()).min(1, 'حداقل یک خدمت الزامی است'),
  duration_days: z.number().min(1, 'مدت زمان باید بزرگ‌تر از صفر باشد'),
  is_active: z.boolean(),
  photo_count: z.number().min(0),
  description: z.string(),
});

type FormValues = z.infer<typeof schema>;

const emptyValues: FormValues = {
  service_ids: [],
  duration_days: 30,
  is_active: true,
  photo_count: 1,
  description: '',
};

/** صفحه مدیریت ضوابط خدمات */
export function RegulationsPage() {
  const [modalOpen, setModalOpen] = useState(false);
  const [editing, setEditing] = useState<Regulation | null>(null);

  const { data = [], isLoading, refetch } = useApiQuery({
    queryKey: ['regulations'],
    queryFn: fetchRegulations,
  });

  const { data: services = [] } = useApiQuery({
    queryKey: ['services'],
    queryFn: fetchServices,
  });

  const { control, handleSubmit, reset, formState: { errors } } = useForm<FormValues>({
    resolver: zodResolver(schema),
    defaultValues: emptyValues,
  });

  useEffect(() => {
    if (!modalOpen) return;
    if (editing) {
      reset({
        service_ids: editing.service_ids,
        duration_days: editing.duration_days,
        is_active: editing.is_active,
        photo_count: editing.photo_count,
        description: editing.description,
      });
    } else {
      reset(emptyValues);
    }
  }, [editing, modalOpen, reset]);

  const createMutation = useApiMutation({
    mutationFn: createRegulation,
    successMessage: 'ضابطه ایجاد شد',
    onSuccess: () => {
      setModalOpen(false);
      refetch();
    },
  });

  const updateMutation = useApiMutation({
    mutationFn: (payload: RegulationPayload) => updateRegulation(editing!.id, payload),
    successMessage: 'ضابطه به‌روزرسانی شد',
    onSuccess: () => {
      setModalOpen(false);
      setEditing(null);
      refetch();
    },
  });

  const deleteMutation = useApiMutation({
    mutationFn: deleteRegulation,
    successMessage: 'ضابطه حذف شد',
    onSuccess: () => refetch(),
  });

  const serviceName = (id: number) => services.find((s) => s.id === id)?.name ?? String(id);

  const columns: ColumnsType<Regulation> = [
    {
      title: 'خدمات',
      dataIndex: 'service_ids',
      key: 'service_ids',
      render: (ids: number[]) => (
        <div style={{ display: 'flex', flexWrap: 'wrap', gap: 4 }}>
          {ids.map((id) => (
            <Tag key={id}>{serviceName(id)}</Tag>
          ))}
        </div>
      ),
    },
    { title: 'مدت (روز)', dataIndex: 'duration_days', key: 'duration_days', width: 100 },
    { title: 'تعداد عکس', dataIndex: 'photo_count', key: 'photo_count', width: 100 },
    { title: 'توضیحات', dataIndex: 'description', key: 'description', ellipsis: true },
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
          <PermissionGuard permission="regulation.update">
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
          <PermissionGuard permission="regulation.delete">
            <Button
              type="link"
              danger
              icon={<DeleteOutlined />}
              onClick={() =>
                confirmDialog({
                  title: 'حذف ضابطه',
                  content: 'این ضابطه حذف شود؟',
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
        title="ضوابط"
        extra={
          <PermissionGuard permission="regulation.create">
            <Button
              type="primary"
              icon={<PlusOutlined />}
              onClick={() => {
                setEditing(null);
                setModalOpen(true);
              }}
            >
              ضابطه جدید
            </Button>
          </PermissionGuard>
        }
      />
      <DataTable columns={columns} data={data} loading={isLoading} rowKey="id" />
      <Modal
        title={editing ? 'ویرایش ضابطه' : 'ایجاد ضابطه'}
        open={modalOpen}
        onCancel={() => {
          setModalOpen(false);
          setEditing(null);
        }}
        footer={null}
        destroyOnHidden
        width={640}
      >
        <Form
          layout="vertical"
          onFinish={handleSubmit((values) => {
            const payload: RegulationPayload = {
              service_ids: values.service_ids,
              duration_days: values.duration_days,
              is_active: values.is_active,
              photo_count: values.photo_count,
              description: values.description ?? '',
            };
            if (editing) updateMutation.mutate(payload);
            else createMutation.mutate(payload);
          })}
        >
          <Form.Item
            label="لیست خدمات"
            validateStatus={errors.service_ids ? 'error' : ''}
            help={errors.service_ids?.message}
          >
            <Controller
              name="service_ids"
              control={control}
              render={({ field }) => (
                <Select
                  mode="multiple"
                  allowClear
                  showSearch
                  optionFilterProp="label"
                  style={{ width: '100%' }}
                  value={field.value}
                  onChange={field.onChange}
                  options={services.map((s) => ({
                    value: s.id,
                    label: `${s.service_code} — ${s.name}`,
                  }))}
                />
              )}
            />
          </Form.Item>
          <Form.Item
            label="مدت زمان (روز)"
            validateStatus={errors.duration_days ? 'error' : ''}
            help={errors.duration_days?.message}
          >
            <Controller
              name="duration_days"
              control={control}
              render={({ field }) => <InputNumber {...field} style={{ width: '100%' }} min={1} />}
            />
          </Form.Item>
          <Form.Item label="تعداد عکس">
            <Controller
              name="photo_count"
              control={control}
              render={({ field }) => <InputNumber {...field} style={{ width: '100%' }} min={0} />}
            />
          </Form.Item>
          <Form.Item label="توضیحات">
            <Controller
              name="description"
              control={control}
              render={({ field }) => <Input.TextArea {...field} rows={2} />}
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
