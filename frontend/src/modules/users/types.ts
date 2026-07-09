export type User = {
  id: number;
  username: string;
  address: string;
  name: string;
  family: string;
  phone_number: string;
  medical_code: string | null;
  role_id: number;
  role_name: string;
  is_active: boolean;
  is_locked: boolean;
  last_login_at: string | null;
};

export type Role = {
  id: number;
  name: string;
};

export type CreateUserPayload = {
  username: string;
  password: string;
  address?: string;
  name: string;
  family: string;
  phone_number?: string;
  medical_code?: string | null;
  role_id: number;
};

export type UpdateUserPayload = {
  address?: string;
  name?: string;
  family?: string;
  phone_number?: string;
  medical_code?: string | null;
  role_id?: number;
  is_active?: boolean;
};

export type UserFormValues = {
  username: string;
  password?: string;
  address: string;
  name: string;
  family: string;
  phone_number: string;
  medical_code: string;
  role_id: number;
  is_active?: boolean;
};
