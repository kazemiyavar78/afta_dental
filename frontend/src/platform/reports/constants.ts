import type { ReportMarginPreset, ReportPageSize } from './types';

/** مقادیر حاشیه بر اساس preset (میلی‌متر) */
export const REPORT_MARGIN_PRESETS: Record<Exclude<ReportMarginPreset, 'custom'>, number> = {
  small: 8,
  medium: 12,
  large: 18,
};

/** برچسب فارسی preset حاشیه */
export const REPORT_MARGIN_PRESET_LABELS: Record<ReportMarginPreset, string> = {
  small: 'کم (۸ میلی‌متر)',
  medium: 'متوسط (۱۲ میلی‌متر)',
  large: 'زیاد (۱۸ میلی‌متر)',
  custom: 'سفارشی',
};

/** برچسب فارسی اندازه صفحه */
export const REPORT_PAGE_SIZE_LABELS: Record<ReportPageSize, string> = {
  A4: 'A4',
  A5: 'A5',
  LETTER: 'Letter',
  LEGAL: 'Legal',
  CUSTOM: 'سفارشی',
};

/** پیکربندی پیش‌فرض صفحه A4 */
export const DEFAULT_REPORT_PAGE = {
  size: 'A4' as const,
  orientation: 'portrait' as const,
  margin: REPORT_MARGIN_PRESETS.medium,
};

/** URL فونت Vazirmatn برای PDF (local — public/fonts) */
export const PDF_FONT_VAZIRMATN_REGULAR = '/fonts/Vazirmatn-Regular.ttf';

export const PDF_FONT_VAZIRMATN_BOLD = '/fonts/Vazirmatn-Bold.ttf';

/** پیام‌های خطای export */
export const REPORT_EXPORT_ERRORS = {
  pdf: 'خطا در تولید فایل PDF',
  excel: 'خطا در ایجاد فایل Excel',
  print: 'خطا در چاپ گزارش',
  popup: 'اجازه باز شدن پنجره چاپ داده نشد',
} as const;
