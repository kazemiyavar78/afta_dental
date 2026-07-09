import { Card, Form, Space, Button } from 'antd';
import type { ReactNode } from 'react';

type FormLayoutProps = {
  title?: string;
  children: ReactNode;
  onSubmit: () => void;
  onCancel?: () => void;
  loading?: boolean;
  submitLabel?: string;
};

/**
 * قالب فرم استاندارد با دکمه‌های ذخیره/انصراف.
 * @example <FormLayout onSubmit={handleSubmit} onCancel={() => navigate(-1)}>...</FormLayout>
 */
export function FormLayout({
  title,
  children,
  onSubmit,
  onCancel,
  loading,
  submitLabel = 'ذخیره',
}: FormLayoutProps) {
  return (
    <Card title={title}>
      <Form layout="vertical" onFinish={onSubmit}>
        {children}
        <Form.Item>
          <Space>
            <Button type="primary" htmlType="submit" loading={loading}>
              {submitLabel}
            </Button>
            {onCancel && <Button onClick={onCancel}>انصراف</Button>}
          </Space>
        </Form.Item>
      </Form>
    </Card>
  );
}
