// این ماژول نمونه است؛ باقی فیلدها/صفحات طبق همین الگو در فازهای بعدی اضافه می‌شوند.

import { httpClient } from '@/platform/api/httpClient';
import type { CreateOrganizationPayload, Organization } from './types';

export async function fetchOrganizations(): Promise<Organization[]> {
  const { data } = await httpClient.get<Organization[]>('/organization');
  return data;
}

export async function createOrganization(payload: CreateOrganizationPayload): Promise<Organization> {
  const { data } = await httpClient.post<Organization>('/organization', payload);
  return data;
}
