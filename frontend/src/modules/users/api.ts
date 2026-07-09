import { httpClient } from '@/platform/api/httpClient';
import type { CreateUserPayload, Role, UpdateUserPayload, User } from './types';

/** دریافت لیست کاربران */
export async function fetchUsers(): Promise<User[]> {
  const { data } = await httpClient.get<User[]>('/users');
  return data;
}

/** دریافت یک کاربر */
export async function fetchUser(id: number): Promise<User> {
  const { data } = await httpClient.get<User>(`/users/${id}`);
  return data;
}

/** ایجاد کاربر */
export async function createUser(payload: CreateUserPayload): Promise<User> {
  const { data } = await httpClient.post<User>('/users', payload);
  return data;
}

/** به‌روزرسانی کاربر */
export async function updateUser(id: number, payload: UpdateUserPayload): Promise<User> {
  const { data } = await httpClient.put<User>(`/users/${id}`, payload);
  return data;
}

/** دریافت لیست نقش‌ها */
export async function fetchRoles(): Promise<Role[]> {
  const { data } = await httpClient.get<Role[]>('/roles');
  return data;
}
