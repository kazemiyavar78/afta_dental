import { httpClient } from '@/platform/api/httpClient';
import type { SpecialCode, SpecialCodePayload } from './types';

/** لیست کدهای خاص را دریافت می‌کند. */
export async function fetchSpecialCodes(): Promise<SpecialCode[]> {
  const { data } = await httpClient.get<SpecialCode[]>('/special-codes');
  return data;
}

/** کد خاص را با کد جستجو می‌کند. */
export async function fetchSpecialCodeByCode(code: string): Promise<SpecialCode> {
  const { data } = await httpClient.get<SpecialCode>(`/special-codes/by-code/${encodeURIComponent(code)}`);
  return data;
}

/** کد خاص جدید ایجاد می‌کند. */
export async function createSpecialCode(payload: SpecialCodePayload): Promise<SpecialCode> {
  const { data } = await httpClient.post<SpecialCode>('/special-codes', payload);
  return data;
}

/** کد خاص را بروزرسانی می‌کند. */
export async function updateSpecialCode(id: number, payload: SpecialCodePayload): Promise<SpecialCode> {
  const { data } = await httpClient.put<SpecialCode>(`/special-codes/${id}`, payload);
  return data;
}

/** کد خاص را حذف می‌کند. */
export async function deleteSpecialCode(id: number): Promise<void> {
  await httpClient.delete(`/special-codes/${id}`);
}
