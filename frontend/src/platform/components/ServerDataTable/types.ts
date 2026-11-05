import type { ReactNode } from 'react';
import type { ListQuery, PageResult } from '@/platform/list/query';

/** نوع فیلتر سرستون. `none` برای ستونی است که جستجو نباید داشته باشد. */
export type ColumnFilterKind = 'text' | 'enum' | 'jalaliDateRange' | 'numberRange' | 'none';

/** یک گزینه فیلتر انتخابی. value به سرور می‌رود و label فقط نمایش است. */
export type ColumnOption = {
  value: string;
  label: string;
};

/** تعریف یک ستون جدول سمت سرور. */
export type ServerColumn<T> = {
  key: string;
  title: string;
  filter: ColumnFilterKind;
  sortable: boolean;
  options?: ColumnOption[];
  width?: number;
  ellipsis?: boolean;
  render?: (value: unknown, record: T) => ReactNode;
};

/** ورودی جدول مشترک. ستون عملیات جداست تا همیشه آخر بماند. */
export type ServerDataTableProps<T extends object> = {
  columns: ServerColumn<T>[];
  queryKey: readonly unknown[];
  fetchPage: (query: ListQuery) => Promise<PageResult<T>>;
  rowKey: keyof T | ((record: T) => string);
  actions?: {
    render: (record: T) => ReactNode;
    width?: number;
  };
  expandable?: (record: T) => ReactNode;
};
