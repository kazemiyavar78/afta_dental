import { describe, expect, it } from 'vitest';
import { formatCurrency, formatNumber, formatPercent } from './number';

describe('number formatters', () => {
  it('formatNumber uses Persian locale', () => {
    expect(formatNumber(1234567)).toMatch(/۱|1/);
  });

  it('formatCurrency formats money', () => {
    expect(formatCurrency(1000)).toBeTruthy();
  });

  it('formatPercent adds percent sign', () => {
    expect(formatPercent(12.5)).toContain('٪');
  });

  it('handles null/NaN', () => {
    expect(formatNumber(null)).toBe('۰');
    expect(formatNumber('abc')).toBe('—');
  });
});
