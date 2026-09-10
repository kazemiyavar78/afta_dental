import type { ReportMetaItem } from '@/platform/reports';
import { formatGregorianDate } from '@/platform/reports';
import type { RevenueByOrganizationMeta } from '../types';

/** برچسب فیلتر سازمان را برای meta گزارش می‌سازد */
export function buildRevenueOrgFilterLabel(meta: RevenueByOrganizationMeta): string {
  if (meta.only_supplementary) {
    return `فقط سازمان تکمیلی — پایه: ${meta.free_organization_name ?? 'سازمان آزاد'}`;
  }
  if (meta.has_supplementary_filter) {
    return meta.detail_supplementary ? 'ریز بیمه‌های تکمیلی: بله' : 'ریز بیمه‌های تکمیلی: خیر';
  }
  return 'فقط پذیرش بدون بیمه تکمیلی';
}

/** meta rows برای header/export گزارش درآمد */
export function buildRevenueOrgMetaRows(meta: RevenueByOrganizationMeta): ReportMetaItem[][] {
  return [
    [
      { label: 'از تاریخ', value: `${formatGregorianDate(meta.from_date)} — ${meta.from_time}` },
      { label: 'تا تاریخ', value: `${formatGregorianDate(meta.to_date)} — ${meta.to_time}` },
    ],
    [{ label: 'فیلتر', value: buildRevenueOrgFilterLabel(meta) }],
  ];
}

/** re-export formatMoney برای backward compatibility */
export { formatCurrency as formatMoney } from '@/platform/reports';
