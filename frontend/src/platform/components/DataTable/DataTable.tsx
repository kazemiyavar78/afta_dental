import { Table } from 'antd';
import type { TableProps, TablePaginationConfig } from 'antd';
import { useState } from 'react';

type DataTableProps<T extends object> = {
  columns: TableProps<T>['columns'];
  data: T[];
  loading?: boolean;
  rowKey: keyof T | ((record: T) => string);
  pageSize?: number;
  // فهرست‌های لاگ و نشست به ServerDataTable رفته‌اند. بقیه صفحه‌ها تا فاز UI همین جدول کلاینت‌ساید را دارند.
};

/**
 * جدول عمومی با pagination.
 * @example <DataTable columns={cols} data={users} rowKey="id" loading={isLoading} />
 */
export function DataTable<T extends object>({
  columns,
  data,
  loading,
  rowKey,
  pageSize = 10,
}: DataTableProps<T>) {
  const [pagination, setPagination] = useState<TablePaginationConfig>({
    current: 1,
    pageSize,
    showSizeChanger: true,
    showTotal: (total) => `مجموع ${total} رکورد`,
  });

  return (
    <Table<T>
      columns={columns}
      dataSource={data}
      loading={loading}
      rowKey={rowKey as string | ((record: T) => string)}
      pagination={pagination}
      onChange={(pag) => setPagination((prev) => ({ ...prev, ...pag }))}
      scroll={{ x: true }}
    />
  );
}
