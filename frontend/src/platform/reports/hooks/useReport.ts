import { useCallback, useMemo, useRef, useState } from 'react';
import { message } from 'antd';
import type {
  ReportDefinition,
  ReportExportContext,
  ReportMetaItem,
  ReportRuntimePageConfig,
  ReportSection,
  ReportViewerProps,
} from '../types';
import { DEFAULT_REPORT_PAGE, REPORT_EXPORT_ERRORS } from '../constants';
import { exportReportExcel } from '../exporters/excel';
import { exportReportPdf } from '../exporters/pdf';
import { printReport } from '../exporters/print';
import { resolvePageConfig } from '../utils/html';

type UseReportOptions<T extends object> = Pick<
  ReportViewerProps<T>,
  | 'definition'
  | 'data'
  | 'sections'
  | 'summary'
  | 'meta'
  | 'grandTotal'
  | 'grandTotalLabel'
>;

type UseReportReturn = {
  pageConfig: ReportRuntimePageConfig;
  setPageConfig: (config: Partial<ReportRuntimePageConfig>) => void;
  exportContext: ReportExportContext<object>;
  handlePrint: () => void;
  handlePdf: () => Promise<void>;
  handleExcel: () => void;
  exporting: boolean;
};

/**
 * Hook مدیریت runtime config و export actions گزارش.
 * @param options داده و definition گزارش
 */
export function useReport<T extends object>(options: UseReportOptions<T>): UseReportReturn {
  const {
    definition,
    data,
    sections,
    summary,
    meta,
    grandTotal,
    grandTotalLabel,
  } = options;

  const [pageConfig, setPageConfigState] = useState<ReportRuntimePageConfig>(() =>
    resolvePageConfig(definition.page ?? DEFAULT_REPORT_PAGE),
  );
  const [exporting, setExporting] = useState(false);

  const setPageConfig = useCallback((partial: Partial<ReportRuntimePageConfig>) => {
    setPageConfigState((prev) => ({ ...prev, ...partial }));
  }, []);

  const exportContext = useMemo(
    (): ReportExportContext<object> => ({
      definition: definition as ReportDefinition<object>,
      data: data as object[],
      sections: sections as ReportSection<object>[] | undefined,
      summary,
      meta: meta as ReportMetaItem[][] | undefined,
      pageConfig,
      grandTotal,
      grandTotalLabel,
    }),
    [definition, data, sections, summary, meta, pageConfig, grandTotal, grandTotalLabel],
  );

  const exportContextRef = useRef(exportContext);
  exportContextRef.current = exportContext;

  const hasExportData = data.length > 0 || (sections?.length ?? 0) > 0;

  const handlePrint = useCallback(() => {
    if (!hasExportData) {
      message.warning('داده‌ای برای چاپ وجود ندارد.');
      return;
    }
    try {
      printReport(exportContextRef.current);
    } catch (err) {
      message.error(err instanceof Error ? err.message : REPORT_EXPORT_ERRORS.print);
    }
  }, [hasExportData]);

  const handlePdf = useCallback(async () => {
    if (!hasExportData) {
      message.warning('داده‌ای برای export وجود ندارد.');
      return;
    }
    setExporting(true);
    try {
      await exportReportPdf(exportContextRef.current);
    } catch (err) {
      message.error(err instanceof Error ? err.message : REPORT_EXPORT_ERRORS.pdf);
    } finally {
      setExporting(false);
    }
  }, [hasExportData]);

  const handleExcel = useCallback(() => {
    if (!hasExportData) {
      message.warning('داده‌ای برای export وجود ندارد.');
      return;
    }
    try {
      exportReportExcel(exportContextRef.current);
    } catch (err) {
      message.error(err instanceof Error ? err.message : REPORT_EXPORT_ERRORS.excel);
    }
  }, [hasExportData]);

  return {
    pageConfig,
    setPageConfig,
    exportContext,
    handlePrint,
    handlePdf,
    handleExcel,
    exporting,
  };
}
