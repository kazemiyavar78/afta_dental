import type { ReportColumn, ReportColumnFormat } from '../types';
import { formatCurrency, formatNumber, formatPercent } from './number';
import { formatDate, formatDateTime } from './date';

/** مقدار ستون را بر اساس format پیش‌فرض به string تبدیل می‌کند */
export function formatCellValue(
  value: unknown,
  format: ReportColumnFormat = 'text',
): string {
  if (value == null || value === '') return '—';

  switch (format) {
    case 'number':
      return formatNumber(value);
    case 'currency':
      return formatCurrency(value);
    case 'date':
      return formatDate(String(value));
    case 'datetime':
      return formatDateTime(String(value));
    case 'percent':
      return formatPercent(value);
    default:
      return String(value);
  }
}

/** مقدار یک فیلد از row را می‌خواند */
export function getRowValue<T extends object>(
  row: T,
  key: keyof T | string,
): unknown {
  return (row as Record<string, unknown>)[String(key)];
}

/** مقدار سلول را برای UI format می‌کند */
export function formatColumnValue<T extends object>(
  column: ReportColumn<T>,
  row: T,
  index: number,
  mode: 'ui' | 'print' | 'pdf' | 'excel' = 'ui',
): string {
  const value = getRowValue(row, column.key);

  if (mode === 'print' && column.printFormatter) {
    return column.printFormatter(value, row, index);
  }
  if (mode === 'pdf' && column.pdfFormatter) {
    return column.pdfFormatter(value, row, index);
  }
  if (mode === 'excel' && column.excelFormatter) {
    const result = column.excelFormatter(value, row, index);
    return result instanceof Date ? result.toISOString() : String(result);
  }

  if (column.format) {
    return formatCellValue(value, column.format);
  }

  if (value == null || value === '') return '—';
  return String(value);
}

/** مقدار خام برای Excel export */
export function getExcelCellValue<T extends object>(
  column: ReportColumn<T>,
  row: T,
  index: number,
): string | number | Date {
  const value = getRowValue(row, column.key);
  if (column.excelFormatter) {
    return column.excelFormatter(value, row, index);
  }
  if (column.format === 'currency' || column.format === 'number') {
    const num = Number(value ?? 0);
    return Number.isNaN(num) ? 0 : num;
  }
  if (value == null || value === '') return '';
  return String(value);
}
