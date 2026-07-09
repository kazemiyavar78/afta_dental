// این ماژول نمونه است؛ باقی فیلدها/صفحات طبق همین الگو در فازهای بعدی اضافه می‌شوند.

import { Form, Input, Button, Card, Space } from 'antd';
import { useForm, Controller } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { useNavigate } from 'react-router-dom';
import { PageHeader } from '@/platform/components/PageHeader';
import { useApiMutation } from '@/platform/hooks/useApiMutation';
import { organizationSchema, type OrganizationFormValues } from '../hooks';
import { createOrganization } from '../api';

export function OrganizationFormPage() {
  const navigate = useNavigate();
  const { control, handleSubmit, formState: { errors } } = useForm<OrganizationFormValues>({
    resolver: zodResolver(organizationSchema),
    defaultValues: { name: '' },
  });

  const mutation = useApiMutation({
    mutationFn: createOrganization,
    successMessage: 'سازمان با موفقیت ایجاد شد',
    onSuccess: () => navigate('/organization'),
  });

  return (
    <>
      <PageHeader title="سازمان جدید" />
      <Card>
        <Form layout="vertical" onFinish={handleSubmit((v) => mutation.mutate(v))}>
          <Form.Item label="نام" validateStatus={errors.name ? 'error' : ''} help={errors.name?.message}>
            <Controller name="name" control={control} render={({ field }) => <Input {...field} />} />
          </Form.Item>
          <Space>
            <Button type="primary" htmlType="submit" loading={mutation.isPending}>ذخیره</Button>
            <Button onClick={() => navigate('/organization')}>انصراف</Button>
          </Space>
        </Form>
      </Card>
    </>
  );
}
