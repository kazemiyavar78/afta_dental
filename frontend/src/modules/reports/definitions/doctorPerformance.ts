import type { ReportDefinition, ReportMetaItem } from '@/platform/reports';
import { formatCurrency, formatGregorianDate } from '@/platform/reports';
import type {
  DoctorPerformanceByOrganizationRow,
  DoctorPerformanceByServiceRow,
  DoctorPerformanceMeta,
  DoctorPerformancePatientRow,
  DoctorPerformanceReportType,
  RevenueByDoctorsRow,
} from '../types';
import { organizationReportColumns, organizationReportSummary } from './revenueByOrganization';

/** ReportDefinition — کارکرد به تفکیک سازمان */
export const doctorPerformanceByOrgDefinition: ReportDefinition<DoctorPerformanceByOrganizationRow> =
  {
    id: 'doctor-performance-by-organization',
    title: 'گزارش کارکرد پزشکان — به تفکیک سازمان',
    page: { size: 'A4', orientation: 'portrait', margin: 12 },
    columns: organizationReportColumns as ReportDefinition<DoctorPerformanceByOrganizationRow>['columns'],
    summary: organizationReportSummary,
    header: { showGeneratedAt: true },
    footer: { text: 'صفحات به‌صورت خودکار بر اساس اندازه A4 شکسته می‌شوند.' },
    toolbar: { print: true, pdf: true, excel: true, pageSetup: true },
    export: { sheetName: 'کارکرد سازمان' },
  };

/** ReportDefinition — کارکرد به تفکیک خدمت */
export const doctorPerformanceByServiceDefinition: ReportDefinition<DoctorPerformanceByServiceRow> =
  {
    id: 'doctor-performance-by-service',
    title: 'گزارش کارکرد پزشکان — به تفکیک خدمت',
    page: { size: 'A4', orientation: 'portrait', margin: 12 },
    columns: [
      {
        key: 'service_name',
        title: 'خدمت',
        width: 22,
        align: 'right',
        format: 'text',
      },
      {
        key: 'service_amount',
        title: 'نرخ',
        width: 11,
        format: 'currency',
        formatter: (v) => formatCurrency(v),
      },
      {
        key: 'service_tariff',
        title: 'تعرفه',
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
    export: { sheetName: 'کارکرد خدمت' },
  };

/** ReportDefinition — لیست بیماران پزشکان */
export const doctorPerformancePatientListDefinition: ReportDefinition<DoctorPerformancePatientRow> =
  {
    id: 'doctor-performance-patient-list',
    title: 'گزارش کارکرد پزشکان — لیست بیماران',
    page: { size: 'A4', orientation: 'portrait', margin: 12 },
    columns: [
      {
        key: 'patient_name',
        title: 'نام و نام خانوادگی',
        width: 12,
        align: 'right',
        format: 'text',
      },
      { key: 'national_code', title: 'کدملی', width: 9, format: 'text' },
      {
        key: 'service_amount',
        title: 'نرخ خدمات',
        width: 9,
        format: 'currency',
        formatter: (v) => formatCurrency(v),
      },
      {
        key: 'service_tariff',
        title: 'تعرفه خدمات',
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
        width: 9,
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
        width: 23,
        align: 'right',
        format: 'text',
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
      ],
    },
    header: { showGeneratedAt: true },
    footer: { text: 'صفحات به‌صورت خودکار بر اساس اندازه A4 شکسته می‌شوند.' },
    toolbar: { print: true, pdf: true, excel: true, pageSetup: true },
    export: { sheetName: 'لیست بیماران' },
  };

/** ReportDefinition — درآمد به تفکیک پزشکان */
export const revenueByDoctorsDefinition: ReportDefinition<RevenueByDoctorsRow> = {
  id: 'revenue-by-doctors',
  title: 'گزارش درآمد به تفکیک پزشکان',
  page: { size: 'A4', orientation: 'portrait', margin: 12 },
  columns: [
    { key: 'medical_code', title: 'کد پزشک', width: 10, format: 'text' },
    {
      key: 'doctor_name',
      title: 'نام و نام خانوادگی',
      width: 18,
      align: 'right',
      format: 'text',
    },
    {
      key: 'reception_count',
      title: 'تعداد',
      width: 8,
      format: 'number',
    },
    {
      key: 'service_amount',
      title: 'نرخ',
      width: 10,
      format: 'currency',
      formatter: (v) => formatCurrency(v),
    },
    {
      key: 'service_tariff',
      title: 'تعرفه',
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
      title: 'یارانه',
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
  ],
  summary: {
    label: 'جمع کل',
    fields: [
      { key: 'reception_count', label: 'تعداد پذیرش', format: 'number' },
      { key: 'service_amount', label: 'جمع نرخ', format: 'currency' },
      { key: 'service_tariff', label: 'جمع تعرفه', format: 'currency' },
      { key: 'organization_share', label: 'جمع سهم سازمان', format: 'currency' },
      { key: 'supplementary_share', label: 'جمع تکمیلی', format: 'currency' },
      { key: 'subsidy_share', label: 'جمع یارانه', format: 'currency' },
      { key: 'payable_amount', label: 'جمع قابل پرداخت', format: 'currency' },
    ],
  },
  header: { showGeneratedAt: true },
  footer: { text: 'صفحات به‌صورت خودکار بر اساس اندازه A4 شکسته می‌شوند.' },
  toolbar: { print: true, pdf: true, excel: true, pageSetup: true },
  export: { sheetName: 'درآمد پزشکان' },
};

/** meta rows برای گزارش کارکرد پزشکان */
export function buildDoctorPerformanceMetaRows(meta: DoctorPerformanceMeta): ReportMetaItem[][] {
  const rows: ReportMetaItem[][] = [
    [
      { label: 'از تاریخ', value: `${formatGregorianDate(meta.from_date)} — ${meta.from_time}` },
      { label: 'تا تاریخ', value: `${formatGregorianDate(meta.to_date)} — ${meta.to_time}` },
    ],
    [
      {
        label: 'پزشکان',
        value: meta.doctor_names.length ? meta.doctor_names.join('، ') : '—',
      },
    ],
  ];

  if (meta.separate_doctors) {
    rows.push([{ label: 'نمایش', value: 'پزشکان به صورت جدا' }]);
  }

  if (meta.report_type === 'by_organization') {
    let filterLabel = 'فقط پذیرش بدون بیمه تکمیلی';
    if (meta.only_supplementary) {
      filterLabel = `فقط سازمان تکمیلی — پایه: ${meta.free_organization_name ?? 'سازمان آزاد'}`;
    } else if (meta.has_supplementary_filter) {
      filterLabel = meta.detail_supplementary ? 'ریز بیمه‌های تکمیلی: بله' : 'ریز بیمه‌های تکمیلی: خیر';
    }
    rows.push([{ label: 'فیلتر', value: filterLabel }]);
  }

  if (meta.report_type === 'by_service' && meta.service_names?.length) {
    rows.push([{ label: 'خدمات', value: meta.service_names.join('، ') }]);
  }

  return rows;
}

/** definition مناسب بر اساس نوع گزارش */
export function getDoctorPerformanceDefinition(type: DoctorPerformanceReportType) {
  switch (type) {
    case 'by_organization':
      return doctorPerformanceByOrgDefinition;
    case 'by_service':
      return doctorPerformanceByServiceDefinition;
    case 'patient_list':
      return doctorPerformancePatientListDefinition;
    case 'revenue_by_doctors':
      return revenueByDoctorsDefinition;
  }
}

/** sections را برای ReportViewer آماده می‌کند */
export function mapDoctorSections<T extends { doctor_id: number; doctor_name: string; rows: object[]; summary: object }>(
  sections: T[] | undefined,
) {
  return sections?.map((s) => ({
    id: String(s.doctor_id),
    title: s.doctor_name,
    rows: s.rows,
    summary: s.summary as Record<string, unknown>,
  }));
}
