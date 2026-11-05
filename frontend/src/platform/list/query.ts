/** یک صفحه از فهرست سمت سرور، هم‌شکل پاسخ listquery. */
export type PageResult<T> = {
  items: T[];
  total: number;
  page: number;
  page_size: number;
};

/** وضعیت جدول که به API فهرست فرستاده می‌شود. فیلترها هنوز تخت نشده‌اند. */
export type ListQuery = {
  page: number;
  page_size: number;
  sort?: string;
  order?: 'asc' | 'desc';
  filters: Record<string, string>;
};

/**
 * پارامترهای جدول را به query مسطح HTTP تبدیل می‌کند.
 * فیلترها با کلید `filter[column]` می‌روند تا با قرارداد listquery سرور یکی باشند.
 * @param query صفحه، اندازه، مرتب‌سازی و فیلترهای پر شده
 * @returns پارامترهای axios؛ فیلتر خالی حذف می‌شود
 */
export function toRequestParams(query: ListQuery): Record<string, string | number> {
  const params: Record<string, string | number> = {
    page: query.page,
    page_size: query.page_size,
  };
  if (query.sort) params.sort = query.sort;
  if (query.order) params.order = query.order;
  for (const key of Object.keys(query.filters).sort()) {
    const value = query.filters[key]?.trim();
    if (value) params[`filter[${key}]`] = value;
  }
  return params;
}
