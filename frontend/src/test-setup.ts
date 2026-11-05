import '@testing-library/jest-dom/vitest';
import dayjs from 'dayjs';
import jalaliday from 'jalaliday';

// matcherهای DOM (مثل toBeInTheDocument) و تقویم شمسی برای همهٔ تست‌های Vitest.
// vitest.config.ts همین فایل را به‌عنوان setup دارد؛ فایل setup جدا ساخته نشده است.
dayjs.extend(jalaliday);
