import { Card, Form, Input, Button, Alert } from 'antd';
import { useForm, Controller } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { useNavigate, useLocation } from 'react-router-dom';
import { loginSchema, type LoginFormValues } from '../types';
import { loginApi } from '../api';
import { useAuth } from '@/platform/auth/useAuth';
import { useApiMutation } from '@/platform/hooks/useApiMutation';

/** صفحه ورود */
export function LoginPage() {
  const navigate = useNavigate();
  const location = useLocation();
  const { refresh, isAuthenticated } = useAuth();
  const from = (location.state as { from?: { pathname: string } })?.from?.pathname ?? '/users';

  const {
    control,
    handleSubmit,
    formState: { errors },
  } = useForm<LoginFormValues>({
    resolver: zodResolver(loginSchema),
    defaultValues: { username: '', password: '' },
  });

  const mutation = useApiMutation({
    mutationFn: loginApi,
    onSuccess: async () => {
      await refresh();
      navigate(from, { replace: true });
    },
  });

  if (isAuthenticated) {
    navigate(from, { replace: true });
    return null;
  }

  return (
    <Card style={{ width: 400 }}>
      <Form layout="vertical" onFinish={handleSubmit((v) => mutation.mutate(v))}>
        <Form.Item
          label="نام کاربری"
          validateStatus={errors.username ? 'error' : ''}
          help={errors.username?.message}
        >
          <Controller
            name="username"
            control={control}
            render={({ field }) => <Input {...field} autoComplete="username" />}
          />
        </Form.Item>

        <Form.Item
          label="رمز عبور"
          validateStatus={errors.password ? 'error' : ''}
          help={errors.password?.message}
        >
          <Controller
            name="password"
            control={control}
            render={({ field }) => (
              <Input.Password {...field} autoComplete="current-password" />
            )}
          />
        </Form.Item>

        {mutation.isError && (
          <Alert type="error" message={mutation.error.message} style={{ marginBottom: 16 }} />
        )}

        <Button type="primary" htmlType="submit" block loading={mutation.isPending}>
          ورود
        </Button>
      </Form>
    </Card>
  );
}
