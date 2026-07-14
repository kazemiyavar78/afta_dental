// این ماژول نمونه است؛ باقی فیلدها/صفحات طبق همین الگو در فازهای بعدی اضافه می‌شوند.

import { httpClient } from '@/platform/api/httpClient';
import type {
  CreateOrganizationPayload,
  Organization,
  UpdateOrganizationPayload,
} from './types';

/** دریافت لیست سازمان‌ها */
export async function fetchOrganizations(): Promise<Organization[]> {
  const { data } = await httpClient.get<Organization[]>('/organization');
  return data;
}

/** دریافت یک سازمان */
export async function fetchOrganization(id: number): Promise<Organization> {
  const { data } = await httpClient.get<Organization>(`/organization/${id}`);
  return data;
}

/** ایجاد سازمان */
export async function createOrganization(payload: CreateOrganizationPayload): Promise<Organization> {
  const { data } = await httpClient.post<Organization>('/organization', payload);
  return data;
}

/** به‌روزرسانی سازمان */
export async function updateOrganization(
  id: number,
  payload: UpdateOrganizationPayload,
): Promise<Organization> {
  const { data } = await httpClient.put<Organization>(`/organization/${id}`, payload);
  return data;
}

/** حذف سازمان */
export async function deleteOrganization(id: number): Promise<void> {
  await httpClient.delete(`/organization/${id}`);
}
