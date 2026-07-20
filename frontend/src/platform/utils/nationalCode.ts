/**
 * اعتبارسنجی کد ملی ایرانی (۱۰ رقم + رقم کنترل).
 * @param code کد ملی
 * @returns true اگر معتبر باشد
 */
export function isValidIranianNationalCode(code: string): boolean {
  const c = code.trim();
  if (!/^\d{10}$/.test(c)) return false;
  if (/^(\d)\1{9}$/.test(c)) return false;
  let sum = 0;
  for (let i = 0; i < 9; i++) {
    sum += Number(c[i]) * (10 - i);
  }
  const remainder = sum % 11;
  const check = Number(c[9]);
  return remainder < 2 ? check === remainder : check === 11 - remainder;
}
