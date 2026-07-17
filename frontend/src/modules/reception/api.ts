import { httpClient } from '@/platform/api/httpClient';
import type {
  CalculatePayload,
  CalculatedServiceLine,
  DoctorUser,
  ReceptionDetail,
  UpsertReceptionPayload,
} from './types';

/** دریافت لیست پذیرش‌ها */
export async function fetchReceptions(): Promise<ReceptionDetail[]> {
  const { data } = await httpClient.get<ReceptionDetail[]>('/reception');
  return data;
}

/** دریافت یک پذیرش */
export async function fetchReception(id: number): Promise<ReceptionDetail> {
  const { data } = await httpClient.get<ReceptionDetail>(`/reception/${id}`);
  return data;
}

/** ناوبری بین پذیرش‌ها */
export async function navigateReception(
  dir: 'first' | 'prev' | 'next' | 'last',
  cursor?: number | null,
): Promise<ReceptionDetail> {
  const { data } = await httpClient.get<ReceptionDetail>('/reception/nav', {
    params: { dir, cursor: cursor ?? undefined },
  });
  return data;
}

/** ایجاد پذیرش */
export async function createReception(payload: UpsertReceptionPayload): Promise<ReceptionDetail> {
  const { data } = await httpClient.post<ReceptionDetail>('/reception', payload);
  return data;
}

/** ویرایش پذیرش */
export async function updateReception(
  id: number,
  payload: UpsertReceptionPayload,
): Promise<ReceptionDetail> {
  const { data } = await httpClient.put<ReceptionDetail>(`/reception/${id}`, payload);
  return data;
}

/** حذف نرم پذیرش */
export async function deleteReception(id: number): Promise<void> {
  await httpClient.delete(`/reception/${id}`);
}

/** بازیابی پذیرش حذف‌شده */
export async function restoreReception(id: number): Promise<ReceptionDetail> {
  const { data } = await httpClient.post<ReceptionDetail>(`/reception/${id}/restore`);
  return data;
}

/** محاسبه زنده خدمات */
export async function calculateReceptionServices(
  payload: CalculatePayload,
): Promise<CalculatedServiceLine[]> {
  const { data } = await httpClient.post<{ services: CalculatedServiceLine[] }>(
    '/reception/calculate',
    payload,
  );
  return data.services;
}

/** لیست پزشکان */
export async function fetchDoctors(): Promise<DoctorUser[]> {
  const { data } = await httpClient.get<DoctorUser[] | null>('/users/doctors');
  return data ?? [];
}

/** لیست دستیاران */
export async function fetchAssistants(): Promise<DoctorUser[]> {
  const { data } = await httpClient.get<DoctorUser[] | null>('/users/assistants');
  return data ?? [];
}
