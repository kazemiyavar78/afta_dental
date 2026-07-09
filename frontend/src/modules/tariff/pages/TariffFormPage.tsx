// این ماژول نمونه است؛ باقی فیلدها/صفحات طبق همین الگو در فازهای بعدی اضافه می‌شوند.

import { Form, Input, InputNumber, Button, Card, Space } from 'antd';
import { useForm, Controller } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { useNavigate } from 'react-router-dom';
import { PageHeader } from '@/platform/components/PageHeader';
import { useApiMutation } from '@/platform/hooks/useApiMutation';
import { tariffSchema, type TariffFormValues } from '../hooks';
import { createTariff } from '../api';

export function TariffFormPage() {
  const navigate = useNavigate();
  const { control, handleSubmit } = useForm<TariffFormValues>({
    resolver: zodResolver(tariffSchema),
    defaultValues: { name: '', amount: 0 },
  });

  const mutation = useApiMutation({
    mutationFn: createTariff,
    successMessage: 'تعرفه با موفقیت ایجاد شد',
    onSuccess: () => navigate('/tariff'),
  });

  return (
    <>
      <PageHeader title="تعرفه جدید" />
      <Card>
        <Form layout="vertical" onFinish={handleSubmit((v) => mutation.mutate(v))}>
          <Form.Item label="نام">
            <Controller name="name" control={control} render={({ field }) => <Input {...field} />} />
          </Form.Item>
          <Form.Item label="مبلغ">
            <Controller name="amount" control={control} render={({ field }) => <InputNumber {...field} style={{ width: '100%' }} />} />
          </Form.Item>
          <Space>
            <Button type="primary" htmlType="submit" loading={mutation.isPending}>ذخیره</Button>
            <Button onClick={() => navigate('/tariff')}>انصراف</Button>
          </Space>
        </Form>
      </Card>
    </>
  );
}
