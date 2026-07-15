import { httpClient } from '@/platform/api/httpClient';
import type { OrganizationPackage, OrganizationPackagePayload } from './types';

/** دریافت لیست بسته‌های تعرفه */
export async function fetchOrganizationPackages(): Promise<OrganizationPackage[]> {
  const { data } = await httpClient.get<OrganizationPackage[] | null>('/organization-packages');
  return data ?? [];
}

/** ایجاد بسته تعرفه */
export async function createOrganizationPackage(
  payload: OrganizationPackagePayload,
): Promise<OrganizationPackage> {
  const { data } = await httpClient.post<OrganizationPackage>('/organization-packages', payload);
  return data;
}

/** به‌روزرسانی بسته تعرفه */
export async function updateOrganizationPackage(
  id: number,
  payload: OrganizationPackagePayload,
): Promise<OrganizationPackage> {
  const { data } = await httpClient.put<OrganizationPackage>(`/organization-packages/${id}`, payload);
  return data;
}

/** حذف بسته تعرفه */
export async function deleteOrganizationPackage(id: number): Promise<void> {
  await httpClient.delete(`/organization-packages/${id}`);
}
