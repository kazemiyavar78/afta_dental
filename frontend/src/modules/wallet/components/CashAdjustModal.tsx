import { useEffect } from 'react';
import { Button, Form, Input, InputNumber, Modal, Radio, Space, Typography } from 'antd';
import { useForm, Controller } from 'react-hook-form';
import { z } from 'zod';
import { zodResolver } from '@hookform/resolvers/zod';
import { useApiMutation } from '@/platform/hooks/useApiMutation';
import { useApiQuery } from '@/platform/hooks/useApiQuery';
import { adjustCash, fetchWalletBalance } from '@/modules/wallet/api';
import type { Patient } from '@/modules/patients/types';

const schema = z.object({
  amount: z.number().positive('مبلغ باید بزرگ‌تر از صفر باشد'),
  increase: z.boolean(),
  description: z.string().optional(),
});

type FormValues = z.infer<typeof schema>;

type CashAdjustModalProps = {
  open: boolean;
  patient: Patient | null;
  onClose: () => void;
};

/** مودال افزایش/کاهش اعتبار نقدی پرونده بیمار */
export function CashAdjustModal({ open, patient, onClose }: CashAdjustModalProps) {
  const { control, handleSubmit, reset, formState: { errors } } = useForm<FormValues>({
    resolver: zodResolver(schema),
    defaultValues: { amount: undefined as unknown as number, increase: true, description: '' },
  });

  const { data: balance, refetch } = useApiQuery({
    queryKey: ['wallet-balance', patient?.id],
    queryFn: () => fetchWalletBalance(patient!.id),
    enabled: open && !!patient,
  });

  useEffect(() => {
    if (open) {
      reset({ amount: undefined as unknown as number, increase: true, description: '' });
      if (patient) refetch();
    }
  }, [open, patient, reset, refetch]);

  const mutation = useApiMutation({
    mutationFn: adjustCash,
    successMessage: 'تراکنش نقدی ثبت شد',
    onSuccess: () => {
      onClose();
    },
  });

  if (!patient) return null;

  return (
    <Modal
      title={`اعتبار نقدی — ${patient.first_name} ${patient.last_name}`}
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
            amount: values.amount,
            increase: values.increase,
            description: values.description?.trim() || undefined,
          }),
        )}
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
