import { Typography } from 'antd';
import type { ReportHeaderConfig, ReportMetaItem } from '../types';
import { formatGeneratedAt } from '../formatters/date';

type ReportHeaderProps = {
  config?: ReportHeaderConfig;
  title?: string;
  meta?: ReportMetaItem[][];
};

/**
 * Header گزارش — عنوان، زیرعنوان و متادیتای فیلترها.
 */
export function ReportHeader({ config, title, meta }: ReportHeaderProps) {
  const displayTitle = config?.title ?? title ?? '';
  const metaRows = config?.metaRows ?? meta ?? [];

  if (!displayTitle && metaRows.length === 0) return null;

  return (
    <div style={{ marginBottom: 16 }}>
      {displayTitle && (
        <Typography.Title level={4} style={{ marginBottom: 4 }}>
          {displayTitle}
        </Typography.Title>
      )}
      {config?.subtitle && (
        <Typography.Text type="secondary">{config.subtitle}</Typography.Text>
      )}
      {metaRows.map((row, i) => (
        <div
          key={i}
          style={{
            display: 'flex',
            gap: 24,
            flexWrap: 'wrap',
            fontSize: 12,
            color: '#666',
            marginTop: 8,
          }}
        >
          {row.map((item) => (
            <span key={item.label}>
              {item.label}: <strong>{item.value}</strong>
            </span>
          ))}
        </div>
      ))}
      {(config?.showGeneratedAt ?? false) && (
        <Typography.Text type="secondary" style={{ fontSize: 11, display: 'block', marginTop: 4 }}>
          تاریخ تولید: {formatGeneratedAt()}
        </Typography.Text>
      )}
    </div>
  );
}
