import * as XLSX from 'xlsx';
import type { ReportColumn, ReportExportContext } from '../types';
import { buildReportFilename, buildSheetName } from '../utils/filename';
import { getExcelCellValue } from '../formatters/value';
import { getVisibleColumns } from '../utils/html';

/** یک sheet از داده گزارش می‌سازد */
function buildWorksheetData<T extends object>(
  columns: ReportColumn<T>[],
  rows: T[],
  summary?: Record<string, unknown>,
  summaryLabel = 'جمع کل',
): (string | number)[][] {
  const visibleColumns = getVisibleColumns(columns, 'export');
  const header = ['ردیف', ...visibleColumns.map((c) => c.title)];

  const dataRows = rows.map((row, index) => [
    index + 1,
    ...visibleColumns.map((col) => {
      const val = getExcelCellValue(col, row, index);
      return val instanceof Date ? val.toISOString() : val;
    }),
  ]);

  const result: (string | number)[][] = [header, ...dataRows];

  if (summary && rows.length > 0) {
    const summaryRow: (string | number)[] = [
      '',
      summaryLabel,
      ...visibleColumns.slice(1).map((col) => {
        const val = summary[String(col.key)];
        if (val == null) return '';
        if (col.format === 'currency' || col.format === 'number') {
          return Number(val) || 0;
        }
        return String(val);
      }),
    ];
    result.push(summaryRow);
  }

  return result;
}

/** عرض ستون‌ها را تنظیم می‌کند */
function applyColumnWidths<T extends object>(
  ws: XLSX.WorkSheet,
  columns: ReportColumn<T>[],
): void {
  const visibleColumns = getVisibleColumns(columns, 'export');
  ws['!cols'] = [{ wch: 6 }, ...visibleColumns.map((col) => ({ wch: col.width ? Number(col.width) / 4 || 15 : 15 }))];
}

/**
 * گزارش را به فایل Excel export می‌کند.
 * @param ctx context export
 */
export function exportReportExcel<T extends object>(ctx: ReportExportContext<T>): void {
  const { definition, data, sections, summary, pageConfig: _page, grandTotal, grandTotalLabel } = ctx;
  const wb = XLSX.utils.book_new();
  const sheetName = buildSheetName(ctx.definition.export?.sheetName ?? definition.title);

  if (sections?.length) {
    sections.forEach((section, sectionIndex) => {
      const wsData = buildWorksheetData(
        definition.columns,
        section.rows,
        section.summary,
        String(section.summary?.row_label ?? 'جمع'),
      );
      const ws = XLSX.utils.aoa_to_sheet(wsData);
      applyColumnWidths(ws, definition.columns);
      const name = buildSheetName(
        sections.length > 1 ? `${section.title}` : sheetName,
      );
      XLSX.utils.book_append_sheet(wb, ws, `${name}`.slice(0, 31) || `Sheet${sectionIndex + 1}`);
    });

    if (grandTotal && sections.length > 1) {
      const wsData = buildWorksheetData(definition.columns, [], grandTotal, grandTotalLabel ?? 'جمع کل');
      const ws = XLSX.utils.aoa_to_sheet(wsData);
      applyColumnWidths(ws, definition.columns);
      XLSX.utils.book_append_sheet(wb, ws, buildSheetName('جمع کل'));
    }
  } else {
    const wsData = buildWorksheetData(
      definition.columns,
      data,
      summary,
      String(summary?.row_label ?? definition.summary?.label ?? 'جمع کل'),
    );
    const ws = XLSX.utils.aoa_to_sheet(wsData);
    applyColumnWidths(ws, definition.columns);
    XLSX.utils.book_append_sheet(wb, ws, sheetName);
  }

  const filename =
    ctx.definition.export?.filename ??
    buildReportFilename(definition.id, 'xlsx', definition.title);

  XLSX.writeFile(wb, filename);
}
