import { Typography } from 'antd';
import type { ReportFooterConfig } from '../types';

type ReportFooterProps = {
  config?: ReportFooterConfig;
};

/**
 * Footer گزارش در UI.
 */
export function ReportFooter({ config }: ReportFooterProps) {
  if (!config?.text) return null;

  return (
    <Typography.Text type="secondary" style={{ display: 'block', marginTop: 12, fontSize: 11 }}>
      {config.text}
    </Typography.Text>
  );
}
