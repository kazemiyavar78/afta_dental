import { DatePicker } from 'antd';
import type { DatePickerProps } from 'antd';
import dayjs, { type Dayjs } from 'dayjs';

type JalaliDatePickerProps = Omit<DatePickerProps, 'value' | 'onChange' | 'picker'> & {
  /** مقدار ذخیره‌شده به صورت میلادی `YYYY-MM-DD` */
  value?: string | null;
  /** خروجی همیشه میلادی `YYYY-MM-DD` برای ارسال به بک‌اند */
  onChange?: (value: string) => void;
};

/**
 * رشته میلادی `YYYY-MM-DD` را به dayjs با نمایش شمسی تبدیل می‌کند.
 * @param value تاریخ میلادی از API/فرم
 */
export function fromGregorianDateString(value?: string | null): Dayjs | null {
  if (!value) return null;
  const parsed = dayjs(`${value}T00:00:00`);
  if (!parsed.isValid()) return null;
  return parsed.calendar('jalali');
}

/**
 * dayjs انتخاب‌شده را به رشته میلادی `YYYY-MM-DD` برای بک‌اند تبدیل می‌کند.
 * تبدیل با تقویم میلادی صریح است تا تاریخ شمسیِ نمایش به سرور نرود.
 * @param value مقدار انتخاب‌شده در DatePicker
 * @returns تاریخ میلادی یا رشته خالی اگر مقدار نامعتبر باشد
 */
export function toGregorianDateString(value: Dayjs | null | undefined): string {
  if (!value || !value.isValid()) return '';
  return value.calendar('gregory').format('YYYY-MM-DD');
}

const PERSIAN_DIGITS = '۰۱۲۳۴۵۶۷۸۹';
const ARABIC_DIGITS = '٠١٢٣٤٥٦٧٨٩';

/**
 * رقم فارسی و عربی را به انگلیسی برمی‌گرداند تا تاریخ قابل خواندن باشد.
 * @param value متن تاریخ یا عدد
 * @returns همان متن با رقم انگلیسی
 */
export function toEnglishDigits(value: string): string {
  return value.replace(/[۰-۹٠-٩]/g, (digit) => {
    const persian = PERSIAN_DIGITS.indexOf(digit);
    if (persian >= 0) return String(persian);
    return String(ARABIC_DIGITS.indexOf(digit));
  });
}

/**
 * متن تاریخ شمسی را قبل از ارسال به میلادی `YYYY-MM-DD` تبدیل می‌کند.
 * D-16: سرور فقط تاریخ میلادی می‌خواند.
 * @param value تاریخ شمسی مثل ۱۴۰۵/۰۸/۱۴
 * @returns تاریخ میلادی، یا رشته خالی اگر قابل خواندن نباشد
 */
export function jalaliTextToGregorian(value: string): string {
  const normalized = toEnglishDigits(value).trim().replace(/-/g, '/');
  if (!normalized) return '';
  const parsed = dayjs(normalized, { jalali: true } as { jalali: boolean });
  return toGregorianDateString(parsed);
}

/**
 * DatePicker شمسی که value/onChange آن همیشه میلادی `YYYY-MM-DD` است.
 */
export function JalaliDatePicker({ value, onChange, format = 'YYYY/MM/DD', ...rest }: JalaliDatePickerProps) {
  return (
    <DatePicker
      {...rest}
      format={format}
      value={fromGregorianDateString(value)}
      onChange={(date) => {
        const selected = Array.isArray(date) ? date[0] : date;
        onChange?.(toGregorianDateString(selected ?? null));
      }}
    />
  );
}
