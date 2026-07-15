import { httpClient } from '@/platform/api/httpClient';
import type { CreateRolePayload, Permission, RoleDetail, UpdateRolePayload } from './types';

/** دریافت لیست نقش‌ها */
export async function fetchRoles(): Promise<RoleDetail[]> {
  const { data } = await httpClient.get<RoleDetail[]>('/roles');
  return data;
}

/** دریافت یک نقش */
export async function fetchRole(id: number): Promise<RoleDetail> {
  const { data } = await httpClient.get<RoleDetail>(`/roles/${id}`);
  return data;
}

/** ایجاد نقش */
export async function createRole(payload: CreateRolePayload): Promise<RoleDetail> {
  const { data } = await httpClient.post<RoleDetail>('/roles', payload);
  return data;
}

/** به‌روزرسانی نقش */
export async function updateRole(id: number, payload: UpdateRolePayload): Promise<RoleDetail> {
  const { data } = await httpClient.put<RoleDetail>(`/roles/${id}`, payload);
  return data;
}

/** حذف نقش */
export async function deleteRole(id: number): Promise<void> {
  await httpClient.delete(`/roles/${id}`);
}

/** دریافت لیست مجوزهای سیستم */
export async function fetchPermissions(): Promise<Permission[]> {
  const { data } = await httpClient.get<Permission[]>('/permissions');
  return data;
}
