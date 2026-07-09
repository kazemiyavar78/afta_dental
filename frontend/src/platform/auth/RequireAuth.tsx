import type { ReactNode } from 'react';
import { Navigate, useLocation } from 'react-router-dom';
import { Spin, Result } from 'antd';
import { useAuth } from './useAuth';

type RequireAuthProps = {
  children: ReactNode;
  requiredPermission?: string;
};

/** Wrapper مسیرهایی که نیاز به لاگین دارند */
export function RequireAuth({ children, requiredPermission }: RequireAuthProps) {
  const { isAuthenticated, isLoading, hasPermission } = useAuth();
  const location = useLocation();

  if (isLoading) {
    return (
      <div style={{ display: 'flex', justifyContent: 'center', padding: 48 }}>
        <Spin size="large" tip="در حال بارگذاری..." />
      </div>
    );
  }

  if (!isAuthenticated) {
    return <Navigate to="/login" state={{ from: location }} replace />;
  }

  if (requiredPermission && !hasPermission(requiredPermission)) {
    return (
      <Result
        status="403"
        title="دسترسی غیرمجاز"
        subTitle="شما مجوز دسترسی به این صفحه را ندارید."
      />
    );
  }

  return <>{children}</>;
}
