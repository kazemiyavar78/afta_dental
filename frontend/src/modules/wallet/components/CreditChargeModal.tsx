import { useEffect, useMemo } from 'react';
import { Button, Form, Input, InputNumber, Modal, Select, Space, Typography } from 'antd';
import { useForm, Controller } from 'react-hook-form';
import { z } from 'zod';
import { zodResolver } from '@hookform/resolvers/zod';
import { useApiMutation } from '@/platform/hooks/useApiMutation';
import { useApiQuery } from '@/platform/hooks/useApiQuery';
import { chargeFromCreditAccount, fetchActiveCreditAccounts, fetchWalletBalance } from '@/modules/wallet/api';
import type { Patient } from '@/modules/patients/types';

const schema = z.object({
  credit_account_id: z.number().positive('انتخاب حساب اعتباری الزامی است'),
  amount: z.number().positive('مبلغ باید بزرگ‌تر از صفر باشد'),
  description: z.string().optional(),
});

type FormValues = z.infer<typeof schema>;

type CreditChargeModalProps = {
  open: boolean;
  patient: Patient | null;
  onClose: () => void;
};

/** مودال شارژ پرونده مالی از محل کسر حساب اعتباری */
export function CreditChargeModal({ open, patient, onClose }: CreditChargeModalProps) {
  const { control, handleSubmit, reset, watch, formState: { errors } } = useForm<FormValues>({
    resolver: zodResolver(schema),
    defaultValues: {
      credit_account_id: undefined as unknown as number,
      amount: undefined as unknown as number,
      description: '',
    },
  });

  const selectedId = watch('credit_account_id');

  const { data: balance, refetch } = useApiQuery({
    queryKey: ['wallet-balance', patient?.id],
    queryFn: () => fetchWalletBalance(patient!.id),
    enabled: open && !!patient,
  });

  const { data: accounts = [], isLoading: accountsLoading } = useApiQuery({
    queryKey: ['credit-accounts', 'active'],
    queryFn: fetchActiveCreditAccounts,
    enabled: open,
  });

  const selected = useMemo(
    () => accounts.find((account) => account.id === selectedId),
    [accounts, selectedId],
  );

  useEffect(() => {
    if (open) {
      reset({
        credit_account_id: undefined as unknown as number,
        amount: undefined as unknown as number,
        description: '',
      });
      if (patient) refetch();
    }
  }, [open, patient, reset, refetch]);

  const mutation = useApiMutation({
    mutationFn: chargeFromCreditAccount,
    successMessage: 'شارژ از حساب اعتباری ثبت شد',
    onSuccess: () => {
      onClose();
    },
  });

  if (!patient) return null;

  return (
    <Modal
      title={`کسر از اعتبار — ${patient.first_name} ${patient.last_name}`}
      open={open}
      onCancel={onClose}
      footer={null}
      destroyOnHidden
    >
      <Typography.Paragraph type="secondary">
        شماره پرونده: {patient.file_number} — موجودی فعلی:{' '}
        {(balance?.balance ?? 0).toLocaleString('fa-IR')} ریال
      </Typography.Paragraph>
      <Form
        layout="vertical"
        onFinish={handleSubmit((values) =>
          mutation.mutate({
            file_id: patient.id,
            credit_account_id: values.credit_account_id,
            amount: values.amount,
            description: values.description?.trim() || undefined,
          }),
        )}
      >
        <Form.Item
          label="حساب اعتباری"
          validateStatus={errors.credit_account_id ? 'error' : ''}
          help={errors.credit_account_id?.message}
        >
          <Controller
            name="credit_account_id"
            control={control}
            render={({ field }) => (
              <Select
                {...field}
                value={field.value || undefined}
                loading={accountsLoading}
                placeholder="انتخاب حساب اعتباری"
                options={accounts.map((account) => ({
                  value: account.id,
                  disabled: account.remaining <= 0,
                  label: `${account.name} — مانده ${account.remaining.toLocaleString('fa-IR')} ریال`,
                }))}
                onChange={(value) => field.onChange(value)}
              />
            )}
          />
        </Form.Item>
        {selected && (
          <Typography.Paragraph type="secondary">
            سقف: {selected.credit_limit.toLocaleString('fa-IR')} ریال — مانده:{' '}
            {selected.remaining.toLocaleString('fa-IR')} ریال
            {selected.description ? ` — ${selected.description}` : ''}
          </Typography.Paragraph>
        )}
        <Form.Item label="مبلغ (ریال)" validateStatus={errors.amount ? 'error' : ''} help={errors.amount?.message}>
          <Controller
            name="amount"
            control={control}
            render={({ field }) => (
              <InputNumber
                style={{ width: '100%' }}
                min={1}
                max={selected?.remaining || undefined}
                value={field.value}
                onChange={(v) => field.onChange(v ?? undefined)}
              />
            )}
          />
        </Form.Item>
        <Form.Item label="توضیحات">
          <Controller name="description" control={control} render={({ field }) => <Input.TextArea {...field} rows={2} />} />
        </Form.Item>
        <Space>
          <Button type="primary" htmlType="submit" loading={mutation.isPending} disabled={accounts.length === 0}>
            ثبت شارژ
          </Button>
          <Button onClick={onClose}>انصراف</Button>
        </Space>
      </Form>
    </Modal>
  );
}
