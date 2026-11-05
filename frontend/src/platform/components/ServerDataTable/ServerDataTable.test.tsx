import { fireEvent, screen, waitFor } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import { jalaliTextToGregorian } from '@/platform/components/JalaliDatePicker/JalaliDatePicker';
import { renderWithProviders } from '@/test/renderWithProviders';
import { ServerDataTable } from './ServerDataTable';
import type { ServerColumn } from './types';
import type { ListQuery, PageResult } from '@/platform/list/query';

type Row = {
  id: number;
  name: string;
  status: string;
  created_at: string;
  amount: number;
};

const columns: ServerColumn<Row>[] = [
  { key: 'name', title: 'نام', filter: 'text', sortable: true },
  {
    key: 'status',
    title: 'وضعیت',
    filter: 'enum',
    sortable: true,
    options: [{ value: 'ok', label: 'سالم' }],
  },
  { key: 'created_at', title: 'زمان', filter: 'jalaliDateRange', sortable: true },
  { key: 'amount', title: 'مبلغ', filter: 'numberRange', sortable: false },
];

const emptyPage: PageResult<Row> = { items: [], total: 0, page: 1, page_size: 20 };

/**
 * جدول را با fetch جعلی رندر می‌کند.
 * @param fetchPage تابعی که پارامترهای ارسالی را ضبط می‌کند
 * @param route آدرس اولیه برای بازگرداندن فیلتر از URL
 */
function renderTable(fetchPage: (query: ListQuery) => Promise<PageResult<Row>>, route = '/') {
  return renderWithProviders(
    <ServerDataTable columns={columns} queryKey={['rows']} fetchPage={fetchPage} rowKey="id" />,
    { route },
  );
}

describe('ServerDataTable', () => {
  it('فیلتر متن را به fetch می‌فرستد', async () => {
    const fetchPage = vi.fn().mockResolvedValue(emptyPage);
    renderTable(fetchPage);

    await waitFor(() => expect(fetchPage).toHaveBeenCalled());
    const header = screen.getByRole('columnheader', { name: 'نام' });
    fireEvent.click(header.querySelector('.ant-table-filter-trigger') as Element);
    fireEvent.change(screen.getByLabelText('مقدار فیلتر نام'), { target: { value: 'علی' } });
    fireEvent.click(screen.getByRole('button', { name: 'اعمال' }));

    await waitFor(() => {
      expect(fetchPage).toHaveBeenLastCalledWith(
        expect.objectContaining({
          page: 1,
          page_size: 20,
          filters: { name: 'علی' },
        }),
      );
    });
  });

  it('مرتب‌سازی ستون را با sort و order به fetch می‌فرستد', async () => {
    const fetchPage = vi.fn().mockResolvedValue(emptyPage);
    renderTable(fetchPage);
    await waitFor(() => expect(fetchPage).toHaveBeenCalled());

    fireEvent.click(screen.getByRole('columnheader', { name: 'نام' }));

    await waitFor(() => {
      expect(fetchPage).toHaveBeenLastCalledWith(
        expect.objectContaining({ sort: 'name', order: 'asc' }),
      );
    });
  });

  it('فیلتر enum و تاریخ و عدد و مرتب‌سازی را از URL به fetch برمی‌گرداند', async () => {
    const gregorian = jalaliTextToGregorian('۱۴۰۵/۰۸/۱۴');
    const fetchPage = vi.fn().mockResolvedValue(emptyPage);
    renderTable(
      fetchPage,
      `/?page=3&page_size=50&sort=name&order=desc&filter[status]=ok&filter[created_at_from]=${gregorian}&filter[created_at_to]=${gregorian}&filter[amount_from]=10&filter[amount_to]=20`,
    );

    await waitFor(() => {
      expect(fetchPage).toHaveBeenCalledWith({
        page: 3,
        page_size: 50,
        sort: 'name',
        order: 'desc',
        filters: {
          status: 'ok',
          created_at_from: '2026-11-05',
          created_at_to: '2026-11-05',
          amount_from: '10',
          amount_to: '20',
        },
      });
    });
  });

  it('حالت خالی و خطا را نشان می‌دهد و پاک‌کردن فیلترها را به fetch بدون فیلتر می‌فرستد', async () => {
    const fetchPage = vi.fn().mockResolvedValue(emptyPage);
    renderTable(fetchPage, '/?filter[name]=علی');
    expect(await screen.findByText('موردی یافت نشد')).toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: 'پاک‌کردن همه فیلترها' }));
    await waitFor(() => {
      expect(fetchPage).toHaveBeenLastCalledWith(
        expect.objectContaining({ filters: {}, page: 1 }),
      );
    });
  });

  it('خطای دریافت را نشان می‌دهد', async () => {
    const fetchPage = vi.fn().mockRejectedValue({ message: 'ارتباط قطع شد' });
    renderTable(fetchPage);
    expect(await screen.findByText('ارتباط قطع شد')).toBeInTheDocument();
    expect(screen.getByText('خطا در دریافت اطلاعات')).toBeInTheDocument();
  });
});
