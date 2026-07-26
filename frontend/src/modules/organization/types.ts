// این ماژول نمونه است؛ باقی فیلدها/صفحات طبق همین الگو در فازهای بعدی اضافه می‌شوند.

export type Organization = {
  id: number;
  name: string;
  is_takmili: boolean;
  is_active: boolean;
  package_id: number;
  package_name?: string;
  center_package_id: number;
  center_package_name?: string;
};

export type CreateOrganizationPayload = {
  name: string;
  is_takmili: boolean;
  is_active: boolean;
  package_id: number;
  center_package_id: number;
};

export type UpdateOrganizationPayload = {
  name: string;
  is_takmili: boolean;
  is_active: boolean;
  package_id: number;
  center_package_id: number;
};

export type ExcludedService = {
  id: number;
  organization_id: number;
  service_id: number;
};

export type ExcludedServicesPayload = {
  organization_id: number;
  service_ids: number[];
};
