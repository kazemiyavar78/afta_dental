// این ماژول نمونه است؛ باقی فیلدها/صفحات طبق همین الگو در فازهای بعدی اضافه می‌شوند.

import { Form, Input, Button, Card, Space } from 'antd';
import { useForm, Controller } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { useNavigate } from 'react-router-dom';
import { PageHeader } from '@/platform/components/PageHeader';
import { useApiMutation } from '@/platform/hooks/useApiMutation';
import { fundSchema, type FundFormValues } from '../hooks';
import { createFund } from '../api';

export function FundFormPage() {
  const navigate = useNavigate();
  const { control, handleSubmit } = useForm<FundFormValues>({
    resolver: zodResolver(fundSchema),
    defaultValues: { name: '' },
  });

  const mutation = useApiMutation({
    mutationFn: createFund,
    successMessage: 'صندوق با موفقیت ایجاد شد',
    onSuccess: () => navigate('/fund'),
  });

  return (
    <>
      <PageHeader title="صندوق جدید" />
      <Card>
        <Form layout="vertical" onFinish={handleSubmit((v) => mutation.mutate(v))}>
          <Form.Item label="نام">
            <Controller name="name" control={control} render={({ field }) => <Input {...field} />} />
          </Form.Item>
          <Space>
            <Button type="primary" htmlType="submit" loading={mutation.isPending}>ذخیره</Button>
            <Button onClick={() => navigate('/fund')}>انصراف</Button>
          </Space>
        </Form>
      </Card>
    </>
  );
}
