import { Button, Tooltip } from 'antd';
import { DeleteOutlined } from '@ant-design/icons';
import { useQueryClient } from '@tanstack/react-query';
import { ServerDataTable, type ServerColumn } from '@/platform/components/ServerDataTable';
import { confirmDialog } from '@/platform/components/ConfirmDialog';
import { useApiMutation } from '@/platform/hooks/useApiMutation';
import { formatDateTime } from '@/platform/reports/formatters/date';
import type { ListQuery, PageResult } from '@/platform/list/query';
import { deleteSession } from '../api';
import type { Session } from '../types';

type SessionTableProps = {
  queryKey: readonly unknown[];
  fetchPage: (query: ListQuery) => Promise<PageResult<Session>>;
};

const columns: ServerColumn<Session>[] = [
  {
    key: 'username',
    title: 'نام کاربری',
    filter: 'text',
    sortable: true,
    render: (value) => (typeof value === 'string' && value.trim() ? value : '—'),
  },
  { key: 'ip', title: 'آدرس IP', filter: 'text', sortable: true },
  { key: 'browser', title: 'مرورگر', filter: 'text', sortable: true, ellipsis: true },
  {
    key: 'creation_time',
    title: 'زمان ورود',
    filter: 'jalaliDateRange',
    sortable: true,
    width: 170,
    render: (value) => formatDateTime(typeof value === 'string' ? value : null),
  },
];

/**
 * جدول نشست با فیلتر و مرتب‌سازی سمت سرور برای هر ستون.
 * شناسه عددی کاربر ستون نیست؛ محدوده کاربر از بیرون در fetchPage گذاشته می‌شود.
 * @param props کلید query و تابع دریافت صفحه
 */
export function SessionTable({ queryKey, fetchPage }: SessionTableProps) {
  const queryClient = useQueryClient();
  const deleteSessionMutation = useApiMutation({
    mutationFn: deleteSession,
    successMessage: 'نشست با موفقیت حذف شد',
    onSuccess: () => queryClient.invalidateQueries({ queryKey: [...queryKey] }),
  });

  return (
    <ServerDataTable
      columns={columns}
      queryKey={queryKey}
      fetchPage={fetchPage}
      rowKey="id"
      actions={{
        render: (record) => (
          <Tooltip title="حذف">
            <Button
              type="text"
              danger
              htmlType="button"
              aria-label="حذف"
              icon={<DeleteOutlined />}
              loading={deleteSessionMutation.isPending}
              onClick={() =>
                confirmDialog({
                  title: 'حذف نشست',
                  content: 'آیا از حذف این نشست مطمئن هستید؟',
                  okType: 'danger',
                  onConfirm: () => deleteSessionMutation.mutateAsync(record.id),
                })
              }
            />
          </Tooltip>
        ),
      }}
    />
  );
}
