import type { ReactNode } from 'react';
import { useAuth } from './useAuth';

type PermissionGuardProps = {
  permission: string;
  children: ReactNode;
};

/**
 * اگر کاربر مجوز نداشته باشد، فرزندان رندر نمی‌شوند.
 * @example <PermissionGuard permission="users.create"><Button>ایجاد</Button></PermissionGuard>
 */
export function PermissionGuard({ permission, children }: PermissionGuardProps) {
  const { hasPermission, isLoading } = useAuth();

  if (isLoading) return null;
  if (!hasPermission(permission)) return null;

  return <>{children}</>;
}
