import { Table } from 'antd';
import type { ColumnsType } from 'antd/es/table';
import { useMemo, useState } from 'react';
import type { ReportColumn } from '../types';
import { formatColumnValue } from '../formatters/value';

type ReportTableProps<T extends object> = {
  columns: ReportColumn<T>[];
  data: T[];
  loading?: boolean;
  rowKey: keyof T | ((row: T) => string);
  pageSize?: number;
};

/**
 * جدول گزارش بر پایه Ant Design — از ReportDefinition columns استفاده می‌کند.
 */
export function ReportTable<T extends object>({
  columns,
  data,
  loading,
  rowKey,
  pageSize = 10,
}: ReportTableProps<T>) {
  const [pagination, setPagination] = useState({
    current: 1,
    pageSize,
    showSizeChanger: true,
    showTotal: (total: number) => `مجموع ${total.toLocaleString('fa-IR')} رکورد`,
  });

  const antColumns: ColumnsType<T> = useMemo(
    () =>
      columns
        .filter((col) => !col.hidden)
        .map((col) => ({
          title: col.title,
          dataIndex: col.key as string,
          key: String(col.key),
          width: col.width,
          align: col.align ?? 'center',
          ellipsis: col.format === 'text',
          render: (_value: unknown, record: T, index: number) => {
            if (col.formatter) {
              return col.formatter(_value, record, index);
            }
            return formatColumnValue(col, record, index, 'ui');
          },
        })),
    [columns],
  );

  return (
    <Table<T>
      columns={antColumns}
      dataSource={data}
      loading={loading}
      rowKey={rowKey as string | ((record: T) => string)}
      pagination={pagination}
      onChange={(pag) =>
        setPagination((prev) => ({
          ...prev,
          current: pag.current ?? 1,
          pageSize: pag.pageSize ?? prev.pageSize,
        }))
      }
      scroll={{ x: true }}
    />
  );
}
