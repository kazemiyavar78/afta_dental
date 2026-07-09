import type { ReactNode } from 'react';
import { useAuth } from './useAuth';

type RoleGuardProps = {
  role: string;
  children: ReactNode;
};

/**
 * اگر کاربر نقش مورد نظر را نداشته باشد، فرزندان رندر نمی‌شوند.
 * @example <RoleGuard role="Admin"><Button>تنظیمات</Button></RoleGuard>
 */
export function RoleGuard({ role, children }: RoleGuardProps) {
  const { hasRole, isLoading } = useAuth();

  if (isLoading) return null;
  if (!hasRole(role)) return null;

  return <>{children}</>;
}
