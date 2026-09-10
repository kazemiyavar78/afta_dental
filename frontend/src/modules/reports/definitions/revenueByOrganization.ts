import type { ReportDefinition } from '@/platform/reports';
import { formatCurrency } from '@/platform/reports';
import type { RevenueByOrganizationRow } from '../types';

/** ستون‌های مشترک گزارش درآمد / کارکرد سازمان */
export const organizationReportColumns: ReportDefinition<RevenueByOrganizationRow>['columns'] = [
  {
    key: 'row_label',
    title: 'سازمان',
    width: 22,
    align: 'right',
    format: 'text',
  },
  {
    key: 'service_amount',
    title: 'نرخ خدمات',
    width: 11,
    format: 'currency',
    formatter: (v) => formatCurrency(v),
  },
  {
    key: 'service_tariff',
    title: 'تعرفه خدمات',
    width: 11,
    format: 'currency',
    formatter: (v) => formatCurrency(v),
  },
  {
    key: 'organization_share',
    title: 'سهم سازمان',
    width: 11,
    format: 'currency',
    formatter: (v) => formatCurrency(v),
  },
  {
    key: 'supplementary_share',
    title: 'سهم بیمه تکمیلی',
    width: 11,
    format: 'currency',
    formatter: (v) => formatCurrency(v),
  },
  {
    key: 'subsidy_share',
    title: 'سهم یارانه',
    width: 10,
    format: 'currency',
    formatter: (v) => formatCurrency(v),
  },
  {
    key: 'payable_amount',
    title: 'قابل پرداخت',
    width: 11,
    format: 'currency',
    formatter: (v) => formatCurrency(v),
  },
  {
    key: 'reception_count',
    title: 'تعداد پذیرش',
    width: 9,
    format: 'number',
  },
];

/** خلاصه مالی مشترک */
export const organizationReportSummary: ReportDefinition<RevenueByOrganizationRow>['summary'] = {
  label: 'جمع کل',
  fields: [
    { key: 'service_amount', label: 'جمع نرخ', format: 'currency' },
    { key: 'service_tariff', label: 'جمع تعرفه', format: 'currency' },
    { key: 'organization_share', label: 'جمع سهم سازمان', format: 'currency' },
    { key: 'supplementary_share', label: 'جمع تکمیلی', format: 'currency' },
    { key: 'subsidy_share', label: 'جمع یارانه', format: 'currency' },
    { key: 'payable_amount', label: 'جمع قابل پرداخت', format: 'currency' },
  ],
};

/** ReportDefinition گزارش درآمد به تفکیک سازمان */
export const revenueByOrganizationDefinition: ReportDefinition<RevenueByOrganizationRow> = {
  id: 'revenue-by-organization',
  title: 'گزارش درآمد به تفکیک سازمان',
  description: 'تجمیع درآمد پذیرش‌ها بر اساس سازمان پایه و تکمیلی',
  page: { size: 'A4', orientation: 'portrait', margin: 12 },
  columns: organizationReportColumns,
  summary: organizationReportSummary,
  header: { showGeneratedAt: true },
  footer: { text: 'صفحات به‌صورت خودکار بر اساس اندازه A4 شکسته می‌شوند.' },
  toolbar: { print: true, pdf: true, excel: true, pageSetup: true },
  export: { sheetName: 'درآمد سازمان' },
};
