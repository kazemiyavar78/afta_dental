import type { ReactNode } from 'react';

/** اندازه صفحه گزارش */
export type ReportPageSize = 'A4' | 'A5' | 'LETTER' | 'LEGAL' | 'CUSTOM';

/** جهت صفحه */
export type ReportOrientation = 'portrait' | 'landscape';

/** پیکربندی صفحه */
export type ReportPageConfig = {
  size: ReportPageSize;
  orientation: ReportOrientation;
  margin: number;
  customWidth?: number;
  customHeight?: number;
};

/** preset حاشیه */
export type ReportMarginPreset = 'small' | 'medium' | 'large' | 'custom';

/** تنظیمات runtime صفحه (بدون تغییر definition اصلی) */
export type ReportRuntimePageConfig = ReportPageConfig & {
  marginPreset?: ReportMarginPreset;
};

/** فرمت پیش‌فرض ستون */
export type ReportColumnFormat =
  | 'text'
  | 'number'
  | 'currency'
  | 'date'
  | 'datetime'
  | 'percent';

/** تعریف یک ستون گزارش */
export type ReportColumn<T = unknown> = {
  key: keyof T | string;
  title: string;
  width?: number | string;
  align?: 'left' | 'center' | 'right';
  format?: ReportColumnFormat;
  formatter?: (value: unknown, row: T, index: number) => ReactNode;
  excelFormatter?: (value: unknown, row: T, index: number) => string | number | Date;
  pdfFormatter?: (value: unknown, row: T, index: number) => string;
  printFormatter?: (value: unknown, row: T, index: number) => string;
  hidden?: boolean;
  printable?: boolean;
  exportable?: boolean;
};

/** ردیف خلاصه / جمع */
export type ReportSummaryConfig<T = unknown> = {
  label?: string;
  fields: Array<{
    key: keyof T | string;
    label?: string;
    format?: ReportColumnFormat;
    formatter?: (value: unknown) => ReactNode;
  }>;
};

/** بخش گزارش (مثلاً تفکیک پزشک) */
export type ReportSection<T = unknown> = {
  id: string;
  title: string;
  rows: T[];
  summary?: Record<string, unknown>;
};

/** متادیتای فیلتر / context گزارش */
export type ReportMetaItem = {
  label: string;
  value: string;
};

/** Header گزارش */
export type ReportHeaderConfig = {
  title?: string;
  subtitle?: string;
  showGeneratedAt?: boolean;
  metaRows?: ReportMetaItem[][];
};

/** Footer گزارش */
export type ReportFooterConfig = {
  text?: string;
  showPageNumber?: boolean;
};

/** گزینه‌های toolbar */
export type ReportToolbarOptions = {
  print?: boolean;
  pdf?: boolean;
  excel?: boolean;
  pageSetup?: boolean;
  refresh?: boolean;
  resetFilters?: boolean;
};

/** گزینه‌های export */
export type ReportExportOptions = {
  filename?: string;
  sheetName?: string;
  includeSummary?: boolean;
  includeSections?: boolean;
};

/** تعریف کامل یک گزارش — Single Source of Truth */
export type ReportDefinition<T = unknown> = {
  id: string;
  title: string;
  description?: string;
  columns: ReportColumn<T>[];
  page: ReportPageConfig;
  header?: ReportHeaderConfig;
  footer?: ReportFooterConfig;
  summary?: ReportSummaryConfig<T>;
  toolbar?: ReportToolbarOptions;
  export?: ReportExportOptions;
};

/** Adapter برای تبدیل داده API به سطرهای گزارش */
export type ReportDataAdapter<TInput, TOutput> = (data: TInput) => TOutput[];

/** Props اصلی ReportViewer */
export type ReportViewerProps<T extends object> = {
  definition: ReportDefinition<T>;
  data: T[];
  sections?: ReportSection<T>[];
  summary?: Record<string, unknown>;
  meta?: ReportMetaItem[][];
  loading?: boolean;
  error?: Error | null;
  emptyText?: string;
  pageSize?: number;
  onRefresh?: () => void;
  onResetFilters?: () => void;
  rowKey?: keyof T | ((row: T) => string);
  grandTotal?: Record<string, unknown>;
  grandTotalLabel?: string;
};

/** Context runtime برای export */
export type ReportExportContext<T extends object> = {
  definition: ReportDefinition<T>;
  data: T[];
  sections?: ReportSection<T>[];
  summary?: Record<string, unknown>;
  meta?: ReportMetaItem[][];
  pageConfig: ReportRuntimePageConfig;
  grandTotal?: Record<string, unknown>;
  grandTotalLabel?: string;
};
