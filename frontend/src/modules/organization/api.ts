// این ماژول نمونه است؛ باقی فیلدها/صفحات طبق همین الگو در فازهای بعدی اضافه می‌شوند.

import { httpClient } from '@/platform/api/httpClient';
import type {
  CreateOrganizationPayload,
  ExcludedService,
  ExcludedServicesPayload,
  Organization,
  UpdateOrganizationPayload,
} from './types';

/** دریافت لیست سازمان‌ها */
export async function fetchOrganizations(): Promise<Organization[]> {
  const { data } = await httpClient.get<Organization[] | null>('/organization');
  return data ?? [];
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

/** دریافت خدمات خارج‌شده یک سازمان */
export async function fetchExcludedServices(organizationId: number): Promise<ExcludedService[]> {
  const { data } = await httpClient.get<ExcludedService[] | null>(
    `/excluded-services/organization/${organizationId}`,
  );
  return data ?? [];
}

/** افزودن خدمات خارج‌شده برای سازمان */
export async function addExcludedServices(
  payload: ExcludedServicesPayload,
): Promise<ExcludedService[]> {
  const { data } = await httpClient.post<ExcludedService[]>('/excluded-services', payload);
  return data;
}

/** حذف خدمات خارج‌شده سازمان */
export async function removeExcludedServices(payload: ExcludedServicesPayload): Promise<void> {
  await httpClient.delete('/excluded-services', { data: payload });
}
