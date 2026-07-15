import { BrowserRouter, Navigate, Route, Routes } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { ConfigProvider } from 'antd';
import faIR from 'antd/locale/fa_IR';
import JalaliProvider from 'antd-jalali-v5';
import { AuthProvider } from '@/platform/auth/AuthContext';
import { RequireAuth } from '@/platform/auth/RequireAuth';
import { AppShell } from '@/platform/layout/AppShell';
import { AuthLayout } from '@/platform/layout/AuthLayout';
import { ErrorBoundary } from '@/platform/components/ErrorBoundary';
import { routeConfig } from '@/platform/routes/routeConfig';

const queryClient = new QueryClient({
  defaultOptions: {
    queries: { refetchOnWindowFocus: false, retry: 1 },
  },
});

/** ریشه اپلیکیشن */
export function App() {
  return (
    <ErrorBoundary>
      <ConfigProvider direction="rtl" locale={faIR}>
        <JalaliProvider />
        <QueryClientProvider client={queryClient}>
          <BrowserRouter>
            <AuthProvider>
              <Routes>
                <Route path="/" element={<Navigate to="/users" replace />} />
                {routeConfig.map((route) => (
                  <Route
                    key={route.path}
                    path={route.path}
                    element={
                      route.public ? (
                        <AuthLayout>{route.element}</AuthLayout>
                      ) : (
                        <RequireAuth requiredPermission={route.requiredPermission}>
                          <AppShell>{route.element}</AppShell>
                        </RequireAuth>
                      )
                    }
                  />
                ))}
                <Route path="*" element={<Navigate to="/users" replace />} />
              </Routes>
            </AuthProvider>
          </BrowserRouter>
        </QueryClientProvider>
      </ConfigProvider>
    </ErrorBoundary>
  );
}
