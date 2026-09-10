import { useEffect, type RefObject } from 'react';

/**
 * جلوگیری از scrollIntoView هنگام فوکوس روی فیلدهای داخل یک ظرف — رفع لرزش اسکرول.
 * @param containerRef ref عنصر ریشه (مثلاً reception-workspace)
 */
export function usePreventFocusScroll(containerRef: RefObject<HTMLElement | null>) {
  useEffect(() => {
    const container = containerRef.current;
    if (!container) return;

    const originalScrollIntoView = HTMLElement.prototype.scrollIntoView;
    HTMLElement.prototype.scrollIntoView = function scrollIntoViewPatched(
      this: HTMLElement,
      arg?: boolean | ScrollIntoViewOptions,
    ) {
      if (container.contains(this)) {
        return;
      }
      return originalScrollIntoView.call(this, arg);
    };

    return () => {
      HTMLElement.prototype.scrollIntoView = originalScrollIntoView;
    };
  }, [containerRef]);
}
