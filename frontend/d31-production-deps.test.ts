import { readFileSync } from 'node:fs';
import { describe, expect, it } from 'vitest';

type LockFile = {
  packages: Record<string, { version?: string }>;
};

// lockedVersion نسخه قفل‌شده یک بسته تولید را از lockfile برمی‌گرداند.
// ورودی نام بسته npm است و خروجی فیلد version در node_modules همان بسته، یا undefined اگر نباشد.
// D-31 / AVA_VAN.1: این نسخه باید همان اصلاح ممیزی high باشد. xlsx تا فاز ۵ همان 0.18.5 می‌ماند (PP-18).
function lockedVersion(lock: LockFile, name: string): string | undefined {
  return lock.packages[`node_modules/${name}`]?.version;
}

describe('D-31 production dependency lock', () => {
  it('pins patched axios, follow-redirects, react-router and zod, and leaves xlsx', () => {
    const lock = JSON.parse(readFileSync('package-lock.json', 'utf8')) as LockFile;
    expect(lockedVersion(lock, 'axios')).toBe('1.21.0');
    expect(lockedVersion(lock, 'follow-redirects')).toBe('1.16.2');
    expect(lockedVersion(lock, 'react-router')).toBe('7.18.5');
    expect(lockedVersion(lock, 'react-router-dom')).toBe('7.18.5');
    expect(lockedVersion(lock, 'zod')).toBe('4.6.7');
    expect(lockedVersion(lock, 'xlsx')).toBe('0.18.5');
  });
});
