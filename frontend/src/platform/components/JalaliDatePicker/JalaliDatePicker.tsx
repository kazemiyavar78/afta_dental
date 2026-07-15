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
 * @param value مقدار انتخاب‌شده در DatePicker
 */
export function toGregorianDateString(value: Dayjs | null | undefined): string {
  if (!value || !value.isValid()) return '';
  return value.calendar('gregory').format('YYYY-MM-DD');
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
