import dayjs from 'dayjs';

/** نام فایل export را با تاریخ شمسی تولید می‌کند */
export function buildReportFilename(
  reportId: string,
  extension: 'pdf' | 'xlsx',
  title?: string,
): string {
  const date = dayjs().calendar('jalali').locale('fa').format('YYYY-MM-DD');
  const slug = title
    ? title.replace(/\s+/g, '-').replace(/[^\w\u0600-\u06FF-]/g, '')
    : reportId;
  return `${slug}-${date}.${extension}`;
}

/** نام sheet Excel (حداکثر ۳۱ کاراکتر) */
export function buildSheetName(title: string, maxLength = 31): string {
  const cleaned = title.replace(/[\\/*?:\[\]]/g, '-').trim();
  if (cleaned.length <= maxLength) return cleaned;
  return cleaned.slice(0, maxLength);
}
