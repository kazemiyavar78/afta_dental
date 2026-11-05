import { useMemo } from 'react';
import { Flex, Tag } from 'antd';
import { PageHeader } from '@/platform/components/PageHeader';
import { ServerDataTable, type ServerColumn } from '@/platform/components/ServerDataTable';
import { formatDateTime } from '@/platform/reports/formatters/date';
import { fetchLogs } from '../api';
import { LogDetails } from '../components/LogDetails';
import { IntegrityCard } from '../components/IntegrityCard';
import { summarizeUserAgent } from '../format';
import { useLogMeta } from '../hooks';
import type { LogOption, SecurityLog } from '../types';

const RESULT_OPTIONS = [
  { value: 'success', label: 'موفق' },
  { value: 'failure', label: 'ناموفق' },
];

/**
 * صفحه لاگ امنیتی. هر ستون فیلتر و مرتب‌سازی سمت سرور دارد و شناسه عددی کاربر نشان داده نمی‌شود.
 */
export function LogsPage() {
  const { data: meta } = useLogMeta();
  const columns = useMemo(() => logColumns(meta?.event_types ?? [], meta?.reasons ?? [], meta?.target_types ?? []), [meta]);

  return (
    <Flex vertical gap="large" align="stretch">
      <PageHeader
        title="لاگ‌های امنیتی"
        subtitle="مشاهده و جستجوی رویدادهای امنیتی. فیلتر و مرتب‌سازی سمت سرور است."
      />
      <IntegrityCard />
      <ServerDataTable
        columns={columns}
        queryKey={['logs']}
        fetchPage={fetchLogs}
        rowKey="id"
        expandable={(record) => <LogDetails details={record.details} />}
      />
    </Flex>
  );
}

/**
 * ستون‌های صفحه لاگ را به ترتیب نمایش می‌سازد.
 * @param eventTypes گزینه‌های نوع رویداد
 * @param reasons گزینه‌های دلیل
 * @param targetTypes برچسب نوع هدف وقتی نام هدف خالی است
 */
function logColumns(eventTypes: LogOption[], reasons: LogOption[], targetTypes: LogOption[]): ServerColumn<SecurityLog>[] {
  return [
    {
      key: 'created_at',
      title: 'زمان',
      filter: 'jalaliDateRange',
      sortable: true,
      width: 170,
      render: (value) => formatDateTime(typeof value === 'string' ? value : null),
    },
    {
      key: 'actor_username',
      title: 'نام کاربری انجام‌دهنده',
      filter: 'text',
      sortable: true,
      render: (value) => textOrDash(value),
    },
    {
      key: 'event_type',
      title: 'نوع رویداد',
      filter: 'enum',
      sortable: true,
      options: eventTypes,
      render: (_value, record) => record.event_type_label || '—',
    },
    {
      key: 'result',
      title: 'نتیجه',
      filter: 'enum',
      sortable: true,
      width: 110,
      options: RESULT_OPTIONS,
      render: (value) => <ResultTag result={typeof value === 'string' ? value : ''} />,
    },
    {
      key: 'target_name',
      title: 'کاربر یا موجودیت هدف',
      filter: 'text',
      sortable: true,
      render: (_value, record) => targetText(record, targetTypes),
    },
    {
      key: 'reason',
      title: 'دلیل',
      filter: 'enum',
      sortable: true,
      options: reasons,
      render: (_value, record) => record.reason_label || '—',
    },
    { key: 'ip', title: 'IP', filter: 'text', sortable: true, width: 140 },
    {
      key: 'user_agent',
      title: 'مرورگر',
      filter: 'text',
      sortable: true,
      ellipsis: true,
      render: (value) => summarizeUserAgent(typeof value === 'string' ? value : ''),
    },
    {
      key: 'description',
      title: 'توضیحات',
      filter: 'text',
      sortable: true,
      ellipsis: true,
      render: (value) => textOrDash(value),
    },
  ];
}

/**
 * نتیجه رویداد را با برچسب رنگی موضوع نشان می‌دهد.
 * @param result کد success یا failure
 */
function ResultTag({ result }: { result: string }) {
  if (result === 'success') return <Tag color="success">موفق</Tag>;
  if (result === 'failure') return <Tag color="error">ناموفق</Tag>;
  return <>—</>;
}

/**
 * نام هدف را نشان می‌دهد و اگر خالی بود برچسب فارسی نوع هدف را می‌گذارد.
 * @param record ردیف لاگ
 * @param targetTypes فهرست برچسب نوع هدف
 */
function targetText(record: SecurityLog, targetTypes: LogOption[]): string {
  if (record.target_name?.trim()) return record.target_name;
  const label = targetTypes.find((item) => item.value === record.target_type)?.label;
  return label || '—';
}

/**
 * متن خالی ستون را به خط تیره تبدیل می‌کند.
 * @param value مقدار سلول
 */
function textOrDash(value: unknown): string {
  if (typeof value !== 'string' || !value.trim()) return '—';
  return value;
}
