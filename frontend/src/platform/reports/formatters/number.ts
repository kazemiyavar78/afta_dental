/** عدد را با جداکننده فارسی نمایش می‌دهد */
export function formatNumber(value: unknown, fractionDigits?: number): string {
  const num = Number(value ?? 0);
  if (Number.isNaN(num)) return '—';
  if (fractionDigits != null) {
    return num.toLocaleString('fa-IR', {
      minimumFractionDigits: fractionDigits,
      maximumFractionDigits: fractionDigits,
    });
  }
  return num.toLocaleString('fa-IR');
}

/** alias برای formatNumber — مطابق convention پروژه */
export function formatPersianNumber(value: unknown): string {
  return formatNumber(value);
}

/** مبلغ (ریال) را با جداکننده فارسی نمایش می‌دهد */
export function formatCurrency(value: unknown): string {
  return formatNumber(value);
}

/** درصد را نمایش می‌دهد */
export function formatPercent(value: unknown, fractionDigits = 1): string {
  const num = Number(value ?? 0);
  if (Number.isNaN(num)) return '—';
  return `${num.toLocaleString('fa-IR', {
    minimumFractionDigits: fractionDigits,
    maximumFractionDigits: fractionDigits,
  })}٪`;
}
