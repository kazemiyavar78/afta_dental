import { httpClient } from '@/platform/api/httpClient';
import type {
  CalculateTariffPayload,
  CalculateTariffResponse,
  RecalculateTariffPayload,
  SaveTariffResponse,
  Tariff,
} from './types';

/** تست محاسبه تعرفه برای سازمان (صندوق منفی مجاز است) */
export async function calculateTariff(
  payload: CalculateTariffPayload,
): Promise<CalculateTariffResponse> {
  const { data } = await httpClient.post<CalculateTariffResponse>('/tariff/calculate', payload);
  return data;
}

/** ذخیره/بروزرسانی تعرفه‌های محاسبه‌شده در یک تراکنش */
export async function saveTariffs(payload: CalculateTariffPayload): Promise<SaveTariffResponse> {
  const { data } = await httpClient.post<SaveTariffResponse>('/tariff/save', payload);
  return data;
}

/** دریافت لیست تعرفه‌های ذخیره‌شده یک سازمان */
export async function fetchTariffsByOrganization(organizationId: number): Promise<Tariff[]> {
  const { data } = await httpClient.get<Tariff[]>(`/tariff/organization/${organizationId}`);
  return data;
}

/** بازمحاسبه و ذخیره یک تعرفه خاص با سه مبلغ مرکز */
export async function recalculateTariff(
  id: number,
  payload: RecalculateTariffPayload,
): Promise<Tariff> {
  const { data } = await httpClient.put<Tariff>(`/tariff/${id}/recalculate`, payload);
  return data;
}
