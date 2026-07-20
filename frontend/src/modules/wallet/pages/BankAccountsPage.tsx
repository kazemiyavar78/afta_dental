import { useEffect, useState } from 'react';
import { Button, Card, Form, Input, Modal, Space } from 'antd';
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
  createBankAccount,
  deleteBankAccount,
  fetchBankAccounts,
  updateBankAccount,
} from '../api';
import type { BankAccount, BankAccountPayload } from '../types';

const schema = z.object({
  bank_name: z.string().min(1, 'نام بانک الزامی است'),
  sheba_number: z.string().min(1, 'شماره شبا الزامی است'),
  account_number: z.string().min(1, 'شماره حساب الزامی است'),
  card_number: z.string().min(1, 'شماره کارت الزامی است'),
  account_name: z.string().min(1, 'نام حساب الزامی است'),
});

type FormValues = z.infer<typeof schema>;

const emptyValues: FormValues = {
  bank_name: '',
  sheba_number: '',
  account_number: '',
  card_number: '',
  account_name: '',
};

/** صفحه تعریف حساب‌های بانکی دریافت/پرداخت */
export function BankAccountsPage() {
  const [modalOpen, setModalOpen] = useState(false);
  const [editing, setEditing] = useState<BankAccount | null>(null);

  const { data = [], isLoading, refetch } = useApiQuery({
    queryKey: ['bank-accounts'],
    queryFn: fetchBankAccounts,
  });

  const { control, handleSubmit, reset, formState: { errors } } = useForm<FormValues>({
    resolver: zodResolver(schema),
    defaultValues: emptyValues,
  });

  useEffect(() => {
    if (!modalOpen) return;
    if (editing) {
      reset({
        bank_name: editing.bank_name,
        sheba_number: editing.sheba_number,
        account_number: editing.account_number,
        card_number: editing.card_number,
        account_name: editing.account_name,
      });
    } else {
      reset(emptyValues);
    }
  }, [editing, modalOpen, reset]);

  const createMutation = useApiMutation({
    mutationFn: createBankAccount,
    successMessage: 'حساب بانکی ایجاد شد',
    onSuccess: () => {
      setModalOpen(false);
      refetch();
    },
  });

  const updateMutation = useApiMutation({
    mutationFn: (payload: BankAccountPayload) => updateBankAccount(editing!.id, payload),
    successMessage: 'حساب بانکی به‌روزرسانی شد',
    onSuccess: () => {
      setModalOpen(false);
      setEditing(null);
      refetch();
    },
  });

  const deleteMutation = useApiMutation({
    mutationFn: deleteBankAccount,
    successMessage: 'حساب بانکی حذف شد',
    onSuccess: () => refetch(),
  });

  const columns: ColumnsType<BankAccount> = [
    { title: 'نام بانک', dataIndex: 'bank_name', key: 'bank_name' },
    { title: 'نام حساب', dataIndex: 'account_name', key: 'account_name' },
    { title: 'شماره شبا', dataIndex: 'sheba_number', key: 'sheba_number' },
    { title: 'شماره حساب', dataIndex: 'account_number', key: 'account_number' },
    { title: 'شماره کارت', dataIndex: 'card_number', key: 'card_number' },
    {
      title: 'عملیات',
      key: 'actions',
      width: 220,
      render: (_, record) => (
        <>
          <PermissionGuard permission="bank_account.update">
            <Button
              type="link"
              icon={<EditOutlined />}
              disabled={record.has_transactions}
              onClick={() => {
                setEditing(record);
                setModalOpen(true);
              }}
            >
              ویرایش
            </Button>
          </PermissionGuard>
          <PermissionGuard permission="bank_account.delete">
            <Button
              type="link"
              danger
              icon={<DeleteOutlined />}
              disabled={record.has_transactions}
              loading={deleteMutation.isPending}
              onClick={() =>
                confirmDialog({
                  title: 'حذف حساب بانکی',
                  content: `آیا از حذف حساب «${record.account_name}» مطمئن هستید؟`,
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
        title="حساب‌های بانکی"
        extra={
          <PermissionGuard permission="bank_account.create">
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
        title={editing ? 'ویرایش حساب بانکی' : 'حساب بانکی جدید'}
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
            if (editing) updateMutation.mutate(values);
            else createMutation.mutate(values);
          })}
        >
          <Form.Item label="نام بانک" validateStatus={errors.bank_name ? 'error' : ''} help={errors.bank_name?.message}>
            <Controller name="bank_name" control={control} render={({ field }) => <Input {...field} />} />
          </Form.Item>
          <Form.Item label="نام حساب" validateStatus={errors.account_name ? 'error' : ''} help={errors.account_name?.message}>
            <Controller name="account_name" control={control} render={({ field }) => <Input {...field} />} />
          </Form.Item>
          <Form.Item label="شماره شبا" validateStatus={errors.sheba_number ? 'error' : ''} help={errors.sheba_number?.message}>
            <Controller name="sheba_number" control={control} render={({ field }) => <Input {...field} />} />
          </Form.Item>
          <Form.Item label="شماره حساب" validateStatus={errors.account_number ? 'error' : ''} help={errors.account_number?.message}>
            <Controller name="account_number" control={control} render={({ field }) => <Input {...field} />} />
          </Form.Item>
          <Form.Item label="شماره کارت" validateStatus={errors.card_number ? 'error' : ''} help={errors.card_number?.message}>
            <Controller name="card_number" control={control} render={({ field }) => <Input {...field} />} />
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
        حسابی که حداقل یک تراکنش داشته باشد قابل ویرایش یا حذف نیست.
      </Card>
    </>
  );
}
