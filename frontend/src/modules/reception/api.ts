// این ماژول نمونه است؛ باقی فیلدها/صفحات طبق همین الگو در فازهای بعدی اضافه می‌شوند.

import { httpClient } from '@/platform/api/httpClient';
import type { CreateReceptionPayload, Reception } from './types';

export async function fetchReceptions(): Promise<Reception[]> {
  const { data } = await httpClient.get<Reception[]>('/reception');
  return data;
}

export async function createReception(payload: CreateReceptionPayload): Promise<Reception> {
  const { data } = await httpClient.post<Reception>('/reception', payload);
  return data;
}
