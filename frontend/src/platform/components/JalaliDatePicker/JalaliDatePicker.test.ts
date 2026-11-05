import { describe, expect, it } from 'vitest';
import dayjs from 'dayjs';
import { jalaliTextToGregorian, toGregorianDateString } from './JalaliDatePicker';

describe('jalaliTextToGregorian', () => {
  it('D-16 تاریخ ۱۴۰۵/۰۸/۱۴ را به 2026-11-05 تبدیل می‌کند', () => {
    const jalali = '۱۴۰۵/۰۸/۱۴';
    const parsed = dayjs('1405/08/14', { jalali: true } as { jalali: boolean });
    expect(parsed.calendar('gregory').format('YYYY-MM-DD')).toBe('2026-11-05');
    expect(toGregorianDateString(parsed)).toBe('2026-11-05');
    expect(jalaliTextToGregorian(jalali)).toBe('2026-11-05');
  });
});
