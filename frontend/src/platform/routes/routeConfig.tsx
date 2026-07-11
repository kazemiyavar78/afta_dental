import type { ReactNode } from 'react';
import { LoginPage } from '@/modules/auth/pages/LoginPage';
import { UserListPage } from '@/modules/users/pages/UserListPage';
import { UserFormPage } from '@/modules/users/pages/UserFormPage';
import { ListSessions } from '@/modules/users/pages/ListSessions';
import { ReceptionListPage } from '@/modules/reception/pages/ReceptionListPage';
import { ReceptionFormPage } from '@/modules/reception/pages/ReceptionFormPage';
import { OrganizationListPage } from '@/modules/organization/pages/OrganizationListPage';
import { OrganizationFormPage } from '@/modules/organization/pages/OrganizationFormPage';
import { FundListPage } from '@/modules/fund/pages/FundListPage';
import { FundFormPage } from '@/modules/fund/pages/FundFormPage';
import { TariffListPage } from '@/modules/tariff/pages/TariffListPage';
import { TariffFormPage } from '@/modules/tariff/pages/TariffFormPage';
import { SettingsPage } from '@/modules/settings/pages/SettingsPage';
import { ProfilePage } from '@/modules/profile/pages/ProfilePage';
import { LogsPage } from '@/modules/logs/pages/LogsPage';

/** تعریف مسیر اپلیکیشن */
export type AppRoute = {
  path: string;
  element: ReactNode;
  public?: boolean;
  requiredPermission?: string;
};

/** نگاشت مرکزی مسیرها — افزودن ماژول جدید فقط یک آیتم به این آرایه */
export const routeConfig: AppRoute[] = [
  { path: '/login', element: <LoginPage />, public: true },

  { path: '/users', element: <UserListPage />, requiredPermission: 'users.read' },
  { path: '/users/new', element: <UserFormPage />, requiredPermission: 'users.create' },
  { path: '/users/:id/edit', element: <UserFormPage />, requiredPermission: 'users.update' },
  {
    path: '/users/:id/sessions',
    element: <ListSessions />,
    requiredPermission: 'users.listSessions',
  },

  { path: '/reception', element: <ReceptionListPage />, requiredPermission: 'reception.read' },
  { path: '/reception/new', element: <ReceptionFormPage />, requiredPermission: 'reception.create' },

  { path: '/organization', element: <OrganizationListPage />, requiredPermission: 'organization.read' },
  { path: '/organization/new', element: <OrganizationFormPage />, requiredPermission: 'organization.create' },

  { path: '/fund', element: <FundListPage />, requiredPermission: 'fund.read' },
  { path: '/fund/new', element: <FundFormPage />, requiredPermission: 'fund.create' },

  { path: '/tariff', element: <TariffListPage />, requiredPermission: 'tariff.read' },
  { path: '/tariff/new', element: <TariffFormPage />, requiredPermission: 'tariff.create' },

  { path: '/settings', element: <SettingsPage />, requiredPermission: 'security.settings' },

  { path: '/logs', element: <LogsPage />, requiredPermission: 'logs.read' },

  { path: '/profile', element: <ProfilePage /> },
];
