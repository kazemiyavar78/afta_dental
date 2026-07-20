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
  isAdmin: boolean;
  isLoading: boolean;
  isAuthenticated: boolean;
  setUser: (user: AuthUser | null, permissions: string[], isAdmin?: boolean) => void;
  setLoading: (loading: boolean) => void;
  clear: () => void;
};

/** استور Zustand برای وضعیت احراز هویت */
export const useAuthStore = create<AuthState>((set) => ({
  user: null,
  permissions: [],
  isAdmin: false,
  isLoading: true,
  isAuthenticated: false,
  setUser: (user, permissions, isAdmin = false) =>
    set({
      user,
      permissions,
      isAdmin,
      isAuthenticated: user !== null,
      isLoading: false,
    }),
  setLoading: (isLoading) => set({ isLoading }),
  clear: () =>
    set({
      user: null,
      permissions: [],
      isAdmin: false,
      isAuthenticated: false,
      isLoading: false,
    }),
}));
