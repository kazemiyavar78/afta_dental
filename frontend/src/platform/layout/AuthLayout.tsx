import { Layout, Typography } from 'antd';
import type { ReactNode } from 'react';

const { Content } = Layout;

type AuthLayoutProps = {
  children: ReactNode;
};

/** قالب صفحه لاگین (بدون سایدبار) */
export function AuthLayout({ children }: AuthLayoutProps) {
  return (
    <Layout style={{ minHeight: '100vh', background: '#f0f2f5' }}>
      <Content
        style={{
          display: 'flex',
          flexDirection: 'column',
          alignItems: 'center',
          justifyContent: 'center',
        }}
      >
        <Typography.Title level={2} style={{ marginBottom: 32 }}>
          سیستم پذیرش بیماران
        </Typography.Title>
        {children}
      </Content>
    </Layout>
  );
}
