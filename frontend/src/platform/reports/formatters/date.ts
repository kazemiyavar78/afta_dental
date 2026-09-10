import dayjs from 'dayjs';

/** تاریخ میلادی YYYY-MM-DD را به شمسی تبدیل می‌کند */
export function formatGregorianDate(value: string | null | undefined): string {
  if (!value) return '—';
  const parsed = dayjs(`${value}T00:00:00`);
  if (!parsed.isValid()) return '—';
  return parsed.calendar('jalali').locale('fa').format('YYYY/MM/DD');
}

/** تاریخ ISO یا YYYY-MM-DD را برای نمایش UI format می‌کند */
export function formatDate(value: string | null | undefined): string {
  if (!value) return '—';
  const parsed = dayjs(value);
  if (!parsed.isValid()) return '—';
  return parsed.calendar('jalali').locale('fa').format('YYYY/MM/DD');
}

/** تاریخ و ساعت را به شمسی format می‌کند */
export function formatDateTime(value?: string | Date | null): string {
  if (!value) return '—';
  const parsed = dayjs(value);
  if (!parsed.isValid()) return '—';
  return parsed.calendar('jalali').locale('fa').format('YYYY/MM/DD HH:mm');
}

/** تاریخ/ساعت تولید گزارش */
export function formatGeneratedAt(): string {
  return formatDateTime(new Date());
}
