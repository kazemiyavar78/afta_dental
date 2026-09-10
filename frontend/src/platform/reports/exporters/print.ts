import type { ReportColumn, ReportExportContext, ReportMetaItem, ReportRuntimePageConfig } from '../types';
import { REPORT_EXPORT_ERRORS } from '../constants';
import { formatGeneratedAt } from '../formatters/date';
import { formatColumnValue } from '../formatters/value';
import {
  alignToCss,
  buildPageConfigLabel,
  buildPageCss,
  buildRowIndexColumn,
  escapeHtml,
  getPageDimensionsMm,
  getVisibleColumns,
} from '../utils/html';

/** CSS پایه چاپ RTL — شامل پیش‌نمایش صفحه روی screen */
function buildBasePrintStyles(pageCss: string, pageConfig: ReportRuntimePageConfig): string {
  const dims = getPageDimensionsMm(pageConfig);

  return `
    ${pageCss}
    * { box-sizing: border-box; }
    html, body {
      font-family: Tahoma, "Segoe UI", sans-serif;
      font-size: 11px;
      color: #111;
      margin: 0;
      direction: rtl;
    }
    @media screen {
      html { background: #666; }
      body {
        width: ${dims.width}mm;
        min-height: ${dims.height}mm;
        margin: 16px auto;
        padding: ${pageConfig.margin}mm;
        background: #fff;
        box-shadow: 0 2px 16px rgba(0,0,0,0.35);
      }
    }
    @media print {
      html, body {
        width: auto;
        min-height: auto;
        margin: 0;
        padding: 0;
        background: #fff;
        box-shadow: none;
      }
      .page-setup-info { display: none; }
      body { -webkit-print-color-adjust: exact; print-color-adjust: exact; }
    }
    .report-header {
      text-align: center;
      margin-bottom: 12px;
      border-bottom: 2px solid #222;
      padding-bottom: 8px;
    }
    .report-header h1 {
      margin: 0 0 4px;
      font-size: 16px;
    }
    .report-header .subtitle {
      margin: 0;
      font-size: 11px;
      color: #555;
    }
    .report-meta {
      margin-bottom: 10px;
      font-size: 10px;
      color: #444;
      line-height: 1.8;
    }
    .report-meta-row {
      display: flex;
      justify-content: space-between;
      gap: 12px;
      flex-wrap: wrap;
    }
    .report-section {
      margin-bottom: 18px;
      page-break-inside: avoid;
    }
    .report-section-title {
      margin: 0 0 8px;
      font-size: 13px;
      color: #1677ff;
      border-bottom: 1px solid #ddd;
      padding-bottom: 4px;
    }
    .report-grand-total {
      margin-top: 16px;
      page-break-inside: avoid;
    }
    table {
      width: 100%;
      border-collapse: collapse;
      table-layout: fixed;
    }
    th, td {
      border: 1px solid #333;
      padding: 4px 5px;
      text-align: center;
      word-wrap: break-word;
      vertical-align: middle;
    }
    th {
      background: #efefef;
      font-weight: 700;
      font-size: 10px;
    }
    tr { page-break-inside: avoid; }
    thead { display: table-header-group; }
    tfoot { display: table-footer-group; }
    .summary-row td {
      background: #f5f5f5;
      font-weight: 700;
    }
    .report-footer {
      margin-top: 10px;
      font-size: 10px;
      text-align: left;
      color: #555;
    }
    .page-setup-info {
      font-size: 9px;
      color: #666;
      margin-bottom: 8px;
      padding: 4px 8px;
      background: #f0f0f0;
      border-radius: 4px;
      text-align: center;
    }
  `;
}

/** HTML header */
function buildHeaderHtml(
  title: string,
  subtitle?: string,
  meta?: ReportMetaItem[][],
  showGeneratedAt = true,
): string {
  const metaHtml = (meta ?? [])
    .map(
      (row) =>
        `<div class="report-meta-row">${row
          .map((item) => `<span>${escapeHtml(item.label)}: ${escapeHtml(item.value)}</span>`)
          .join('')}</div>`,
    )
    .join('');

  const generatedAtRow =
    showGeneratedAt && metaHtml
      ? ''
      : showGeneratedAt
        ? `<div class="report-meta-row"><span>تاریخ چاپ: ${escapeHtml(formatGeneratedAt())}</span></div>`
        : '';

  return `
    <div class="report-header">
      <h1>${escapeHtml(title)}</h1>
      ${subtitle ? `<p class="subtitle">${escapeHtml(subtitle)}</p>` : ''}
    </div>
    ${metaHtml || generatedAtRow ? `<div class="report-meta">${metaHtml}${generatedAtRow}</div>` : ''}
  `;
}

/** HTML یک جدول */
function buildTableHtml<T extends object>(
  columns: ReportColumn<T>[],
  rows: T[],
  summary?: Record<string, unknown>,
  summaryLabel = 'جمع کل',
  includeRowIndex = true,
): string {
  const visibleColumns = getVisibleColumns(columns, 'print');
  const colCount = visibleColumns.length + (includeRowIndex ? 1 : 0);

  const headerCells = includeRowIndex
    ? `<th style="width:4%">ردیف</th>${visibleColumns
        .map((col) => `<th${col.width ? ` style="width:${col.width}${typeof col.width === 'number' ? '%' : ''}"` : ''}>${escapeHtml(col.title)}</th>`)
        .join('')}`
    : visibleColumns
        .map((col) => `<th${col.width ? ` style="width:${col.width}${typeof col.width === 'number' ? '%' : ''}"` : ''}>${escapeHtml(col.title)}</th>`)
        .join('');

  const bodyRows = rows
    .map((row, index) => {
      const cells = visibleColumns
        .map((col) => {
          const text = formatColumnValue(col, row, index, 'print');
          const align = alignToCss(col.align);
          return `<td style="text-align:${align}">${escapeHtml(text)}</td>`;
        })
        .join('');
      const indexCell = includeRowIndex ? `<td>${buildRowIndexColumn(index)}</td>` : '';
      return `<tr>${indexCell}${cells}</tr>`;
    })
    .join('');

  let summaryRow = '';
  if (summary && rows.length > 0) {
    const summaryCells = visibleColumns
      .map((col, colIndex) => {
        if (colIndex === 0 && includeRowIndex) {
          return `<td colspan="2"><strong>${escapeHtml(summaryLabel)}</strong></td>`;
        }
        if (colIndex === 0 && !includeRowIndex) {
          return `<td><strong>${escapeHtml(summaryLabel)}</strong></td>`;
        }
        const val = summary[String(col.key)];
        const text =
          val != null
            ? formatColumnValue({ ...col, formatter: undefined }, { [String(col.key)]: val } as T, 0, 'print')
            : '';
        return `<td><strong>${escapeHtml(text)}</strong></td>`;
      })
      .join('');

    if (includeRowIndex) {
      summaryRow = `<tr class="summary-row">${summaryCells}</tr>`;
    } else {
      summaryRow = `<tr class="summary-row">${summaryCells}</tr>`;
    }
  }

  return `
    <table>
      <thead><tr>${headerCells}</tr></thead>
      <tbody>
        ${bodyRows || `<tr><td colspan="${colCount}">موردی برای چاپ یافت نشد</td></tr>`}
        ${summaryRow}
      </tbody>
    </table>
  `;
}

/** HTML کامل گزارش */
export function buildPrintHtml<T extends object>(ctx: ReportExportContext<T>): string {
  const { definition, data, sections, summary, meta, pageConfig, grandTotal, grandTotalLabel } =
    ctx;
  const title = definition.header?.title ?? definition.title;
  const pageCss = buildPageCss(pageConfig);

  let contentHtml = '';

  if (sections?.length) {
    contentHtml = sections
      .map(
        (section) => `
        <div class="report-section">
          <h2 class="report-section-title">${escapeHtml(section.title)}</h2>
          ${buildTableHtml(definition.columns, section.rows, section.summary, section.summary?.row_label as string ?? 'جمع')}
        </div>`,
      )
      .join('');

    if (grandTotal && sections.length > 1) {
      contentHtml += `
        <div class="report-grand-total">
          <h2 class="report-section-title">${escapeHtml(grandTotalLabel ?? 'جمع کل همه')}</h2>
          ${buildTableHtml(definition.columns, [], grandTotal, grandTotalLabel ?? 'جمع کل', false)}
        </div>`;
    }
  } else {
    contentHtml = buildTableHtml(
      definition.columns,
      data,
      summary,
      (summary?.row_label as string) ?? definition.summary?.label ?? 'جمع کل',
    );
  }

  const footerText =
    definition.footer?.text ?? 'صفحات به‌صورت خودکار بر اساس اندازه صفحه شکسته می‌شوند.';

  const pageSetupInfo = buildPageConfigLabel(pageConfig);

  return `<!DOCTYPE html>
<html lang="fa" dir="rtl">
<head>
  <meta charset="utf-8" />
  <title>${escapeHtml(title)}</title>
  <style>${buildBasePrintStyles(pageCss, pageConfig)}</style>
</head>
<body>
  <div class="page-setup-info">تنظیمات صفحه: ${escapeHtml(pageSetupInfo)}</div>
  ${buildHeaderHtml(title, definition.header?.subtitle ?? definition.description, meta, definition.header?.showGeneratedAt ?? true)}
  ${contentHtml}
  <div class="report-footer">${escapeHtml(footerText)}</div>
  <script>
    window.onload = function () {
      window.focus();
      window.print();
    };
  </script>
</body>
</html>`;
}

/**
 * گزارش را در پنجره جدید برای چاپ باز می‌کند.
 * @param ctx context export
 */
export function printReport<T extends object>(ctx: ReportExportContext<T>): void {
  const html = buildPrintHtml(ctx);
  const win = window.open('', '_blank');
  if (!win) {
    throw new Error(REPORT_EXPORT_ERRORS.popup);
  }
  win.document.open();
  win.document.write(html);
  win.document.close();
}

/**
 * HTML چاپ را بدون باز کردن پنجره برمی‌گرداند (برای documentهای custom).
 * @param ctx context export
 */
export function buildReportPrintDocument<T extends object>(ctx: ReportExportContext<T>): string {
  return buildPrintHtml(ctx);
}
