import { Alert, Divider, Empty, Spin, Typography } from 'antd';
import { useState } from 'react';
import type { ReportViewerProps, ReportSummaryConfig } from '../types';
import { useReport } from '../hooks/useReport';
import { ReportFooter } from './ReportFooter';
import { ReportHeader } from './ReportHeader';
import { ReportPageSetup } from './ReportPageSetup';
import { ReportSummary } from './ReportSummary';
import { ReportTable } from './ReportTable';
import { ReportToolbar } from './ReportToolbar';

/**
 * Component اصلی orchestration گزارش — Table, Toolbar, Export.
 */
export function ReportViewer<T extends object>({
  definition,
  data,
  sections,
  summary,
  meta,
  loading,
  error,
  emptyText = 'داده‌ای یافت نشد',
  pageSize,
  onRefresh,
  onResetFilters,
  rowKey = 'id' as keyof T,
  grandTotal,
  grandTotalLabel,
}: ReportViewerProps<T>) {
  const [pageSetupOpen, setPageSetupOpen] = useState(false);

  const {
    pageConfig,
    setPageConfig,
    handlePrint,
    handlePdf,
    handleExcel,
    exporting,
  } = useReport({
    definition,
    data,
    sections,
    summary,
    meta,
    grandTotal,
    grandTotalLabel,
  });

  const hasData = data.length > 0 || (sections?.length ?? 0) > 0;
  const toolbarDisabled = !hasData || loading;

  if (error) {
    return <Alert type="error" message={error.message || 'خطا در بارگذاری گزارش'} showIcon />;
  }

  return (
    <div>
      <div style={{ marginBottom: 16, display: 'flex', justifyContent: 'flex-end' }}>
        <ReportToolbar
          options={definition.toolbar}
          onPrint={handlePrint}
          onPdf={handlePdf}
          onExcel={handleExcel}
          onPageSetup={() => setPageSetupOpen(true)}
          onRefresh={onRefresh}
          onResetFilters={onResetFilters}
          disabled={toolbarDisabled}
          loading={exporting}
        />
      </div>

      <ReportHeader config={definition.header} title={definition.title} meta={meta} />

      <Spin spinning={loading}>
        {sections?.length ? (
          sections.map((section, index) => (
            <div key={section.id} style={{ marginBottom: index < sections.length - 1 ? 32 : 0 }}>
              <Typography.Title level={5} style={{ marginBottom: 12, color: '#1677ff' }}>
                {section.title}
              </Typography.Title>
              <ReportTable
                columns={definition.columns}
                data={section.rows}
                loading={loading}
                rowKey={rowKey}
                pageSize={pageSize}
              />
              {section.summary && definition.summary && (
                <ReportSummary
                  config={definition.summary as ReportSummaryConfig<Record<string, unknown>>}
                  summary={section.summary}
                />
              )}
              {index < sections.length - 1 && <Divider />}
            </div>
          ))
        ) : hasData ? (
          <>
            <ReportTable
              columns={definition.columns}
              data={data}
              loading={loading}
              rowKey={rowKey}
              pageSize={pageSize}
            />
            {summary && definition.summary && (
              <ReportSummary
                config={definition.summary as ReportSummaryConfig<Record<string, unknown>>}
                summary={summary}
              />
            )}
          </>
        ) : (
          !loading && <Empty description={emptyText} />
        )}

        {grandTotal && definition.summary && sections && sections.length > 1 && (
          <>
            <Divider />
            <Typography.Title level={5} style={{ marginBottom: 12 }}>
              {grandTotalLabel ?? 'جمع کل همه'}
            </Typography.Title>
            <ReportSummary
              config={definition.summary as ReportSummaryConfig<Record<string, unknown>>}
              summary={grandTotal}
            />
          </>
        )}
      </Spin>

      <ReportFooter config={definition.footer} />

      <ReportPageSetup
        open={pageSetupOpen}
        config={pageConfig}
        onClose={() => setPageSetupOpen(false)}
        onApply={setPageConfig}
      />
    </div>
  );
}
