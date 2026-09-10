import { describe, expect, it } from 'vitest';
import { formatCellValue, getRowValue } from './value';

describe('value formatters', () => {
  it('formatCellValue handles currency', () => {
    expect(formatCellValue(1500, 'currency')).toBeTruthy();
  });

  it('formatCellValue returns dash for empty', () => {
    expect(formatCellValue(null, 'text')).toBe('—');
  });

  it('getRowValue reads row field', () => {
    expect(getRowValue({ name: 'test' }, 'name')).toBe('test');
  });
});
