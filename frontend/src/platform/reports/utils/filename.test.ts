import { describe, expect, it } from 'vitest';
import { buildReportFilename, buildSheetName } from './filename';

describe('filename utils', () => {
  it('buildReportFilename includes extension', () => {
    const name = buildReportFilename('revenue', 'pdf', 'گزارش درآمد');
    expect(name.endsWith('.pdf')).toBe(true);
  });

  it('buildSheetName truncates long names', () => {
    const long = 'گزارش'.repeat(20);
    expect(buildSheetName(long).length).toBeLessThanOrEqual(31);
  });

  it('buildSheetName removes invalid chars', () => {
    expect(buildSheetName('test/sheet*name')).not.toContain('/');
  });
});
