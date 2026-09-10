import type { ReportColumn, ReportPageConfig, ReportPageSize, ReportRuntimePageConfig } from '../types';
import { DEFAULT_REPORT_PAGE, REPORT_MARGIN_PRESETS } from '../constants';

/** HTML را برای چاپ امن escape می‌کند */
export function escapeHtml(value: string | number | null | undefined): string {
  if (value == null || value === '') return '';
  return String(value)
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#39;');
}

/** CSS @page size rule تولید می‌کند */
export function buildPageSizeCss(page: ReportPageConfig): string {
  if (page.size === 'CUSTOM' && page.customWidth && page.customHeight) {
    const unit = typeof page.customWidth === 'number' ? 'mm' : '';
    return `${page.customWidth}${unit} ${page.customHeight}${unit}`;
  }
  return `${page.size} ${page.orientation}`;
}

/** CSS @page block کامل */
export function buildPageCss(page: ReportRuntimePageConfig): string {
  return `@page { size: ${buildPageSizeCss(page)}; margin: ${page.margin}mm; }`;
}

/** page config پیش‌فرض از definition */
export function resolvePageConfig(
  definitionPage: ReportPageConfig,
  runtime?: Partial<ReportRuntimePageConfig>,
): ReportRuntimePageConfig {
  return {
    ...DEFAULT_REPORT_PAGE,
    ...definitionPage,
    ...runtime,
    margin: runtime?.margin ?? definitionPage.margin ?? DEFAULT_REPORT_PAGE.margin,
  };
}

/** margin از preset */
export function marginFromPreset(preset: keyof typeof REPORT_MARGIN_PRESETS): number {
  return REPORT_MARGIN_PRESETS[preset];
}

/** align CSS */
export function alignToCss(align?: 'left' | 'center' | 'right'): string {
  return align ?? 'center';
}

/** ستون‌های قابل export/print */
export function getVisibleColumns<T extends object>(
  columns: ReportColumn<T>[],
  mode: 'print' | 'export' | 'ui',
): ReportColumn<T>[] {
  return columns.filter((col) => {
    if (col.hidden) return false;
    if (mode === 'print' && col.printable === false) return false;
    if (mode === 'export' && col.exportable === false) return false;
    return true;
  });
}

/** اندازه صفحات استاندارد (میلی‌متر) */
export const PAGE_DIMENSIONS_MM: Record<
  Exclude<ReportPageSize, 'CUSTOM'>,
  { width: number; height: number }
> = {
  A4: { width: 210, height: 297 },
  A5: { width: 148, height: 210 },
  LETTER: { width: 216, height: 279 },
  LEGAL: { width: 216, height: 356 },
};

/** برچسب فارسی تنظیمات صفحه برای چاپ */
export function buildPageConfigLabel(page: ReportRuntimePageConfig): string {
  const orient = page.orientation === 'landscape' ? 'افقی' : 'عمودی';
  return `${page.size} — ${orient} — حاشیه ${page.margin}mm`;
}

/** ابعاد صفحه به میلی‌متر */
export function getPageDimensionsMm(page: ReportRuntimePageConfig): { width: number; height: number } {
  if (page.size === 'CUSTOM' && page.customWidth && page.customHeight) {
    return { width: page.customWidth, height: page.customHeight };
  }
  const base = PAGE_DIMENSIONS_MM[page.size as keyof typeof PAGE_DIMENSIONS_MM] ?? PAGE_DIMENSIONS_MM.A4;
  if (page.orientation === 'landscape') {
    return { width: base.height, height: base.width };
  }
  return base;
}

/** index ستون ردیف */
export function buildRowIndexColumn(index: number): string {
  return (index + 1).toLocaleString('fa-IR');
}
