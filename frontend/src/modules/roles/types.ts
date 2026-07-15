export type Permission = {
  id: number;
  name: string;
  description: string;
};

export type RoleDetail = {
  id: number;
  name: string;
  description: string;
  permissions: Permission[];
  permission_ids: number[];
  integrity_ok: boolean;
};

export type CreateRolePayload = {
  name: string;
  description: string;
  permission_ids: number[];
};

export type UpdateRolePayload = {
  name: string;
  description: string;
  permission_ids: number[];
};
