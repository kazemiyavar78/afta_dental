import '@testing-library/jest-dom/vitest';
import dayjs from 'dayjs';
import jalaliday from 'jalaliday';

// matcherهای DOM و تقویم شمسی برای همهٔ تست‌های Vitest.
// antd Table در jsdom به matchMedia نیاز دارد.
dayjs.extend(jalaliday);

if (!window.matchMedia) {
  window.matchMedia = (query: string) => ({
    matches: false,
    media: query,
    onchange: null,
    addListener: () => undefined,
    removeListener: () => undefined,
    addEventListener: () => undefined,
    removeEventListener: () => undefined,
    dispatchEvent: () => false,
  });
}

if (!window.ResizeObserver) {
  class ResizeObserverPolyfill {
    observe() {}
    unobserve() {}
    disconnect() {}
  }
  window.ResizeObserver = ResizeObserverPolyfill as typeof ResizeObserver;
}
