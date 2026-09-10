import {
  Document,
  Font,
  Page,
  StyleSheet,
  Text,
  View,
  pdf,
} from '@react-pdf/renderer';
import type { ReportColumn, ReportExportContext, ReportRuntimePageConfig } from '../types';
import {
  PDF_FONT_VAZIRMATN_BOLD,
  PDF_FONT_VAZIRMATN_REGULAR,
  REPORT_EXPORT_ERRORS,
} from '../constants';
import { formatGeneratedAt } from '../formatters/date';
import { formatColumnValue } from '../formatters/value';
import { buildReportFilename } from '../utils/filename';
import { getPageDimensionsMm, getVisibleColumns } from '../utils/html';

let fontsRegistered = false;
let fontRegisterPromise: Promise<void> | null = null;

/** فونت فارسی را register و preload می‌کند */
async function ensurePersianFonts(): Promise<void> {
  if (fontsRegistered) return;
  if (fontRegisterPromise) return fontRegisterPromise;

  fontRegisterPromise = (async () => {
    Font.register({
      family: 'Vazirmatn',
      fonts: [
        { src: PDF_FONT_VAZIRMATN_REGULAR, fontWeight: 400 },
        { src: PDF_FONT_VAZIRMATN_BOLD, fontWeight: 700 },
      ],
    });
    await new Promise((r) => setTimeout(r, 200));
    fontsRegistered = true;
  })();

  return fontRegisterPromise;
}

/** padding صفحه PDF از margin میلی‌متر */
function pagePadding(pageConfig: ReportRuntimePageConfig): number {
  return Math.round(pageConfig.margin * 2.835);
}

/** استایل‌های پایه PDF */
function createStyles(pageConfig: ReportRuntimePageConfig) {
  const padding = pagePadding(pageConfig);
  return StyleSheet.create({
    page: {
      fontFamily: 'Vazirmatn',
      fontSize: 9,
      paddingTop: padding,
      paddingBottom: padding,
      paddingLeft: padding,
      paddingRight: padding,
    },
    header: {
      marginBottom: 12,
      borderBottomWidth: 2,
      borderBottomColor: '#222',
      paddingBottom: 8,
      textAlign: 'center',
    },
    title: {
      fontSize: 14,
      fontWeight: 700,
      marginBottom: 4,
    },
    subtitle: {
      fontSize: 10,
      color: '#555',
    },
    meta: {
      fontSize: 8,
      color: '#444',
      marginBottom: 6,
      lineHeight: 1.6,
      textAlign: 'right',
    },
    sectionTitle: {
      fontSize: 11,
      fontWeight: 700,
      color: '#1677ff',
      marginBottom: 6,
      marginTop: 8,
      textAlign: 'right',
    },
    table: {
      width: '100%',
    },
    tableHeader: {
      flexDirection: 'row-reverse',
      backgroundColor: '#efefef',
      borderWidth: 1,
      borderColor: '#333',
    },
    tableRow: {
      flexDirection: 'row-reverse',
      borderWidth: 1,
      borderColor: '#333',
      borderTopWidth: 0,
    },
    summaryRow: {
      flexDirection: 'row-reverse',
      borderWidth: 1,
      borderColor: '#333',
      borderTopWidth: 0,
      backgroundColor: '#f5f5f5',
    },
    cell: {
      flex: 1,
      padding: 4,
      textAlign: 'center',
      fontSize: 8,
      borderRightWidth: 1,
      borderRightColor: '#333',
    },
    cellLast: {
      borderRightWidth: 0,
    },
    cellBold: {
      fontWeight: 700,
    },
    footer: {
      position: 'absolute',
      bottom: padding,
      left: padding,
      right: padding,
      fontSize: 8,
      color: '#555',
      textAlign: 'left',
    },
  });
}

type PdfTableProps<T extends object> = {
  columns: ReportColumn<T>[];
  rows: T[];
  summary?: Record<string, unknown>;
  summaryLabel?: string;
  styles: ReturnType<typeof createStyles>;
};

/** جدول PDF */
function PdfTable<T extends object>({
  columns,
  rows,
  summary,
  summaryLabel = 'جمع کل',
  styles,
}: PdfTableProps<T>) {
  const visibleColumns = getVisibleColumns(columns, 'print');

  return (
    <View style={styles.table}>
      <View style={styles.tableHeader}>
        <View style={[styles.cell, { flex: 0.5 }]}>
          <Text style={styles.cellBold}>ردیف</Text>
        </View>
        {visibleColumns.map((col, i) => (
          <View
            key={String(col.key)}
            style={[styles.cell, i === visibleColumns.length - 1 ? styles.cellLast : {}]}
          >
            <Text style={styles.cellBold}>{col.title}</Text>
          </View>
        ))}
      </View>

      {rows.map((row, rowIndex) => (
        <View key={rowIndex} style={styles.tableRow} wrap={false}>
          <View style={[styles.cell, { flex: 0.5 }]}>
            <Text>{String(rowIndex + 1)}</Text>
          </View>
          {visibleColumns.map((col, colIndex) => (
            <View
              key={String(col.key)}
              style={[styles.cell, colIndex === visibleColumns.length - 1 ? styles.cellLast : {}]}
            >
              <Text>{formatColumnValue(col, row, rowIndex, 'pdf') || '—'}</Text>
            </View>
          ))}
        </View>
      ))}

      {summary && rows.length > 0 && (
        <View style={styles.summaryRow} wrap={false}>
          <View style={[styles.cell, { flex: 0.5 }]}>
            <Text style={styles.cellBold}>{summaryLabel}</Text>
          </View>
          {visibleColumns.map((col, colIndex) => {
            const val = summary[String(col.key)];
            const text =
              val != null
                ? formatColumnValue(
                    { ...col, formatter: undefined },
                    { [String(col.key)]: val } as T,
                    0,
                    'pdf',
                  )
                : colIndex === 0
                  ? summaryLabel
                  : '';
            return (
              <View
                key={String(col.key)}
                style={[styles.cell, colIndex === visibleColumns.length - 1 ? styles.cellLast : {}]}
              >
                <Text style={styles.cellBold}>{text || '—'}</Text>
              </View>
            );
          })}
        </View>
      )}
    </View>
  );
}

/** سند PDF گزارش */
function ReportPdfDocument<T extends object>(ctx: ReportExportContext<T>) {
  const { definition, data, sections, summary, meta, grandTotal, grandTotalLabel, pageConfig } =
    ctx;
  const title = definition.header?.title ?? definition.title;
  const styles = createStyles(pageConfig);
  const dims = getPageDimensionsMm(pageConfig);
  const pageSize: 'A4' | 'A5' | 'LETTER' | 'LEGAL' | [number, number] =
    pageConfig.size === 'CUSTOM'
      ? ([dims.width, dims.height] as [number, number])
      : (pageConfig.size as 'A4' | 'A5' | 'LETTER' | 'LEGAL');

  const orientLabel = pageConfig.orientation === 'landscape' ? 'افقی' : 'عمودی';

  return (
    <Document>
      <Page size={pageSize} orientation={pageConfig.orientation} style={styles.page} wrap>
        <View style={styles.header}>
          <Text style={styles.title}>{title}</Text>
          {definition.header?.subtitle || definition.description ? (
            <Text style={styles.subtitle}>
              {definition.header?.subtitle ?? definition.description}
            </Text>
          ) : null}
        </View>

        <View style={styles.meta}>
          <Text>
            تنظیمات صفحه: {pageConfig.size} — {orientLabel} — حاشیه {pageConfig.margin}mm
          </Text>
        </View>

        {(meta ?? []).map((row, i) => (
          <View key={i} style={styles.meta}>
            <Text>{row.map((item) => `${item.label}: ${item.value}`).join('  |  ')}</Text>
          </View>
        ))}

        {(definition.header?.showGeneratedAt ?? true) && (
          <View style={styles.meta}>
            <Text>تاریخ تولید: {formatGeneratedAt()}</Text>
          </View>
        )}

        {sections?.length
          ? sections.map((section) => (
              <View key={section.id}>
                <Text style={styles.sectionTitle}>{section.title}</Text>
                <PdfTable
                  columns={definition.columns}
                  rows={section.rows}
                  summary={section.summary}
                  summaryLabel={String(section.summary?.row_label ?? 'جمع')}
                  styles={styles}
                />
              </View>
            ))
          : (
            <PdfTable
              columns={definition.columns}
              rows={data}
              summary={summary}
              summaryLabel={String(summary?.row_label ?? definition.summary?.label ?? 'جمع کل')}
              styles={styles}
            />
          )}

        {grandTotal && sections && sections.length > 1 && (
          <View>
            <Text style={styles.sectionTitle}>{grandTotalLabel ?? 'جمع کل همه'}</Text>
            <PdfTable
              columns={definition.columns}
              rows={[]}
              summary={grandTotal}
              summaryLabel={grandTotalLabel ?? 'جمع کل'}
              styles={styles}
            />
          </View>
        )}

        {definition.footer?.text && (
          <Text style={styles.footer} fixed>
            {definition.footer.text}
          </Text>
        )}
      </Page>
    </Document>
  );
}

/**
 * گزارش را به PDF export می‌کند و دانلود می‌کند.
 * @param ctx context export
 */
export async function exportReportPdf<T extends object>(ctx: ReportExportContext<T>): Promise<void> {
  await ensurePersianFonts();

  try {
    const doc = <ReportPdfDocument {...ctx} />;
    const blob = await pdf(doc).toBlob();
    const url = URL.createObjectURL(blob);
    const link = document.createElement('a');
    link.href = url;
    link.download =
      ctx.definition.export?.filename?.replace(/\.xlsx$/, '.pdf') ??
      buildReportFilename(ctx.definition.id, 'pdf', ctx.definition.title);
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
    URL.revokeObjectURL(url);
  } catch (err) {
    console.error('PDF export error:', err);
    const detail = err instanceof Error ? err.message : REPORT_EXPORT_ERRORS.pdf;
    throw new Error(`${REPORT_EXPORT_ERRORS.pdf}: ${detail}`);
  }
}
