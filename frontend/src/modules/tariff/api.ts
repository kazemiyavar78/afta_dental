// این ماژول نمونه است؛ باقی فیلدها/صفحات طبق همین الگو در فازهای بعدی اضافه می‌شوند.

import { httpClient } from '@/platform/api/httpClient';
import type { CreateTariffPayload, Tariff } from './types';

export async function fetchTariffs(): Promise<Tariff[]> {
  const { data } = await httpClient.get<Tariff[]>('/tariff');
  return data;
}

export async function createTariff(payload: CreateTariffPayload): Promise<Tariff> {
  const { data } = await httpClient.post<Tariff>('/tariff', payload);
  return data;
}
