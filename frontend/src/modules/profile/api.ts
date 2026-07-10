import { httpClient } from '@/platform/api/httpClient';
import type { ChangePasswordPayload, UserProfileResponse } from './types';

/** دریافت اطلاعات پروفایل کاربر */
export async function fetchUserProfile(): Promise<UserProfileResponse> {
  const { data } = await httpClient.get<UserProfileResponse>('/profile');
  return data;
}

/** تغییر رمز عبور کاربر فعلی */
export async function changePassword(payload: ChangePasswordPayload): Promise<void> {
  await httpClient.post('/change-password', payload);
}

/** حذف یک نشست فعال */
export async function deleteSession(sessionId: string): Promise<void> {
  await httpClient.delete(`/sessions/${sessionId}`);
}
