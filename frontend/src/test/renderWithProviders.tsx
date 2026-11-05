import { render, type RenderOptions, type RenderResult } from '@testing-library/react';
import { ConfigProvider } from 'antd';
import faIR from 'antd/locale/fa_IR';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import type { ReactElement, ReactNode } from 'react';
import { MemoryRouter } from 'react-router-dom';

export type RenderWithProvidersOptions = Omit<RenderOptions, 'wrapper'> & {
  route?: string;
};

/**
 * کامپوننت را با همان پوستهٔ نمایش App رندر می‌کند.
 * ui کامپوننت تحت تست است. options مسیر اولیهٔ MemoryRouter و بقیهٔ گزینه‌های render را می‌گیرد.
 * خروجی نتیجهٔ render است تا تست به DOM دسترسی داشته باشد.
 * ConfigProvider راست‌به‌چپ و locale فارسی است. App.tsx theme سفارشی نمی‌دهد و اینجا هم همان پیش‌فرض antd است.
 * React Query در تست retry ندارد تا خطای شبکه چند بار تکرار نشود.
 */
export function renderWithProviders(
  ui: ReactElement,
  options?: RenderWithProvidersOptions,
): RenderResult {
  const { route = '/', ...renderOptions } = options ?? {};
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: { retry: false, refetchOnWindowFocus: false },
      mutations: { retry: false },
    },
  });

  function Wrapper({ children }: { children: ReactNode }) {
    return (
      <ConfigProvider direction="rtl" locale={faIR}>
        <QueryClientProvider client={queryClient}>
          <MemoryRouter initialEntries={[route]}>{children}</MemoryRouter>
        </QueryClientProvider>
      </ConfigProvider>
    );
  }

  return render(ui, { wrapper: Wrapper, ...renderOptions });
}
