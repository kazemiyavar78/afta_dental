import { useMemo } from 'react';
import { Alert, Button, Flex, Table } from 'antd';
import { FilterOutlined } from '@ant-design/icons';
import { keepPreviousData, useQuery } from '@tanstack/react-query';
import type { ColumnsType, TablePaginationConfig } from 'antd/es/table';
import type { FilterValue, SorterResult, TableCurrentDataSource } from 'antd/es/table/interface';
import { useSearchParams } from 'react-router-dom';
import type { ListQuery } from '@/platform/list/query';
import { FilterDropdown, filterKeysFor } from './FilterDropdown';
import { hasActiveFilters, readTableState, writeTableState, type TableState } from './tableState';
import type { ServerColumn, ServerDataTableProps } from './types';

/**
 * جدول صفحه‌بندی‌شده سمت سرور. فیلتر هر ستون، مرتب‌سازی و صفحه در URL می‌ماند.
 * ستون عملیات، اگر داده شود، همیشه بعد از ستون‌های داده رندر می‌شود.
 */
export function ServerDataTable<T extends object>({
  columns,
  queryKey,
  fetchPage,
  rowKey,
  actions,
  expandable,
}: ServerDataTableProps<T>) {
  const [searchParams, setSearchParams] = useSearchParams();
  const tableState = useMemo(() => readTableState(searchParams), [searchParams]);
  const listQuery = useMemo(() => toListQuery(tableState), [tableState]);

  const query = useQuery({
    queryKey: [...queryKey, listQuery],
    queryFn: () => fetchPage(listQuery),
    placeholderData: keepPreviousData,
  });

  const commit = (next: TableState) => {
    setSearchParams(writeTableState(searchParams, next), { replace: true });
  };

  const applyFilter = (patch: Record<string, string>) => {
    const filters = { ...tableState.filters };
    for (const [key, value] of Object.entries(patch)) {
      if (!value.trim()) delete filters[key];
      else filters[key] = value.trim();
    }
    commit({ ...tableState, page: 1, filters });
  };

  const onTableChange = (
    pagination: TablePaginationConfig,
    _filters: Record<string, FilterValue | null>,
    sorter: SorterResult<T> | SorterResult<T>[],
    extra: TableCurrentDataSource<T>,
  ) => {
    const pageSize = pagination.pageSize ?? tableState.pageSize;
    let next: TableState = {
      ...tableState,
      page: pageSize !== tableState.pageSize ? 1 : (pagination.current ?? 1),
      pageSize,
    };
    if (extra.action === 'sort') {
      const single = Array.isArray(sorter) ? sorter[0] : sorter;
      if (single?.order && single.columnKey != null) {
        next = {
          ...next,
          sort: String(single.columnKey),
          order: single.order === 'descend' ? 'desc' : 'asc',
        };
      } else {
        next = { ...next, sort: undefined, order: undefined };
      }
    }
    commit(next);
  };

  const errorMessage = query.isError ? messageOf(query.error) : '';
  const items = query.data?.items ?? [];
  const emptyText = query.isError ? 'خطا در دریافت اطلاعات' : 'موردی یافت نشد';

  return (
    <Flex vertical gap="middle" align="stretch">
      <Flex justify="end">
        <Button
          htmlType="button"
          disabled={!hasActiveFilters(tableState)}
          onClick={() => commit({ ...tableState, page: 1, filters: {} })}
        >
          پاک‌کردن همه فیلترها
        </Button>
      </Flex>
      {query.isError && (
        <Alert
          type="error"
          showIcon
          message={errorMessage}
          action={
            <Button size="small" htmlType="button" onClick={() => query.refetch()}>
              تلاش دوباره
            </Button>
          }
        />
      )}
      <Table<T>
        columns={buildColumns(columns, tableState, applyFilter, actions)}
        dataSource={query.isError && !query.data ? [] : items}
        loading={query.isFetching}
        rowKey={rowKey as string | ((record: T) => string)}
        pagination={{
          current: tableState.page,
          pageSize: tableState.pageSize,
          total: query.data?.total ?? 0,
          showSizeChanger: true,
          pageSizeOptions: [10, 20, 50, 100],
          showTotal: (total) => `مجموع ${total} رکورد`,
        }}
        locale={{ emptyText }}
        scroll={{ x: true }}
        onChange={onTableChange}
        expandable={expandable ? { expandedRowRender: expandable } : undefined}
      />
    </Flex>
  );
}

/**
 * وضعیت URL را به ورودی تابع fetch صفحه تبدیل می‌کند.
 * @param state وضعیت خوانده‌شده از query string
 */
function toListQuery(state: TableState): ListQuery {
  return {
    page: state.page,
    page_size: state.pageSize,
    sort: state.sort,
    order: state.order,
    filters: state.filters,
  };
}

/**
 * پیام خطای React Query را برای Alert برمی‌گرداند. متن استک نشان داده نمی‌شود.
 * @param error مقدار پرتاب‌شده از fetch
 */
function messageOf(error: unknown): string {
  if (error && typeof error === 'object' && 'message' in error && typeof error.message === 'string' && error.message) {
    return error.message;
  }
  return 'خطا در دریافت اطلاعات';
}

/**
 * ستون‌های antd را از تعریف جدول می‌سازد. فیلتر و مرتب‌سازی سمت کلاینت اعمال نمی‌شود.
 * ستون عملیات بعد از بقیه اضافه می‌شود تا در RTL سمت چپ بماند.
 */
function buildColumns<T extends object>(
  columns: ServerColumn<T>[],
  state: TableState,
  applyFilter: (patch: Record<string, string>) => void,
  actions: ServerDataTableProps<T>['actions'],
): ColumnsType<T> {
  const dataColumns: ColumnsType<T> = columns.map((column) => {
    const active = filterKeysFor(column).some((key) => Boolean(state.filters[key]?.trim()));
    return {
      title: column.title,
      dataIndex: column.key,
      key: column.key,
      width: column.width,
      ellipsis: column.ellipsis,
      render: column.render,
      sorter: column.sortable || undefined,
      sortOrder: sortOrderOf(column, state),
      showSorterTooltip: false,
      filtered: active,
      filterIcon: column.filter === 'none' ? undefined : <FilterOutlined />,
      filterDropdown:
        column.filter === 'none'
          ? undefined
          : ({ confirm, visible }) => (
              <FilterDropdown
                column={column}
                filters={state.filters}
                visible={visible}
                onApply={applyFilter}
                onClose={() => confirm({ closeDropdown: true })}
              />
            ),
    };
  });
  if (!actions) return dataColumns;
  return [
    ...dataColumns,
    {
      title: 'عملیات',
      key: 'actions',
      width: actions.width ?? 96,
      render: (_value, record) => actions.render(record),
    },
  ];
}

/**
 * جهت فلش مرتب‌سازی سرستون را از URL می‌خواند.
 * @param column ستون داده
 * @param state وضعیت جدول
 */
function sortOrderOf<T>(column: ServerColumn<T>, state: TableState): 'ascend' | 'descend' | null {
  if (!column.sortable || state.sort !== column.key || !state.order) return null;
  return state.order === 'desc' ? 'descend' : 'ascend';
}
