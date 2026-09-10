import type { ReportSummaryConfig } from '../types';
import { formatCellValue } from '../formatters/value';

type ReportSummaryProps = {
  config?: ReportSummaryConfig<Record<string, unknown>>;
  summary?: Record<string, unknown>;
};

/**
 * نوار خلاصه / جمع مالی گزارش.
 */
export function ReportSummary({ config, summary }: ReportSummaryProps) {
  if (!summary || !config?.fields.length) return null;

  return (
    <div
      style={{
        marginTop: 16,
        padding: '12px 16px',
        background: '#fafafa',
        border: '1px solid #f0f0f0',
        borderRadius: 8,
        display: 'flex',
        gap: 24,
        flexWrap: 'wrap',
        fontWeight: 600,
      }}
    >
      {config.fields.map((field) => {
        const value = summary[String(field.key)];
        if (value == null && field.key !== 'row_label') return null;

        const display =
          field.formatter?.(value) ??
          (field.format ? formatCellValue(value, field.format) : String(value ?? ''));

        const label = field.label ?? String(field.key);

        return (
          <span key={String(field.key)}>
            {label}: {display}
          </span>
        );
      })}
    </div>
  );
}
