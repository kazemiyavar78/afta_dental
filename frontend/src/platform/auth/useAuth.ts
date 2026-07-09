import { useAuthContext } from './AuthContext';
import { useAuthStore } from './authStore';

/**
 * هوک دسترسی به وضعیت احراز هویت.
 * از Context برای متدها و Zustand برای state استفاده می‌کند.
 */
export function useAuth() {
  const context = useAuthContext();
  const store = useAuthStore();
  return { ...context, ...store };
}
