import { httpClient } from '@/platform/api/httpClient';
import { toRequestParams, type ListQuery, type PageResult } from '@/platform/list/query';
import type { LogIntegrity, LogMeta, SecurityLog } from './types';

/**
 * یک صفحه از لاگ امنیتی را با فیلتر و مرتب‌سازی سمت سرور می‌گیرد.
 * @param query پارامترهای جدول
 * @returns صفحه استاندارد listquery
 */
export async function fetchLogs(query: ListQuery): Promise<PageResult<SecurityLog>> {
  const { data } = await httpClient.get<PageResult<SecurityLog>>('/logs', {
    params: toRequestParams(query),
  });
  return data;
}

/**
 * برچسب فارسی نوع رویداد، دلیل و نوع هدف را برای فیلتر کشویی می‌گیرد.
 * @returns فراداده فیلتر
 */
export async function fetchLogMeta(): Promise<LogMeta> {
  const { data } = await httpClient.get<LogMeta>('/logs/meta');
  return data;
}

/**
 * نتیجه آخرین بررسی زنجیره و یافته‌های باز را می‌گیرد. هش در پاسخ نیست.
 * @returns زمان بررسی، نتیجه، شناسه اولین ردیف خراب، تعداد و فهرست یافته‌ها
 */
export async function fetchLogIntegrity(): Promise<LogIntegrity> {
  const { data } = await httpClient.get<LogIntegrity>('/logs/integrity');
  return data;
}
