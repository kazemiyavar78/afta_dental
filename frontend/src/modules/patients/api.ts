import { httpClient } from '@/platform/api/httpClient';
import type { Patient, PatientPayload, PatientSearchParams } from './types';

/** پارامترهای خالی را از شیء جستجو حذف می‌کند. */
function cleanSearchParams(params?: PatientSearchParams): PatientSearchParams | undefined {
  if (!params) return undefined;
  const cleaned: Record<string, string | boolean> = {};
  for (const [key, value] of Object.entries(params)) {
    if (value === undefined || value === null || value === '') continue;
    cleaned[key] = value as string | boolean;
  }
  return Object.keys(cleaned).length > 0 ? (cleaned as PatientSearchParams) : undefined;
}

/** دریافت لیست بیماران (با فیلتر اختیاری) */
export async function fetchPatients(params?: PatientSearchParams): Promise<Patient[]> {
  const { data } = await httpClient.get<Patient[]>('/patients', {
    params: cleanSearchParams(params),
  });
  return data;
}

/** دریافت یک بیمار */
export async function fetchPatient(id: number): Promise<Patient> {
  const { data } = await httpClient.get<Patient>(`/patients/${id}`);
  return data;
}

/** ایجاد بیمار */
export async function createPatient(payload: PatientPayload): Promise<Patient> {
  const { data } = await httpClient.post<Patient>('/patients', payload);
  return data;
}

/** به‌روزرسانی بیمار */
export async function updatePatient(id: number, payload: PatientPayload): Promise<Patient> {
  const { data } = await httpClient.put<Patient>(`/patients/${id}`, payload);
  return data;
}

/** حذف بیمار */
export async function deletePatient(id: number): Promise<void> {
  await httpClient.delete(`/patients/${id}`);
}
