/** وضعیت صفحه‌بندی، مرتب‌سازی و فیلتر که در query string می‌ماند. */
export type TableState = {
  page: number;
  pageSize: number;
  sort?: string;
  order?: 'asc' | 'desc';
  filters: Record<string, string>;
};

const DEFAULT_PAGE = 1;
const DEFAULT_PAGE_SIZE = 20;
const MAX_PAGE_SIZE = 100;
const FILTER_KEY = /^filter\[([A-Za-z_][A-Za-z0-9_]*)\]$/;

/**
 * شماره صحیح مثبت را می‌خواند و در صورت نامعتبر بودن مقدار پیش‌فرض را برمی‌گرداند.
 * @param raw متن query
 * @param fallback مقدار وقتی خالی یا غیرعددی است
 */
function positiveInt(raw: string | null, fallback: number): number {
  if (!raw?.trim()) return fallback;
  const parsed = Number(raw);
  if (!Number.isInteger(parsed) || parsed < 1) return fallback;
  return parsed;
}

/**
 * وضعیت جدول را از URL می‌خواند تا بعد از تازه‌کردن صفحه همان فیلتر بماند.
 * @param params query جاری مرورگر
 * @returns صفحه، اندازه، مرتب‌سازی و فیلترهای `filter[...]`
 */
export function readTableState(params: URLSearchParams): TableState {
  const page = positiveInt(params.get('page'), DEFAULT_PAGE);
  const pageSize = Math.min(positiveInt(params.get('page_size'), DEFAULT_PAGE_SIZE), MAX_PAGE_SIZE);
  const sort = params.get('sort')?.trim() || undefined;
  const orderRaw = params.get('order');
  const order = sort && (orderRaw === 'asc' || orderRaw === 'desc') ? orderRaw : undefined;
  const filters: Record<string, string> = {};
  params.forEach((value, key) => {
    const match = FILTER_KEY.exec(key);
    if (match && value.trim()) filters[match[1]] = value;
  });
  return { page, pageSize, sort, order, filters };
}

/**
 * وضعیت جدول را در query string می‌نویسد و پارامترهای بی‌ربط صفحه را نگه می‌دارد.
 * صفحه ۱ و اندازه ۲۰ نوشته نمی‌شوند چون پیش‌فرض سرور همان است.
 * @param current query قبلی
 * @param state وضعیت جدید جدول
 */
export function writeTableState(current: URLSearchParams, state: TableState): URLSearchParams {
  const next = new URLSearchParams(current);
  for (const key of [...next.keys()]) {
    if (key === 'page' || key === 'page_size' || key === 'sort' || key === 'order' || key.startsWith('filter[')) {
      next.delete(key);
    }
  }
  if (state.page > DEFAULT_PAGE) next.set('page', String(state.page));
  if (state.pageSize !== DEFAULT_PAGE_SIZE) next.set('page_size', String(state.pageSize));
  if (state.sort) next.set('sort', state.sort);
  if (state.order) next.set('order', state.order);
  for (const key of Object.keys(state.filters).sort()) {
    const value = state.filters[key]?.trim();
    if (value) next.set(`filter[${key}]`, value);
  }
  return next;
}

/**
 * می‌گوید حداقل یک فیلتر در وضعیت جدول هست یا نه.
 * @param state وضعیت خوانده‌شده از URL
 */
export function hasActiveFilters(state: TableState): boolean {
  return Object.values(state.filters).some((value) => value.trim() !== '');
}
