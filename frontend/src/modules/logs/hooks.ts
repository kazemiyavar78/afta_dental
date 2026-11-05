import { useApiQuery } from '@/platform/hooks/useApiQuery';
import { fetchLogIntegrity, fetchLogMeta } from './api';

/**
 * فراداده فیلتر لاگ را با React Query می‌خواند.
 * @returns نتیجه query برای نوع رویداد، دلیل و نوع هدف
 */
export function useLogMeta() {
  return useApiQuery({
    queryKey: ['logs', 'meta'],
    queryFn: fetchLogMeta,
  });
}

/**
 * وضعیت صحت زنجیره را با React Query می‌خواند.
 * @returns نتیجه query برای کارت بالای صفحه لاگ
 */
export function useLogIntegrity() {
  return useApiQuery({
    queryKey: ['logs', 'integrity'],
    queryFn: fetchLogIntegrity,
  });
}
