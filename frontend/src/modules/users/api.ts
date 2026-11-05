import { httpClient } from '@/platform/api/httpClient';
import { toRequestParams, type ListQuery, type PageResult } from '@/platform/list/query';
import type { CreateUserPayload, Role, Session, UpdateUserPayload, User } from './types';

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

/** دریافت لیست نقش‌های قابل انتصاب */
export async function fetchRoles(): Promise<Role[]> {
  const { data } = await httpClient.get<Role[]>('/roles/assignable');
  return data;
}

/**
 * فهرست صفحه‌بندی‌شده نشست‌ها را با فیلتر و مرتب‌سازی سمت سرور می‌گیرد.
 * @param query صفحه، اندازه، مرتب‌سازی و فیلترهای listquery
 * @returns صفحه استاندارد. شناسه عددی کاربر در جدول نشان داده نمی‌شود.
 */
export async function fetchSessionPage(query: ListQuery): Promise<PageResult<Session>> {
  const { data } = await httpClient.get<PageResult<Session>>('/sessions', {
    params: toRequestParams(query),
  });
  return data;
}

/**
 * بازه شناسه کاربر را فقط وقتی آن ستون برای نقش مجاز است به فیلتر اضافه می‌کند.
 * کاربر عادی اگر user_id بفرستد خطای اعتبارسنجی می‌گیرد، چون آن ستون در spec او نیست.
 * @param query وضعیت جدول
 * @param userId صاحب نشست‌هایی که این صفحه باید نشان بدهد
 * @param canFilterByUserId مدیر یا دارنده users.listSessions
 * @returns همان query، و برای نقش مجاز با user_id_from و user_id_to برابر
 */
export function scopeSessionQuery(query: ListQuery, userId: number, canFilterByUserId: boolean): ListQuery {
  if (!canFilterByUserId) return query;
  return {
    ...query,
    filters: {
      ...query.filters,
      user_id_from: String(userId),
      user_id_to: String(userId),
    },
  };
}

/** حذف یک نشست */
export async function deleteSession(sessionId: string): Promise<void> {
  await httpClient.delete(`/sessions/${sessionId}`);
}
