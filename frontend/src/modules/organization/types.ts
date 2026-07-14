// این ماژول نمونه است؛ باقی فیلدها/صفحات طبق همین الگو در فازهای بعدی اضافه می‌شوند.

export type Organization = {
  id: number;
  name: string;
  is_takmili: boolean;
  is_active: boolean;
};

export type CreateOrganizationPayload = {
  name: string;
  is_takmili: boolean;
  is_active: boolean;
};

export type UpdateOrganizationPayload = {
  name: string;
  is_takmili: boolean;
  is_active: boolean;
};
