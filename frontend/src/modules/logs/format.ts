/**
 * نام کوتاه مرورگر را از User-Agent برمی‌گرداند تا ستون جدول شلوغ نشود.
 * @param userAgent رشته ذخیره‌شده در لاگ
 * @returns نام مرورگر، یا خود متن اگر شناخته نشود، یا خط تیره اگر خالی باشد
 */
export function summarizeUserAgent(userAgent: string | null | undefined): string {
  const value = userAgent?.trim() ?? '';
  if (!value) return '—';
  const known: Array<[RegExp, string]> = [
    [/Edg\//, 'Edge'],
    [/OPR\/|Opera/, 'Opera'],
    [/Chrome\//, 'Chrome'],
    [/Firefox\//, 'Firefox'],
    [/Safari\//, 'Safari'],
  ];
  for (const [pattern, name] of known) {
    if (pattern.test(value)) return name;
  }
  if (value.length > 48) return `${value.slice(0, 48)}…`;
  return value;
}
