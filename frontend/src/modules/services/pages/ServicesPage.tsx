import { useEffect } from 'react';
import {
  Button,
  Form,
  Input,
  InputNumber,
  Modal,
  Select,
  Space,
  Switch,
  Tag,
} from 'antd';
import {
  DeleteOutlined,
  EditOutlined,
  PlusOutlined,
} from '@ant-design/icons';
import { useForm, Controller } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import type { ColumnsType } from 'antd/es/table';
import { PageHeader } from '@/platform/components/PageHeader';
import { DataTable } from '@/platform/components/DataTable';
import { PermissionGuard } from '@/platform/auth/PermissionGuard';
import { confirmDialog } from '@/platform/components/ConfirmDialog';
import { useApiQuery } from '@/platform/hooks/useApiQuery';
import { useApiMutation } from '@/platform/hooks/useApiMutation';
import { useState } from 'react';
import { createService, deleteService, fetchServices, updateService } from '../api';
import { serviceSchema, type ServiceFormValues } from '../hooks';
import type { ServiceItem } from '../types';

const featureOptions = [
  { label: 'بدون ویژگی', value: '' },
  { label: '#', value: '#' },
  { label: '*', value: '*' },
  { label: '#*', value: '#*' },
];

const emptyFormValues: ServiceFormValues = {
  service_code: '',
  name: '',
  technical_coefficient: 0,
  professional_coefficient: 0,
  consumption_coefficient: 0,
  service_rate: 0,
  service_tariff: 0,
  international_code: '',
  default_count: 0,
  maximum_count: 0,
  service_features: '',
  is_active: true,
  is_dental_direction: false,
  allow_multiple_use: false,
};

/** صفحه مدیریت خدمات (لیست، ایجاد، ویرایش و حذف در یک صفحه) */
export function ServicesPage() {
  const [modalOpen, setModalOpen] = useState(false);
  const [editing, setEditing] = useState<ServiceItem | null>(null);

  const { data = [], isLoading, refetch } = useApiQuery({
    queryKey: ['services'],
    queryFn: fetchServices,
  });

  const {
    control,
    handleSubmit,
    reset,
    setError,
    formState: { errors },
  } = useForm<ServiceFormValues>({
    resolver: zodResolver(serviceSchema),
    defaultValues: emptyFormValues,
  });

  useEffect(() => {
    if (!modalOpen) return;
    if (editing) {
      reset({
        service_code: editing.service_code,
        name: editing.name,
        technical_coefficient: editing.technical_coefficient,
        professional_coefficient: editing.professional_coefficient,
        consumption_coefficient: editing.consumption_coefficient,
        service_rate: editing.service_rate,
        service_tariff: editing.service_tariff,
        international_code: editing.international_code,
        default_count: editing.default_count,
        maximum_count: editing.maximum_count,
        service_features: editing.service_features,
        is_active: editing.is_active,
        is_dental_direction: editing.is_dental_direction,
        allow_multiple_use: editing.allow_multiple_use,
      });
    } else {
      reset(emptyFormValues);
    }
  }, [editing, modalOpen, reset]);

  const createMutation = useApiMutation({
    mutationFn: createService,
    successMessage: 'خدمت با موفقیت ایجاد شد',
    setError,
    onSuccess: () => {
      setModalOpen(false);
      refetch();
    },
  });

  const updateMutation = useApiMutation({
    mutationFn: (values: ServiceFormValues) => updateService(editing!.id, values),
    successMessage: 'خدمت با موفقیت به‌روزرسانی شد',
    setError,
    onSuccess: () => {
      setModalOpen(false);
      setEditing(null);
      refetch();
    },
  });

  const deleteMutation = useApiMutation({
    mutationFn: deleteService,
    successMessage: 'خدمت با موفقیت حذف شد',
    onSuccess: () => refetch(),
  });

  const openCreate = () => {
    setEditing(null);
    setModalOpen(true);
  };

  const openEdit = (record: ServiceItem) => {
    setEditing(record);
    setModalOpen(true);
  };

  const closeModal = () => {
    setModalOpen(false);
    setEditing(null);
  };

  const isPending = createMutation.isPending || updateMutation.isPending;

  const columns: ColumnsType<ServiceItem> = [
    { title: 'کد', dataIndex: 'service_code', key: 'service_code', width: 100 },
    { title: 'نام', dataIndex: 'name', key: 'name' },
    { title: 'نرخ', dataIndex: 'service_rate', key: 'service_rate', width: 110 },
    { title: 'تعرفه', dataIndex: 'service_tariff', key: 'service_tariff', width: 110 },
    {
      title: 'ویژگی',
      dataIndex: 'service_features',
      key: 'service_features',
      width: 90,
      render: (value: string) => value || '—',
    },
    {
      title: 'وضعیت',
      key: 'status',
      width: 90,
      render: (_, record) => (
        <Tag color={record.is_active ? 'green' : 'red'}>
          {record.is_active ? 'فعال' : 'غیرفعال'}
        </Tag>
      ),
    },
    {
      title: 'عملیات',
      key: 'actions',
      width: 200,
      render: (_, record) => (
        <>
          <PermissionGuard permission="services.update">
            <Button type="link" icon={<EditOutlined />} onClick={() => openEdit(record)}>
              ویرایش
            </Button>
          </PermissionGuard>
          <PermissionGuard permission="services.delete">
            <Button
              type="link"
              danger
              icon={<DeleteOutlined />}
              loading={deleteMutation.isPending}
              onClick={() =>
                confirmDialog({
                  title: 'حذف خدمت',
                  content: `آیا از حذف خدمت «${record.name}» مطمئن هستید؟`,
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
        title="خدمات"
        extra={
          <PermissionGuard permission="services.create">
            <Button type="primary" icon={<PlusOutlined />} onClick={openCreate}>
              خدمت جدید
            </Button>
          </PermissionGuard>
        }
      />
      <DataTable columns={columns} data={data} loading={isLoading} rowKey="id" />

      <Modal
        title={editing ? `ویرایش خدمت: ${editing.name}` : 'خدمت جدید'}
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
            label="کد خدمت"
            validateStatus={errors.service_code ? 'error' : ''}
            help={errors.service_code?.message}
          >
            <Controller name="service_code" control={control} render={({ field }) => <Input {...field} />} />
          </Form.Item>
          <Form.Item
            label="نام خدمت"
            validateStatus={errors.name ? 'error' : ''}
            help={errors.name?.message}
          >
            <Controller name="name" control={control} render={({ field }) => <Input {...field} />} />
          </Form.Item>
          <Space wrap size="large" style={{ display: 'flex', marginBottom: 8 }}>
            <Form.Item label="ضریب فنی" validateStatus={errors.technical_coefficient ? 'error' : ''} help={errors.technical_coefficient?.message}>
              <Controller
                name="technical_coefficient"
                control={control}
                render={({ field }) => (
                  <InputNumber
                    min={0}
                    style={{ width: 140 }}
                    value={field.value}
                    onChange={(v) => field.onChange(v ?? 0)}
                  />
                )}
              />
            </Form.Item>
            <Form.Item label="ضریب حرفه‌ای" validateStatus={errors.professional_coefficient ? 'error' : ''} help={errors.professional_coefficient?.message}>
              <Controller
                name="professional_coefficient"
                control={control}
                render={({ field }) => (
                  <InputNumber
                    min={0}
                    style={{ width: 140 }}
                    value={field.value}
                    onChange={(v) => field.onChange(v ?? 0)}
                  />
                )}
              />
            </Form.Item>
            <Form.Item label="ضریب مصرفی" validateStatus={errors.consumption_coefficient ? 'error' : ''} help={errors.consumption_coefficient?.message}>
              <Controller
                name="consumption_coefficient"
                control={control}
                render={({ field }) => (
                  <InputNumber
                    min={0}
                    style={{ width: 140 }}
                    value={field.value}
                    onChange={(v) => field.onChange(v ?? 0)}
                  />
                )}
              />
            </Form.Item>
          </Space>
          <Space wrap size="large" style={{ display: 'flex', marginBottom: 8 }}>
            <Form.Item label="نرخ خدمت" validateStatus={errors.service_rate ? 'error' : ''} help={errors.service_rate?.message}>
              <Controller
                name="service_rate"
                control={control}
                render={({ field }) => (
                  <InputNumber
                    min={0}
                    style={{ width: 160 }}
                    value={field.value}
                    onChange={(v) => field.onChange(v ?? 0)}
                  />
                )}
              />
            </Form.Item>
            <Form.Item label="تعرفه خدمت" validateStatus={errors.service_tariff ? 'error' : ''} help={errors.service_tariff?.message}>
              <Controller
                name="service_tariff"
                control={control}
                render={({ field }) => (
                  <InputNumber
                    min={0}
                    style={{ width: 160 }}
                    value={field.value}
                    onChange={(v) => field.onChange(v ?? 0)}
                  />
                )}
              />
            </Form.Item>
          </Space>
          <Form.Item label="کد بین‌المللی" validateStatus={errors.international_code ? 'error' : ''} help={errors.international_code?.message}>
            <Controller name="international_code" control={control} render={({ field }) => <Input {...field} />} />
          </Form.Item>
          <Space wrap size="large" style={{ display: 'flex', marginBottom: 8 }}>
            <Form.Item label="تعداد پیش‌فرض" validateStatus={errors.default_count ? 'error' : ''} help={errors.default_count?.message}>
              <Controller
                name="default_count"
                control={control}
                render={({ field }) => (
                  <InputNumber
                    min={0}
                    style={{ width: 140 }}
                    value={field.value}
                    onChange={(v) => field.onChange(v ?? 0)}
                  />
                )}
              />
            </Form.Item>
            <Form.Item label="حداکثر تعداد" validateStatus={errors.maximum_count ? 'error' : ''} help={errors.maximum_count?.message}>
              <Controller
                name="maximum_count"
                control={control}
                render={({ field }) => (
                  <InputNumber
                    min={0}
                    style={{ width: 140 }}
                    value={field.value}
                    onChange={(v) => field.onChange(v ?? 0)}
                  />
                )}
              />
            </Form.Item>
            <Form.Item label="ویژگی خدمت" validateStatus={errors.service_features ? 'error' : ''} help={errors.service_features?.message}>
              <Controller
                name="service_features"
                control={control}
                render={({ field }) => (
                  <Select {...field} options={featureOptions} style={{ width: 160 }} />
                )}
              />
            </Form.Item>
          </Space>
          <Form.Item label="فعال">
            <Controller
              name="is_active"
              control={control}
              render={({ field }) => <Switch checked={field.value} onChange={field.onChange} />}
            />
          </Form.Item>
          <Form.Item label="جهت دندان دارد">
            <Controller
              name="is_dental_direction"
              control={control}
              render={({ field }) => <Switch checked={field.value} onChange={field.onChange} />}
            />
          </Form.Item>
          <Form.Item label="اجازه استفاده بیش از یکبار در پرونده">
            <Controller
              name="allow_multiple_use"
              control={control}
              render={({ field }) => <Switch checked={field.value} onChange={field.onChange} />}
            />
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
