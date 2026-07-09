import { Typography, Space } from 'antd';
import type { ReactNode } from 'react';

const { Title } = Typography;

type PageHeaderProps = {
  title: string;
  subtitle?: string;
  extra?: ReactNode;
};

/**
 * هدر استاندارد صفحات.
 * @example <PageHeader title="کاربران" extra={<Button>ایجاد</Button>} />
 */
export function PageHeader({ title, subtitle, extra }: PageHeaderProps) {
  return (
    <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 24 }}>
      <Space direction="vertical" size={0}>
        <Title level={3} style={{ margin: 0 }}>
          {title}
        </Title>
        {subtitle && <Typography.Text type="secondary">{subtitle}</Typography.Text>}
      </Space>
      {extra}
    </div>
  );
}
