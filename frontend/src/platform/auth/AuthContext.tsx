import { createContext, useContext, useEffect, useCallback, type ReactNode } from 'react';
import { useNavigate } from 'react-router-dom';
import { httpClient, setUnauthorizedHandler } from '../api/httpClient';
import { useAuthStore, type AuthUser } from './authStore';

type MeResponse = {
  user: {
    id: number;
    username: string;
    name: string;
    family: string;
    role_id: number;
    role_name: string;
  };
  permissions: string[];
};

type AuthContextValue = {
  user: AuthUser | null;
  permissions: string[];
  isLoading: boolean;
  isAuthenticated: boolean;
  hasPermission: (permission: string) => boolean;
  hasRole: (role: string) => boolean;
  refresh: () => Promise<void>;
  logout: () => Promise<void>;
};

const AuthContext = createContext<AuthContextValue | null>(null);

type AuthProviderProps = {
  children: ReactNode;
};

/** Provider احراز هویت — در اولین بار لود، GET /api/me را فراخوانی می‌کند */
export function AuthProvider({ children }: AuthProviderProps) {
  const navigate = useNavigate();
  const { user, permissions, isLoading, isAuthenticated, setUser, setLoading, clear } =
    useAuthStore();

  const mapUser = (data: MeResponse): AuthUser => ({
    id: data.user.id,
    username: data.user.username,
    name: data.user.name,
    family: data.user.family,
    roleId: data.user.role_id,
    roleName: data.user.role_name,
  });

  const refresh = useCallback(async () => {
    setLoading(true);
    try {
      const { data } = await httpClient.get<MeResponse>('/me');
      setUser(mapUser(data), data.permissions ?? []);
    } catch {
      clear();
    }
  }, [setUser, setLoading, clear]);

  const logout = useCallback(async () => {
    try {
      await httpClient.post('/logout');
    } catch {
      // حتی در صورت خطا، وضعیت محلی پاک می‌شود
    }
    clear();
    navigate('/login', { replace: true });
  }, [clear, navigate]);

  useEffect(() => {
    setUnauthorizedHandler(() => {
      clear();
      if (window.location.pathname !== '/login') {
        navigate('/login', { replace: true });
      }
    });
  }, [clear, navigate]);

  useEffect(() => {
    refresh();
  }, [refresh]);

  const hasPermission = (permission: string) => permissions.includes(permission);
  const hasRole = (role: string) => user?.roleName === role;

  return (
    <AuthContext.Provider
      value={{
        user,
        permissions,
        isLoading,
        isAuthenticated,
        hasPermission,
        hasRole,
        refresh,
        logout,
      }}
    >
      {children}
    </AuthContext.Provider>
  );
}

/** دسترسی به Context احراز هویت */
export function useAuthContext(): AuthContextValue {
  const ctx = useContext(AuthContext);
  if (!ctx) {
    throw new Error('useAuthContext باید داخل AuthProvider استفاده شود');
  }
  return ctx;
}
