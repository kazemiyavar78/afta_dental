import { Form, Input, Select, Switch, Button, Card, Space, Spin } from 'antd';
import { useForm, Controller } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { useNavigate, useParams } from 'react-router-dom';
import { useEffect } from 'react';
import { PageHeader } from '@/platform/components/PageHeader';
import { useApiQuery } from '@/platform/hooks/useApiQuery';
import { useApiMutation } from '@/platform/hooks/useApiMutation';
import {
  createUserSchema,
  updateUserSchema,
  type CreateUserFormValues,
  type UpdateUserFormValues,
} from '../hooks';
import { createUser, fetchRoles, fetchUser, updateUser } from '../api';

/** صفحه ایجاد/ویرایش کاربر */
export function UserFormPage() {
  const navigate = useNavigate();
  const { id } = useParams<{ id: string }>();
  const isEdit = Boolean(id);
  const userId = Number(id);

  const { data: roles = [] } = useApiQuery({
    queryKey: ['roles'],
    queryFn: fetchRoles,
  });

  const { data: existingUser, isLoading: loadingUser } = useApiQuery({
    queryKey: ['users', userId],
    queryFn: () => fetchUser(userId),
    enabled: isEdit,
  });

  const createForm = useForm<CreateUserFormValues>({
    resolver: zodResolver(createUserSchema),
    defaultValues: {
      username: '',
      password: '',
      address: '',
      name: '',
      family: '',
      phone_number: '',
      medical_code: '',
      role_id: 0,
    },
  });

  const editForm = useForm<UpdateUserFormValues>({
    resolver: zodResolver(updateUserSchema),
    defaultValues: {
      address: '',
      name: '',
      family: '',
      phone_number: '',
      medical_code: '',
      role_id: 0,
      is_active: true,
    },
  });

  useEffect(() => {
    if (existingUser) {
      editForm.reset({
        address: existingUser.address,
        name: existingUser.name,
        family: existingUser.family,
        phone_number: existingUser.phone_number,
        medical_code: existingUser.medical_code ?? '',
        role_id: existingUser.role_id,
        is_active: existingUser.is_active,
      });
    }
  }, [existingUser, editForm]);

  const createMutation = useApiMutation({
    mutationFn: createUser,
    successMessage: 'کاربر با موفقیت ایجاد شد',
    setError: createForm.setError,
    onSuccess: () => navigate('/users'),
  });

  const updateMutation = useApiMutation({
    mutationFn: (values: UpdateUserFormValues) =>
      updateUser(userId, {
        ...values,
        medical_code: values.medical_code || null,
      }),
    successMessage: 'کاربر با موفقیت به‌روزرسانی شد',
    setError: editForm.setError,
    onSuccess: () => navigate('/users'),
  });

  if (isEdit && loadingUser) {
    return <Spin size="large" />;
  }

  const roleOptions = roles.map((r) => ({ label: r.name, value: r.id }));

  if (!isEdit) {
    const { control, handleSubmit, formState: { errors } } = createForm;
    return (
      <>
        <PageHeader title="ایجاد کاربر جدید" />
        <Card>
          <Form layout="vertical" onFinish={handleSubmit((v) => createMutation.mutate({
            username: v.username,
            password: v.password,
            name: v.name,
            family: v.family,
            address: v.address,
            phone_number: v.phone_number,
            medical_code: v.medical_code || null,
            role_id: v.role_id,
          }))}>
            <Form.Item label="نام کاربری" validateStatus={errors.username ? 'error' : ''} help={errors.username?.message}>
              <Controller name="username" control={control} render={({ field }) => <Input {...field} />} />
            </Form.Item>
            <Form.Item label="رمز عبور" validateStatus={errors.password ? 'error' : ''} help={errors.password?.message}>
              <Controller name="password" control={control} render={({ field }) => <Input.Password {...field} />} />
            </Form.Item>
            <Form.Item label="نام" validateStatus={errors.name ? 'error' : ''} help={errors.name?.message}>
              <Controller name="name" control={control} render={({ field }) => <Input {...field} />} />
            </Form.Item>
            <Form.Item label="نام خانوادگی" validateStatus={errors.family ? 'error' : ''} help={errors.family?.message}>
              <Controller name="family" control={control} render={({ field }) => <Input {...field} />} />
            </Form.Item>
            <Form.Item label="تلفن" validateStatus={errors.phone_number ? 'error' : ''} help={errors.phone_number?.message}>
              <Controller name="phone_number" control={control} render={({ field }) => <Input {...field} placeholder="09123456789" />} />
            </Form.Item>
            <Form.Item label="آدرس">
              <Controller name="address" control={control} render={({ field }) => <Input.TextArea {...field} />} />
            </Form.Item>
            <Form.Item label="کد نظام پزشکی">
              <Controller name="medical_code" control={control} render={({ field }) => <Input {...field} />} />
            </Form.Item>
            <Form.Item label="نقش" validateStatus={errors.role_id ? 'error' : ''} help={errors.role_id?.message}>
              <Controller name="role_id" control={control} render={({ field }) => (
                <Select {...field} options={roleOptions} placeholder="انتخاب نقش" />
              )} />
            </Form.Item>
            <Space>
              <Button type="primary" htmlType="submit" loading={createMutation.isPending}>ذخیره</Button>
              <Button onClick={() => navigate('/users')}>انصراف</Button>
            </Space>
          </Form>
        </Card>
      </>
    );
  }

  const { control, handleSubmit, formState: { errors } } = editForm;
  return (
    <>
      <PageHeader title={`ویرایش کاربر: ${existingUser?.username}`} />
      <Card>
        <Form layout="vertical" onFinish={handleSubmit((v) => updateMutation.mutate(v))}>
          <Form.Item label="نام" validateStatus={errors.name ? 'error' : ''} help={errors.name?.message}>
            <Controller name="name" control={control} render={({ field }) => <Input {...field} />} />
          </Form.Item>
          <Form.Item label="نام خانوادگی" validateStatus={errors.family ? 'error' : ''} help={errors.family?.message}>
            <Controller name="family" control={control} render={({ field }) => <Input {...field} />} />
          </Form.Item>
          <Form.Item label="تلفن" validateStatus={errors.phone_number ? 'error' : ''} help={errors.phone_number?.message}>
            <Controller name="phone_number" control={control} render={({ field }) => <Input {...field} />} />
          </Form.Item>
          <Form.Item label="آدرس">
            <Controller name="address" control={control} render={({ field }) => <Input.TextArea {...field} />} />
          </Form.Item>
          <Form.Item label="کد نظام پزشکی">
            <Controller name="medical_code" control={control} render={({ field }) => <Input {...field} />} />
          </Form.Item>
          <Form.Item label="نقش" validateStatus={errors.role_id ? 'error' : ''} help={errors.role_id?.message}>
            <Controller name="role_id" control={control} render={({ field }) => (
              <Select {...field} options={roleOptions} />
            )} />
          </Form.Item>
          <Form.Item label="فعال">
            <Controller name="is_active" control={control} render={({ field }) => (
              <Switch checked={field.value} onChange={field.onChange} />
            )} />
          </Form.Item>
          <Space>
            <Button type="primary" htmlType="submit" loading={updateMutation.isPending}>ذخیره</Button>
            <Button onClick={() => navigate('/users')}>انصراف</Button>
          </Space>
        </Form>
      </Card>
    </>
  );
}
