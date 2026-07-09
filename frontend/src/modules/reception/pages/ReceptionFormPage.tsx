// این ماژول نمونه است؛ باقی فیلدها/صفحات طبق همین الگو در فازهای بعدی اضافه می‌شوند.

import { Form, Input, InputNumber, Button, Card, Space } from 'antd';
import { useForm, Controller } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { useNavigate } from 'react-router-dom';
import { PageHeader } from '@/platform/components/PageHeader';
import { useApiMutation } from '@/platform/hooks/useApiMutation';
import { receptionSchema, type ReceptionFormValues } from '../hooks';
import { createReception } from '../api';

export function ReceptionFormPage() {
  const navigate = useNavigate();
  const { control, handleSubmit, formState: { errors } } = useForm<ReceptionFormValues>({
    resolver: zodResolver(receptionSchema),
    defaultValues: { patient_name: '', doctor_id: 1, reception_date: '' },
  });

  const mutation = useApiMutation({
    mutationFn: createReception,
    successMessage: 'پذیرش با موفقیت ثبت شد',
    onSuccess: () => navigate('/reception'),
  });

  return (
    <>
      <PageHeader title="پذیرش جدید" />
      <Card>
        <Form layout="vertical" onFinish={handleSubmit((v) => mutation.mutate(v))}>
          <Form.Item label="نام بیمار" validateStatus={errors.patient_name ? 'error' : ''} help={errors.patient_name?.message}>
            <Controller name="patient_name" control={control} render={({ field }) => <Input {...field} />} />
          </Form.Item>
          <Form.Item label="شناسه پزشک" validateStatus={errors.doctor_id ? 'error' : ''} help={errors.doctor_id?.message}>
            <Controller name="doctor_id" control={control} render={({ field }) => <InputNumber {...field} style={{ width: '100%' }} />} />
          </Form.Item>
          <Form.Item label="تاریخ (YYYY-MM-DD)" validateStatus={errors.reception_date ? 'error' : ''} help={errors.reception_date?.message}>
            <Controller name="reception_date" control={control} render={({ field }) => <Input {...field} placeholder="2026-07-07" />} />
          </Form.Item>
          <Space>
            <Button type="primary" htmlType="submit" loading={mutation.isPending}>ذخیره</Button>
            <Button onClick={() => navigate('/reception')}>انصراف</Button>
          </Space>
        </Form>
      </Card>
    </>
  );
}
