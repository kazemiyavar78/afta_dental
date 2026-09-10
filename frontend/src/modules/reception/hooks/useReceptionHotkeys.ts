import { useEffect, useRef, type KeyboardEvent as ReactKeyboardEvent } from 'react';

type ReceptionHotkeyHandlers = {
  /** true وقتی مودال باز است و میانبر باید غیرفعال شود */
  disabled?: boolean;
  /** F8 — پذیرش ثبت‌شده: پرینت | پذیرش جدید: ذخیره+پرینت */
  onF8Print: () => void | Promise<void>;
  onAddService: () => void;
  onSaveAndPrint: () => void | Promise<void>;
  /** رفتن به آخرین پذیرش — Esc */
  onLast: () => void;
};

/** تشخیص کلیدهای F1 تا F12 */
function matchFunctionKey(e: KeyboardEvent, fn: number): boolean {
  return e.code === `F${fn}` || e.key === `F${fn}`;
}

/** Select یا DatePicker باز است — Esc باید ابتدا آن‌ها را ببندد */
function isOverlayOpen(): boolean {
  return Boolean(
    document.querySelector('.ant-select-open') ||
      document.querySelector('.ant-picker-dropdown:not(.ant-picker-dropdown-hidden)'),
  );
}

/**
 * میانبرهای صفحه پذیرش: F8 پرینت/ذخیره+پرینت، F1 خدمت، F2 ذخیره+پرینت، Esc آخر.
 * روی window با capture ثبت می‌شود.
 */
export function useReceptionHotkeys({
  disabled,
  onF8Print,
  onAddService,
  onSaveAndPrint,
  onLast,
}: ReceptionHotkeyHandlers) {
  const onF8PrintRef = useRef(onF8Print);
  const onAddServiceRef = useRef(onAddService);
  const onSaveAndPrintRef = useRef(onSaveAndPrint);
  const onLastRef = useRef(onLast);

  useEffect(() => {
    onF8PrintRef.current = onF8Print;
  }, [onF8Print]);

  useEffect(() => {
    onAddServiceRef.current = onAddService;
  }, [onAddService]);

  useEffect(() => {
    onSaveAndPrintRef.current = onSaveAndPrint;
  }, [onSaveAndPrint]);

  useEffect(() => {
    onLastRef.current = onLast;
  }, [onLast]);

  useEffect(() => {
    function handleKeyDown(e: KeyboardEvent) {
      if (disabled) return;

      if (e.key === 'Escape') {
        if (isOverlayOpen()) return;
        e.preventDefault();
        e.stopPropagation();
        onLastRef.current();
        return;
      }

      if (matchFunctionKey(e, 8)) {
        e.preventDefault();
        e.stopPropagation();
        void onF8PrintRef.current();
        return;
      }
      if (matchFunctionKey(e, 1)) {
        e.preventDefault();
        e.stopPropagation();
        onAddServiceRef.current();
        return;
      }
      if (matchFunctionKey(e, 2)) {
        e.preventDefault();
        e.stopPropagation();
        void onSaveAndPrintRef.current();
      }
    }

    window.addEventListener('keydown', handleKeyDown, true);
    return () => window.removeEventListener('keydown', handleKeyDown, true);
  }, [disabled]);
}

/**
 * هندلر میانبر برای attach مستقیم روی ظرف صفحه (fallback).
 * @returns تابع onKeyDown برای Flex/div ریشه
 */
export function createReceptionHotkeyHandler(handlers: ReceptionHotkeyHandlers) {
  return (e: ReactKeyboardEvent<HTMLElement>) => {
    if (handlers.disabled) return;

    if (e.key === 'Escape') {
      if (isOverlayOpen()) return;
      e.preventDefault();
      e.stopPropagation();
      handlers.onLast();
      return;
    }

    if (matchFunctionKey(e.nativeEvent, 8)) {
      e.preventDefault();
      e.stopPropagation();
      void handlers.onF8Print();
      return;
    }
    if (matchFunctionKey(e.nativeEvent, 1)) {
      e.preventDefault();
      e.stopPropagation();
      handlers.onAddService();
      return;
    }
    if (matchFunctionKey(e.nativeEvent, 2)) {
      e.preventDefault();
      e.stopPropagation();
      void handlers.onSaveAndPrint();
    }
  };
}
