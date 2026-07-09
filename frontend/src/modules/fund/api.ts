// این ماژول نمونه است؛ باقی فیلدها/صفحات طبق همین الگو در فازهای بعدی اضافه می‌شوند.

import { httpClient } from '@/platform/api/httpClient';
import type { CreateFundPayload, Fund } from './types';

export async function fetchFunds(): Promise<Fund[]> {
  const { data } = await httpClient.get<Fund[]>('/fund');
  return data;
}

export async function createFund(payload: CreateFundPayload): Promise<Fund> {
  const { data } = await httpClient.post<Fund>('/fund', payload);
  return data;
}
