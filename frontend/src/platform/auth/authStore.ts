import { create } from 'zustand';

/** اطلاعات کاربر لاگین‌شده */
export type AuthUser = {
  id: number;
  username: string;
  name: string;
  family: string;
  roleId: number;
  roleName: string;
};

type AuthState = {
  user: AuthUser | null;
  permissions: string[];
  isLoading: boolean;
  isAuthenticated: boolean;
  setUser: (user: AuthUser | null, permissions: string[]) => void;
  setLoading: (loading: boolean) => void;
  clear: () => void;
};

/** استور Zustand برای وضعیت احراز هویت */
export const useAuthStore = create<AuthState>((set) => ({
  user: null,
  permissions: [],
  isLoading: true,
  isAuthenticated: false,
  setUser: (user, permissions) =>
    set({
      user,
      permissions,
      isAuthenticated: user !== null,
      isLoading: false,
    }),
  setLoading: (isLoading) => set({ isLoading }),
  clear: () =>
    set({
      user: null,
      permissions: [],
      isAuthenticated: false,
      isLoading: false,
    }),
}));
