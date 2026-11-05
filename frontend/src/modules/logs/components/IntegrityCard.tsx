import { useState } from 'react';
import { Alert, Button, Card, Flex, List, Modal, Spin } from 'antd';
import { formatDateTime } from '@/platform/reports/formatters/date';
import { useLogIntegrity } from '../hooks';
import type { OpenIntegrityFinding } from '../types';

/**
 * کارت وضعیت صحت لاگ‌ها را از نتیجه آخرین بررسی زنجیره و یافته‌های باز نشان می‌دهد.
 * سالم سبز است و زمان بررسی را دارد. خراب قرمز است و شناسه اولین ردیف خراب را دارد.
 * دکمه تعداد یافته‌های باز را نشان می‌دهد و با کلیک فهرست نوع، نام و اولین زمان را باز می‌کند.
 */
export function IntegrityCard() {
  const { data, isLoading, isError } = useLogIntegrity();
  const [open, setOpen] = useState(false);
  const findings = data?.open_findings ?? [];

  let body = <Alert type="info" showIcon message="هنوز بررسی نشده" />;
  if (isLoading) body = <Spin />;
  if (isError) body = <Alert type="error" showIcon message="خطا در دریافت وضعیت صحت لاگ‌ها" />;
  if (data?.last_result === 'broken') {
    const id = data.broken_at_id > 0 ? String(data.broken_at_id) : 'ثبت نشده';
    body = <Alert type="error" showIcon message="خراب" description={`اولین ردیف خراب: ${id}`} />;
  } else if (data?.last_result === 'ok') {
    const when = data.last_check_at ? formatDateTime(data.last_check_at) : 'ثبت نشده';
    body = <Alert type="success" showIcon message="سالم" description={`آخرین بررسی: ${when}`} />;
  }

  return (
    <Card title="وضعیت صحت لاگ‌ها">
      <Flex vertical gap="small" align="stretch">
        {body}
        {data ? (
          <Button type="link" onClick={() => setOpen(true)}>
            {`یافته‌های باز: ${data.open_findings_count ?? 0}`}
          </Button>
        ) : null}
      </Flex>
      <Modal
        title="یافته‌های باز یکپارچگی"
        open={open}
        onCancel={() => setOpen(false)}
        footer={[
          <Button key="close" type="primary" onClick={() => setOpen(false)}>
            بستن
          </Button>,
        ]}
      >
        <FindingList findings={findings} />
      </Modal>
    </Card>
  );
}

/**
 * فهرست یافته‌های باز را نشان می‌دهد.
 * @param findings ردیف‌های باز برگشتی از سرور
 */
function FindingList({ findings }: { findings: OpenIntegrityFinding[] }) {
  if (findings.length === 0) {
    return <Alert type="info" showIcon message="یافته بازی نیست" />;
  }
  return (
    <List
      dataSource={findings}
      renderItem={(item) => (
        <List.Item>
          {`${item.entity_type_label || item.entity_type}، ${item.entity_name}، ${formatDateTime(item.first_seen_at)}`}
        </List.Item>
      )}
    />
  );
}
