import { httpClient } from '@/platform/api/httpClient';
import type { LoginFormValues, LoginResponse } from './types';

/** ورود کاربر */
export async function loginApi(data: LoginFormValues): Promise<LoginResponse> {
  const response = await httpClient.post<LoginResponse>('/login', {
    username: data.username,
    password: data.password,
  });
  return response.data;
}

/** خروج کاربر */
export async function logoutApi(): Promise<void> {
  await httpClient.post('/logout');
}
