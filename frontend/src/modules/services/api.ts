import { httpClient } from '@/platform/api/httpClient';
import type { ServiceItem, ServicePayload } from './types';

/** دریافت لیست خدمات */
export async function fetchServices(): Promise<ServiceItem[]> {
  const { data } = await httpClient.get<ServiceItem[]>('/services');
  return data;
}

/** دریافت یک خدمت */
export async function fetchService(id: number): Promise<ServiceItem> {
  const { data } = await httpClient.get<ServiceItem>(`/services/${id}`);
  return data;
}

/** ایجاد خدمت */
export async function createService(payload: ServicePayload): Promise<ServiceItem> {
  const { data } = await httpClient.post<ServiceItem>('/services', payload);
  return data;
}

/** به‌روزرسانی خدمت */
export async function updateService(id: number, payload: ServicePayload): Promise<ServiceItem> {
  const { data } = await httpClient.put<ServiceItem>(`/services/${id}`, payload);
  return data;
}

/** حذف خدمت */
export async function deleteService(id: number): Promise<void> {
  await httpClient.delete(`/services/${id}`);
}
