/** جهت دندان مطابق بک‌اند پذیرش */
export const TeethDirection = {
  UpperLeft: 1,
  UpperRight: 2,
  LowerLeft: 3,
  LowerRight: 4,
} as const;

export type TeethDirectionValue = (typeof TeethDirection)[keyof typeof TeethDirection];

export type ToothQuadrantSelection = {
  /** شماره داخل ربع (۱ = سانترال … ۸ = عقل) */
  teeth_number: number;
  /** جهت ربع */
  teeth_direction: TeethDirectionValue;
};

/**
 * تبدیل شماره یونیورسال SVG (۱–۳۲) به جهت و شماره ۱–۸ داخل ربع.
 * ورودی: شماره یونیورسال؛ خروجی: انتخاب ربعی یا null اگر خارج از بازه باشد.
 */
export function universalToQuadrant(universal: number): ToothQuadrantSelection | null {
  if (universal >= 1 && universal <= 8) {
    return { teeth_direction: TeethDirection.UpperRight, teeth_number: 9 - universal };
  }
  if (universal >= 9 && universal <= 16) {
    return { teeth_direction: TeethDirection.UpperLeft, teeth_number: universal - 8 };
  }
  if (universal >= 17 && universal <= 24) {
    return { teeth_direction: TeethDirection.LowerLeft, teeth_number: 25 - universal };
  }
  if (universal >= 25 && universal <= 32) {
    return { teeth_direction: TeethDirection.LowerRight, teeth_number: universal - 24 };
  }
  return null;
}

/**
 * تبدیل جهت و شماره ۱–۸ به شماره یونیورسال SVG (۱–۳۲).
 * ورودی: جهت و شماره ربع؛ خروجی: شناسه یونیورسال یا null.
 */
export function quadrantToUniversal(
  teeth_direction: number,
  teeth_number: number,
): number | null {
  if (teeth_number < 1 || teeth_number > 8) return null;
  switch (teeth_direction) {
    case TeethDirection.UpperRight:
      return 9 - teeth_number;
    case TeethDirection.UpperLeft:
      return 8 + teeth_number;
    case TeethDirection.LowerLeft:
      return 25 - teeth_number;
    case TeethDirection.LowerRight:
      return 24 + teeth_number;
    default:
      return null;
  }
}

/**
 * شماره نمایشی روی برچسب برای هر دندان یونیورسال (۱–۸ در هر ربع).
 * ورودی: شناسه یونیورسال؛ خروجی: متن برچسب یا null.
 */
export function universalDisplayNumber(universal: number): number | null {
  return universalToQuadrant(universal)?.teeth_number ?? null;
}
