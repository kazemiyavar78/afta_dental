import type { ReportDefinition, ReportMetaItem } from '@/platform/reports';
import { formatCurrency, formatGregorianDate } from '@/platform/reports';
import type {
  UserPerformanceMeta,
  UserPerformanceReceptionDetailRow,
  UserPerformanceReceptionRow,
  UserPerformanceReportType,
  UserPerformanceTransactionDetailRow,
  UserPerformanceTransactionRow,
} from '../types';

/** ستون ردیف با شماره‌گذاری خودکار */
const rowNumberColumn = {
  key: 'row_number' as const,
  title: 'ردیف',
  width: 5,
  format: 'number' as const,
  formatter: (_v: unknown, _r: unknown, index: number) => String(index + 1),
};

/** ReportDefinition — کارکرد بر اساس پذیرش (تجمیع کاربر) */
export const userPerformanceByReceptionDefinition: ReportDefinition<UserPerformanceReceptionRow> = {
  id: 'user-performance-by-reception',
  title: 'گزارش کارکرد کاربران — بر اساس پذیرش',
  page: { size: 'A4', orientation: 'portrait', margin: 12 },
  columns: [
    rowNumberColumn,
    {
      key: 'user_name',
      title: 'نام و نام خانوادگی کاربر',
      width: 18,
      align: 'right',
      format: 'text',
    },
    {
      key: 'service_amount',
      title: 'نرخ خدمات',
      width: 10,
      format: 'currency',
      formatter: (v) => formatCurrency(v),
    },
    {
      key: 'service_tariff',
      title: 'تعرفه خدمات',
      width: 10,
      format: 'currency',
      formatter: (v) => formatCurrency(v),
    },
    {
      key: 'organization_share',
      title: 'سهم سازمان',
      width: 10,
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
      width: 9,
      format: 'currency',
      formatter: (v) => formatCurrency(v),
    },
    {
      key: 'payable_amount',
      title: 'قابل پرداخت',
      width: 10,
      format: 'currency',
      formatter: (v) => formatCurrency(v),
    },
    {
      key: 'reception_count',
      title: 'تعداد پذیرش',
      width: 9,
      format: 'number',
    },
  ],
  summary: {
    label: 'جمع کل',
    fields: [
      { key: 'service_amount', label: 'جمع نرخ', format: 'currency' },
      { key: 'service_tariff', label: 'جمع تعرفه', format: 'currency' },
      { key: 'organization_share', label: 'جمع سهم سازمان', format: 'currency' },
      { key: 'supplementary_share', label: 'جمع تکمیلی', format: 'currency' },
      { key: 'subsidy_share', label: 'جمع یارانه', format: 'currency' },
      { key: 'payable_amount', label: 'جمع قابل پرداخت', format: 'currency' },
      { key: 'reception_count', label: 'تعداد پذیرش', format: 'number' },
    ],
  },
  header: { showGeneratedAt: true },
  footer: { text: 'صفحات به‌صورت خودکار بر اساس اندازه A4 شکسته می‌شوند.' },
  toolbar: { print: true, pdf: true, excel: true, pageSetup: true },
  export: { sheetName: 'کارکرد پذیرش', includeSections: true },
};

/** ReportDefinition — کارکرد بر اساس پذیرش (تفکیک بیمار) */
export const userPerformanceReceptionDetailDefinition: ReportDefinition<UserPerformanceReceptionDetailRow> =
  {
    id: 'user-performance-reception-detail',
    title: 'گزارش کارکرد کاربران — ریز پذیرش‌ها',
    page: { size: 'A4', orientation: 'portrait', margin: 12 },
    columns: [
      rowNumberColumn,
      { key: 'reception_id', title: 'شماره پذیرش', width: 9, format: 'number' },
      {
        key: 'patient_name',
        title: 'نام و نام خانوادگی بیمار',
        width: 14,
        align: 'right',
        format: 'text',
      },
      { key: 'national_code', title: 'کد ملی بیمار', width: 10, format: 'text' },
      {
        key: 'service_amount',
        title: 'نرخ خدمت',
        width: 9,
        format: 'currency',
        formatter: (v) => formatCurrency(v),
      },
      {
        key: 'service_tariff',
        title: 'تعرفه خدمت',
        width: 9,
        format: 'currency',
        formatter: (v) => formatCurrency(v),
      },
      {
        key: 'organization_share',
        title: 'سهم سازمان',
        width: 9,
        format: 'currency',
        formatter: (v) => formatCurrency(v),
      },
      {
        key: 'supplementary_share',
        title: 'سهم بیمه تکمیلی',
        width: 10,
        format: 'currency',
        formatter: (v) => formatCurrency(v),
      },
      {
        key: 'subsidy_share',
        title: 'سهم یارانه',
        width: 8,
        format: 'currency',
        formatter: (v) => formatCurrency(v),
      },
      {
        key: 'payable_amount',
        title: 'قابل پرداخت',
        width: 9,
        format: 'currency',
        formatter: (v) => formatCurrency(v),
      },
      {
        key: 'services_received',
        title: 'خدمات پذیرش شده',
        width: 20,
        align: 'right',
        format: 'text',
      },
    ],
    summary: {
      label: 'جمع',
      fields: [
        { key: 'service_amount', label: 'جمع نرخ', format: 'currency' },
        { key: 'service_tariff', label: 'جمع تعرفه', format: 'currency' },
        { key: 'organization_share', label: 'جمع سهم سازمان', format: 'currency' },
        { key: 'supplementary_share', label: 'جمع تکمیلی', format: 'currency' },
        { key: 'subsidy_share', label: 'جمع یارانه', format: 'currency' },
        { key: 'payable_amount', label: 'جمع قابل پرداخت', format: 'currency' },
        { key: 'reception_count', label: 'تعداد پذیرش', format: 'number' },
      ],
    },
    header: { showGeneratedAt: true },
    footer: { text: 'صفحات به‌صورت خودکار بر اساس اندازه A4 شکسته می‌شوند.' },
    toolbar: { print: true, pdf: true, excel: true, pageSetup: true },
    export: { sheetName: 'ریز پذیرش', includeSections: true },
  };

/** ReportDefinition — کارکرد بر اساس تراکنش مالی (تجمیع کاربر) */
export const userPerformanceByTransactionDefinition: ReportDefinition<UserPerformanceTransactionRow> = {
  id: 'user-performance-by-transaction',
  title: 'گزارش کارکرد کاربران — بر اساس تراکنش مالی',
  page: { size: 'A4', orientation: 'portrait', margin: 12 },
  columns: [
    rowNumberColumn,
    {
      key: 'user_name',
      title: 'نام و نام خانوادگی کاربر',
      width: 20,
      align: 'right',
      format: 'text',
    },
    {
      key: 'total_received',
      title: 'دریافتی کل',
      width: 12,
      format: 'currency',
      formatter: (v) => formatCurrency(v),
    },
    {
      key: 'total_paid',
      title: 'پرداختی کل',
      width: 12,
      format: 'currency',
      formatter: (v) => formatCurrency(v),
    },
    {
      key: 'received_tx_count',
      title: 'تعداد تراکنش دریافتی',
      width: 12,
      format: 'number',
    },
    {
      key: 'paid_tx_count',
      title: 'تعداد تراکنش پرداختی',
      width: 12,
      format: 'number',
    },
    {
      key: 'total_tx_count',
      title: 'تعداد کل تراکنش‌ها',
      width: 12,
      format: 'number',
    },
  ],
  summary: {
    label: 'جمع کل',
    fields: [
      { key: 'total_received', label: 'جمع دریافتی', format: 'currency' },
      { key: 'total_paid', label: 'جمع پرداختی', format: 'currency' },
      { key: 'received_tx_count', label: 'تراکنش دریافتی', format: 'number' },
      { key: 'paid_tx_count', label: 'تراکنش پرداختی', format: 'number' },
      { key: 'total_tx_count', label: 'کل تراکنش‌ها', format: 'number' },
    ],
  },
  header: { showGeneratedAt: true },
  footer: { text: 'صفحات به‌صورت خودکار بر اساس اندازه A4 شکسته می‌شوند.' },
  toolbar: { print: true, pdf: true, excel: true, pageSetup: true },
  export: { sheetName: 'کارکرد تراکنش', includeSections: true },
};

/** ReportDefinition — کارکرد بر اساس تراکنش (ریز تراکنش‌ها) */
export const userPerformanceTransactionDetailDefinition: ReportDefinition<UserPerformanceTransactionDetailRow> =
  {
    id: 'user-performance-transaction-detail',
    title: 'گزارش کارکرد کاربران — ریز تراکنش‌ها',
    page: { size: 'A4', orientation: 'portrait', margin: 12 },
    columns: [
      rowNumberColumn,
      { key: 'transaction_id', title: 'شماره تراکنش', width: 10, format: 'number' },
      { key: 'file_number', title: 'شماره پرونده', width: 10, format: 'text' },
      {
        key: 'patient_name',
        title: 'نام و نام خانوادگی بیمار',
        width: 16,
        align: 'right',
        format: 'text',
      },
      { key: 'national_code', title: 'کد ملی بیمار', width: 10, format: 'text' },
      {
        key: 'received',
        title: 'دریافتی',
        width: 12,
        format: 'currency',
        formatter: (v) => formatCurrency(v),
      },
      {
        key: 'paid',
        title: 'پرداختی',
        width: 12,
        format: 'currency',
        formatter: (v) => formatCurrency(v),
      },
    ],
    summary: {
      label: 'جمع',
      fields: [
        { key: 'total_received', label: 'جمع دریافتی', format: 'currency' },
        { key: 'total_paid', label: 'جمع پرداختی', format: 'currency' },
        { key: 'received_tx_count', label: 'تراکنش دریافتی', format: 'number' },
        { key: 'paid_tx_count', label: 'تراکنش پرداختی', format: 'number' },
        { key: 'total_tx_count', label: 'کل تراکنش‌ها', format: 'number' },
      ],
    },
    header: { showGeneratedAt: true },
    footer: { text: 'صفحات به‌صورت خودکار بر اساس اندازه A4 شکسته می‌شوند.' },
    toolbar: { print: true, pdf: true, excel: true, pageSetup: true },
    export: { sheetName: 'ریز تراکنش', includeSections: true },
  };

/** meta rows برای گزارش کارکرد کاربران */
export function buildUserPerformanceMetaRows(meta: UserPerformanceMeta): ReportMetaItem[][] {
  const rows: ReportMetaItem[][] = [
    [
      { label: 'از تاریخ', value: `${formatGregorianDate(meta.from_date)} — ${meta.from_time}` },
      { label: 'تا تاریخ', value: `${formatGregorianDate(meta.to_date)} — ${meta.to_time}` },
    ],
    [
      {
        label: 'کاربران',
        value: meta.user_names.length ? meta.user_names.join('، ') : '—',
      },
    ],
  ];

  if (meta.separate_by_patient) {
    rows.push([{ label: 'نمایش', value: 'لیست به تفکیک بیماران / پرونده' }]);
  }

  return rows;
}

/** definition مناسب بر اساس نوع گزارش */
export function getUserPerformanceDefinition(
  type: UserPerformanceReportType,
  separateByPatient: boolean,
) {
  if (type === 'by_reception') {
    return separateByPatient
      ? userPerformanceReceptionDetailDefinition
      : userPerformanceByReceptionDefinition;
  }
  return separateByPatient
    ? userPerformanceTransactionDetailDefinition
    : userPerformanceByTransactionDefinition;
}

/** sections را برای ReportViewer آماده می‌کند */
export function mapUserSections<
  T extends { user_id: number; user_name: string; rows: object[]; summary: object },
>(sections: T[] | undefined) {
  return sections?.map((s) => ({
    id: String(s.user_id),
    title: s.user_name,
    rows: s.rows,
    summary: s.summary as Record<string, unknown>,
  }));
}
