// این ماژول نمونه است؛ باقی فیلدها/صفحات طبق همین الگو در فازهای بعدی اضافه می‌شوند.

import { Form, Input, Switch, Button, Card, Space, Spin } from 'antd';
import { useForm, Controller } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { useNavigate, useParams } from 'react-router-dom';
import { useEffect } from 'react';
import { PageHeader } from '@/platform/components/PageHeader';
import { useApiQuery } from '@/platform/hooks/useApiQuery';
import { useApiMutation } from '@/platform/hooks/useApiMutation';
import { organizationSchema, type OrganizationFormValues } from '../hooks';
import { createOrganization, fetchOrganization, updateOrganization } from '../api';

/** صفحه ایجاد/ویرایش سازمان */
export function OrganizationFormPage() {
  const navigate = useNavigate();
  const { id } = useParams<{ id: string }>();
  const isEdit = Boolean(id);
  const organizationId = Number(id);

  const { data: existing, isLoading: loadingExisting } = useApiQuery({
    queryKey: ['organizations', organizationId],
    queryFn: () => fetchOrganization(organizationId),
    enabled: isEdit,
  });

  const { control, handleSubmit, reset, setError, formState: { errors } } = useForm<OrganizationFormValues>({
    resolver: zodResolver(organizationSchema),
    defaultValues: {
      name: '',
      is_takmili: false,
      is_active: true,
    },
  });

  useEffect(() => {
    if (existing) {
      reset({
        name: existing.name,
        is_takmili: existing.is_takmili,
        is_active: existing.is_active,
      });
    }
  }, [existing, reset]);

  const createMutation = useApiMutation({
    mutationFn: createOrganization,
    successMessage: 'سازمان با موفقیت ایجاد شد',
    setError,
    onSuccess: () => navigate('/organization'),
  });

  const updateMutation = useApiMutation({
    mutationFn: (values: OrganizationFormValues) => updateOrganization(organizationId, values),
    successMessage: 'سازمان با موفقیت به‌روزرسانی شد',
    setError,
    onSuccess: () => navigate('/organization'),
  });

  if (isEdit && loadingExisting) {
    return <Spin size="large" />;
  }

  const isPending = createMutation.isPending || updateMutation.isPending;

  return (
    <>
      <PageHeader title={isEdit ? `ویرایش سازمان: ${existing?.name ?? ''}` : 'سازمان جدید'} />
      <Card>
        <Form
          layout="vertical"
          onFinish={handleSubmit((values) => {
            if (isEdit) {
              updateMutation.mutate(values);
            } else {
              createMutation.mutate(values);
            }
          })}
        >
          <Form.Item label="نام" validateStatus={errors.name ? 'error' : ''} help={errors.name?.message}>
            <Controller name="name" control={control} render={({ field }) => <Input {...field} />} />
          </Form.Item>
          <Form.Item label="بیمه تکمیلی">
            <Controller
              name="is_takmili"
              control={control}
              render={({ field }) => <Switch checked={field.value} onChange={field.onChange} />}
            />
          </Form.Item>
          <Form.Item label="فعال">
            <Controller
              name="is_active"
              control={control}
              render={({ field }) => <Switch checked={field.value} onChange={field.onChange} />}
            />
          </Form.Item>
          <Space>
            <Button type="primary" htmlType="submit" loading={isPending}>ذخیره</Button>
            <Button onClick={() => navigate('/organization')}>انصراف</Button>
          </Space>
        </Form>
      </Card>
    </>
  );
}
