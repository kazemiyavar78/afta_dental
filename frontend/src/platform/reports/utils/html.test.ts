import { describe, expect, it } from 'vitest';
import { buildPageCss, escapeHtml, resolvePageConfig } from './html';

describe('html utils', () => {
  it('escapeHtml escapes special chars', () => {
    expect(escapeHtml('<script>')).toBe('&lt;script&gt;');
  });

  it('buildPageCss generates @page rule', () => {
    const css = buildPageCss({ size: 'A4', orientation: 'portrait', margin: 12 });
    expect(css).toContain('@page');
    expect(css).toContain('A4 portrait');
  });

  it('resolvePageConfig merges defaults', () => {
    const config = resolvePageConfig({ size: 'A5', orientation: 'landscape', margin: 8 });
    expect(config.size).toBe('A5');
    expect(config.orientation).toBe('landscape');
  });
});
