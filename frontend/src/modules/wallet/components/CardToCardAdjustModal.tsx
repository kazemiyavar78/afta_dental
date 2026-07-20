import { useEffect } from 'react';
import { Button, Form, Input, InputNumber, Modal, Radio, Select, Space, Typography } from 'antd';
import { useForm, Controller } from 'react-hook-form';
import { z } from 'zod';
import { zodResolver } from '@hookform/resolvers/zod';
import { useApiMutation } from '@/platform/hooks/useApiMutation';
import { useApiQuery } from '@/platform/hooks/useApiQuery';
import { adjustCardToCard, fetchBankAccounts, fetchWalletBalance } from '@/modules/wallet/api';
import type { Patient } from '@/modules/patients/types';

const schema = z.object({
  amount: z.number().positive('مبلغ باید بزرگ‌تر از صفر باشد'),
  increase: z.boolean(),
  bank_account_id: z.number().positive('انتخاب حساب بانکی الزامی است'),
  counterparty_card: z.string().min(4, 'شماره کارت الزامی است'),
  tracking_number: z.string().min(1, 'شماره پیگیری الزامی است'),
  paid_at: z.string().optional(),
  description: z.string().optional(),
});

type FormValues = z.infer<typeof schema>;

type CardToCardAdjustModalProps = {
  open: boolean;
  patient: Patient | null;
  onClose: () => void;
};

/** مودال افزایش/کاهش اعتبار کارت‌به‌کارت */
export function CardToCardAdjustModal({ open, patient, onClose }: CardToCardAdjustModalProps) {
  const { control, handleSubmit, reset, formState: { errors } } = useForm<FormValues>({
    resolver: zodResolver(schema),
    defaultValues: {
      amount: undefined as unknown as number,
      increase: true,
      bank_account_id: undefined as unknown as number,
      counterparty_card: '',
      tracking_number: '',
      paid_at: '',
      description: '',
    },
  });

  const { data: balance, refetch: refetchBalance } = useApiQuery({
    queryKey: ['wallet-balance', patient?.id],
    queryFn: () => fetchWalletBalance(patient!.id),
    enabled: open && !!patient,
  });

  const { data: accounts = [] } = useApiQuery({
    queryKey: ['bank-accounts'],
    queryFn: fetchBankAccounts,
    enabled: open,
  });

  useEffect(() => {
    if (open) {
      reset({
        amount: undefined as unknown as number,
        increase: true,
        bank_account_id: undefined as unknown as number,
        counterparty_card: '',
        tracking_number: '',
        paid_at: '',
        description: '',
      });
      if (patient) refetchBalance();
    }
  }, [open, patient, reset, refetchBalance]);

  const mutation = useApiMutation({
    mutationFn: adjustCardToCard,
    successMessage: 'تراکنش کارت‌به‌کارت ثبت شد',
    onSuccess: () => onClose(),
  });

  if (!patient) return null;

  return (
    <Modal
      title={`کارت‌به‌کارت — ${patient.first_name} ${patient.last_name}`}
      open={open}
      onCancel={onClose}
      footer={null}
      destroyOnHidden
      width={560}
    >
      <Typography.Paragraph type="secondary">
        شماره پرونده: {patient.file_number} — موجودی فعلی:{' '}
        {(balance?.balance ?? 0).toLocaleString('fa-IR')} ریال
      </Typography.Paragraph>
      <Form
        layout="vertical"
        onFinish={handleSubmit((values) => {
          let paidAt: string | null = null;
          if (values.paid_at?.trim()) {
            const d = new Date(values.paid_at);
            paidAt = Number.isNaN(d.getTime()) ? values.paid_at.trim() : d.toISOString();
          }
          mutation.mutate({
            file_id: patient.id,
            amount: values.amount,
            increase: values.increase,
            bank_account_id: values.bank_account_id,
            counterparty_card: values.counterparty_card.trim(),
            tracking_number: values.tracking_number.trim(),
            paid_at: paidAt,
            description: values.description?.trim() || undefined,
          });
        })}
      >
        <Form.Item label="نوع عملیات">
          <Controller
            name="increase"
            control={control}
            render={({ field }) => (
              <Radio.Group
                value={field.value}
                onChange={(e) => field.onChange(e.target.value)}
                options={[
                  { label: 'افزایش اعتبار', value: true },
                  { label: 'کاهش اعتبار', value: false },
                ]}
              />
            )}
          />
        </Form.Item>
        <Form.Item
          label="حساب بانکی کلینیک"
          validateStatus={errors.bank_account_id ? 'error' : ''}
          help={errors.bank_account_id?.message}
        >
          <Controller
            name="bank_account_id"
            control={control}
            render={({ field }) => (
              <Select
                style={{ width: '100%' }}
                value={field.value}
                onChange={field.onChange}
                placeholder="انتخاب حساب"
                options={accounts.map((a) => ({
                  value: a.id,
                  label: `${a.bank_name} — ${a.account_name} (${a.card_number})`,
                }))}
              />
            )}
          />
        </Form.Item>
        <Form.Item
          label="شماره کارت طرف مقابل"
          validateStatus={errors.counterparty_card ? 'error' : ''}
          help={errors.counterparty_card?.message}
        >
          <Controller name="counterparty_card" control={control} render={({ field }) => <Input {...field} />} />
        </Form.Item>
        <Form.Item
          label="شماره تراکنش / پیگیری"
          validateStatus={errors.tracking_number ? 'error' : ''}
          help={errors.tracking_number?.message}
        >
          <Controller name="tracking_number" control={control} render={({ field }) => <Input {...field} />} />
        </Form.Item>
        <Form.Item label="زمان پرداخت (اختیاری — در صورت خالی بودن زمان ثبت)">
          <Controller
            name="paid_at"
            control={control}
            render={({ field }) => <Input {...field} type="datetime-local" />}
          />
        </Form.Item>
        <Form.Item label="مبلغ (ریال)" validateStatus={errors.amount ? 'error' : ''} help={errors.amount?.message}>
          <Controller
            name="amount"
            control={control}
            render={({ field }) => (
              <InputNumber
                style={{ width: '100%' }}
                min={1}
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
          <Button type="primary" htmlType="submit" loading={mutation.isPending}>
            ثبت
          </Button>
          <Button onClick={onClose}>انصراف</Button>
        </Space>
      </Form>
    </Modal>
  );
}
