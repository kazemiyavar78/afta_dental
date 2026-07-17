import type { ReactNode } from 'react';
import { LoginPage } from '@/modules/auth/pages/LoginPage';
import { UserListPage } from '@/modules/users/pages/UserListPage';
import { UserFormPage } from '@/modules/users/pages/UserFormPage';
import { ListSessions } from '@/modules/users/pages/ListSessions';
import { ReceptionWorkspacePage } from '@/modules/reception/pages/ReceptionWorkspacePage';
import { OrganizationListPage } from '@/modules/organization/pages/OrganizationListPage';
import { OrganizationFormPage } from '@/modules/organization/pages/OrganizationFormPage';
import { OrganizationPackagesPage } from '@/modules/organization-packages/pages/OrganizationPackagesPage';
import { ServicesPage } from '@/modules/services/pages/ServicesPage';
import { PatientsPage } from '@/modules/patients/pages/PatientsPage';
import { FundListPage } from '@/modules/fund/pages/FundListPage';
import { FundFormPage } from '@/modules/fund/pages/FundFormPage';
import { TariffListPage } from '@/modules/tariff/pages/TariffListPage';
import { TariffFormPage } from '@/modules/tariff/pages/TariffFormPage';
import { SettingsPage } from '@/modules/settings/pages/SettingsPage';
import { ProfilePage } from '@/modules/profile/pages/ProfilePage';
import { LogsPage } from '@/modules/logs/pages/LogsPage';
import { RolesPage } from '@/modules/roles/pages/RolesPage';

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

  { path: '/roles', element: <RolesPage />, requiredPermission: 'roles.read' },

  { path: '/reception', element: <ReceptionWorkspacePage />, requiredPermission: 'reception.read' },
  { path: '/reception/new', element: <ReceptionWorkspacePage />, requiredPermission: 'reception.create' },

  { path: '/patients', element: <PatientsPage />, requiredPermission: 'patient.read' },

  { path: '/organization', element: <OrganizationListPage />, requiredPermission: 'organization.read' },
  { path: '/organization/new', element: <OrganizationFormPage />, requiredPermission: 'organization.create' },
  { path: '/organization/:id/edit', element: <OrganizationFormPage />, requiredPermission: 'organization.update' },

  {
    path: '/organization-packages',
    element: <OrganizationPackagesPage />,
    requiredPermission: 'organization_packages.read',
  },

  { path: '/services', element: <ServicesPage />, requiredPermission: 'services.read' },

  { path: '/fund', element: <FundListPage />, requiredPermission: 'fund.read' },
  { path: '/fund/new', element: <FundFormPage />, requiredPermission: 'fund.create' },

  { path: '/tariff', element: <TariffListPage />, requiredPermission: 'tariff.read' },
  { path: '/tariff/new', element: <TariffFormPage />, requiredPermission: 'tariff.read' },

  { path: '/settings', element: <SettingsPage />, requiredPermission: 'security.settings' },

  { path: '/logs', element: <LogsPage />, requiredPermission: 'logs.read' },

  { path: '/profile', element: <ProfilePage /> },
];
