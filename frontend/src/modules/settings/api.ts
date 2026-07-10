import { httpClient } from '@/platform/api/httpClient';
import type { SecuritySetting, UpdateSecuritySettingPayload } from './types';

export async function fetchSecuritySettings(): Promise<SecuritySetting[]> {
  const { data } = await httpClient.get<SecuritySetting[]>('/security/settings');
  return data;
}

export async function updateSecuritySetting(
  payload: UpdateSecuritySettingPayload,
): Promise<{ message: string }> {
  const { data } = await httpClient.put<{ message: string }>('/security/settings', payload);
  return data;
}