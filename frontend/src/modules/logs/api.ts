import { httpClient } from '@/platform/api/httpClient';
import type { LogFilters, SecurityLog } from './types';

/**
 * دریافت همه لاگ‌ها با فیلتر — فقط ادمین.
 */
export async function fetchLogs(filters: LogFilters = {}): Promise<SecurityLog[]> {
  const params: Record<string, string> = {};
  if (filters.ip) params.ip = filters.ip;
  if (filters.username) params.username = filters.username;
  if (filters.from_date) params.from_date = filters.from_date;
  if (filters.to_date) params.to_date = filters.to_date;

  const { data } = await httpClient.get<SecurityLog[]>('/logs', { params });
  return data;
}

/** دریافت لاگ‌های کاربر جاری */
export async function fetchMyLogs(): Promise<SecurityLog[]> {
  const { data } = await httpClient.get<SecurityLog[]>('/logs/user');
  return data;
}
