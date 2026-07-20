import { httpClient } from '@/platform/api/httpClient';
import type { Regulation, RegulationPayload } from './types';

/** لیست ضوابط را دریافت می‌کند. */
export async function fetchRegulations(): Promise<Regulation[]> {
  const { data } = await httpClient.get<Regulation[]>('/regulations');
  return data;
}

/** ضابطه جدید ایجاد می‌کند. */
export async function createRegulation(payload: RegulationPayload): Promise<Regulation> {
  const { data } = await httpClient.post<Regulation>('/regulations', payload);
  return data;
}

/** ضابطه را بروزرسانی می‌کند. */
export async function updateRegulation(id: number, payload: RegulationPayload): Promise<Regulation> {
  const { data } = await httpClient.put<Regulation>(`/regulations/${id}`, payload);
  return data;
}

/** ضابطه را حذف می‌کند. */
export async function deleteRegulation(id: number): Promise<void> {
  await httpClient.delete(`/regulations/${id}`);
}
