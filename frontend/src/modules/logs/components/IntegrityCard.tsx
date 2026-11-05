import { Alert, Card, Spin } from 'antd';
import { formatDateTime } from '@/platform/reports/formatters/date';
import { useLogIntegrity } from '../hooks';

/**
 * کارت وضعیت صحت لاگ‌ها را از نتیجه آخرین بررسی زنجیره نشان می‌دهد.
 * سالم سبز است و زمان بررسی را دارد. خراب قرمز است و شناسه اولین ردیف خراب را دارد.
 */
export function IntegrityCard() {
  const { data, isLoading, isError } = useLogIntegrity();

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

  return <Card title="وضعیت صحت لاگ‌ها">{body}</Card>;
}
