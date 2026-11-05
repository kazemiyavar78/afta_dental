import { useEffect, useState } from 'react';
import { Button, Card, Form, Input, InputNumber, Modal, Space, Switch, Tag } from 'antd';
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
  createCreditAccount,
  deleteCreditAccount,
  fetchCreditAccounts,
  updateCreditAccount,
} from '../api';
import type { CreditAccount, CreditAccountPayload } from '../types';

const schema = z.object({
  name: z.string().min(1, 'نام اعتبار الزامی است'),
  description: z.string(),
  credit_limit: z.number().positive('سقف اعتبار باید بزرگ‌تر از صفر باشد'),
  is_active: z.boolean(),
});

type FormValues = z.infer<typeof schema>;

const emptyValues: FormValues = {
  name: '',
  description: '',
  credit_limit: undefined as unknown as number,
  is_active: true,
};

/** مبلغ ریالی را با جداکننده فارسی نمایش می‌دهد. */
function formatRial(value: number): string {
  return `${value.toLocaleString('fa-IR')} ریال`;
}

/** صفحه تعریف و مدیریت حساب‌های اعتباری */
export function CreditAccountsPage() {
  const [modalOpen, setModalOpen] = useState(false);
  const [editing, setEditing] = useState<CreditAccount | null>(null);

  const { data = [], isLoading, refetch } = useApiQuery({
    queryKey: ['credit-accounts'],
    queryFn: fetchCreditAccounts,
  });

  const { control, handleSubmit, reset, formState: { errors } } = useForm<FormValues>({
    resolver: zodResolver(schema),
    defaultValues: emptyValues,
  });

  useEffect(() => {
    if (!modalOpen) return;
    if (editing) {
      reset({
        name: editing.name,
        description: editing.description,
        credit_limit: editing.credit_limit,
        is_active: editing.is_active,
      });
    } else {
      reset(emptyValues);
    }
  }, [editing, modalOpen, reset]);

  const createMutation = useApiMutation({
    mutationFn: createCreditAccount,
    successMessage: 'حساب اعتباری ایجاد شد',
    onSuccess: () => {
      setModalOpen(false);
      refetch();
    },
  });

  const updateMutation = useApiMutation({
    mutationFn: (payload: CreditAccountPayload) => updateCreditAccount(editing!.id, payload),
    successMessage: 'حساب اعتباری به‌روزرسانی شد',
    onSuccess: () => {
      setModalOpen(false);
      setEditing(null);
      refetch();
    },
  });

  const deleteMutation = useApiMutation({
    mutationFn: deleteCreditAccount,
    successMessage: 'حساب اعتباری حذف شد',
    onSuccess: () => refetch(),
  });

  const columns: ColumnsType<CreditAccount> = [
    { title: 'نام اعتبار', dataIndex: 'name', key: 'name' },
    { title: 'توضیحات', dataIndex: 'description', key: 'description', ellipsis: true },
    {
      title: 'سقف اعتبار',
      dataIndex: 'credit_limit',
      key: 'credit_limit',
      render: (value: number) => formatRial(value),
    },
    {
      title: 'استفاده‌شده',
      dataIndex: 'used_amount',
      key: 'used_amount',
      render: (value: number) => formatRial(value),
    },
    {
      title: 'مانده',
      dataIndex: 'remaining',
      key: 'remaining',
      render: (value: number) => formatRial(value),
    },
    {
      title: 'وضعیت',
      dataIndex: 'is_active',
      key: 'is_active',
      width: 110,
      render: (active: boolean) => (
        <Tag color={active ? 'green' : 'default'}>{active ? 'فعال' : 'غیرفعال'}</Tag>
      ),
    },
    {
      title: 'عملیات',
      key: 'actions',
      width: 220,
      render: (_, record) => (
        <>
          <PermissionGuard permission="credit_account.update">
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
          <PermissionGuard permission="credit_account.delete">
            <Button
              type="link"
              danger
              icon={<DeleteOutlined />}
              disabled={record.has_transactions}
              loading={deleteMutation.isPending}
              onClick={() =>
                confirmDialog({
                  title: 'حذف حساب اعتباری',
                  content: `آیا از حذف حساب «${record.name}» مطمئن هستید؟`,
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

  const isPending = createMutation.isPending || updateMutation.isPending;

  return (
    <>
      <PageHeader
        title="حساب‌های اعتباری"
        extra={
          <PermissionGuard permission="credit_account.create">
            <Button
              type="primary"
              icon={<PlusOutlined />}
              onClick={() => {
                setEditing(null);
                setModalOpen(true);
              }}
            >
              حساب جدید
            </Button>
          </PermissionGuard>
        }
      />
      <DataTable columns={columns} data={data} loading={isLoading} rowKey="id" />

      <Modal
        title={editing ? 'ویرایش حساب اعتباری' : 'حساب اعتباری جدید'}
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
            const payload: CreditAccountPayload = {
              name: values.name.trim(),
              description: values.description.trim(),
              credit_limit: values.credit_limit,
              is_active: values.is_active,
            };
            if (editing) updateMutation.mutate(payload);
            else createMutation.mutate(payload);
          })}
        >
          <Form.Item label="نام اعتبار" validateStatus={errors.name ? 'error' : ''} help={errors.name?.message}>
            <Controller name="name" control={control} render={({ field }) => <Input {...field} />} />
          </Form.Item>
          <Form.Item label="توضیحات" validateStatus={errors.description ? 'error' : ''} help={errors.description?.message}>
            <Controller
              name="description"
              control={control}
              render={({ field }) => <Input.TextArea {...field} rows={3} />}
            />
          </Form.Item>
          <Form.Item
            label="سقف اعتبار (ریال)"
            validateStatus={errors.credit_limit ? 'error' : ''}
            help={errors.credit_limit?.message ?? (editing?.has_transactions ? `حداقل سقف: ${formatRial(editing.used_amount)}` : undefined)}
          >
            <Controller
              name="credit_limit"
              control={control}
              render={({ field }) => (
                <InputNumber
                  style={{ width: '100%' }}
                  min={editing?.has_transactions ? editing.used_amount : 1}
                  value={field.value}
                  onChange={(v) => field.onChange(v ?? undefined)}
                />
              )}
            />
          </Form.Item>
          <Form.Item label="وضعیت">
            <Controller
              name="is_active"
              control={control}
              render={({ field }) => (
                <Switch
                  checked={field.value}
                  onChange={field.onChange}
                  checkedChildren="فعال"
                  unCheckedChildren="غیرفعال"
                />
              )}
            />
          </Form.Item>
          <Space>
            <Button type="primary" htmlType="submit" loading={isPending}>
              ذخیره
            </Button>
            <Button
              onClick={() => {
                setModalOpen(false);
                setEditing(null);
              }}
            >
              انصراف
            </Button>
          </Space>
        </Form>
      </Modal>
      <Card style={{ marginTop: 16 }} size="small">
        حسابی که پرداخت داشته باشد قابل حذف نیست. سقف اعتبار را نمی‌توان کمتر از مبلغ استفاده‌شده گذاشت. حساب غیرفعال در پرداخت جدید قابل انتخاب نیست.
      </Card>
    </>
  );
}
