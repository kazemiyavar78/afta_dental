export type SecurityLog = {
  id: number;
  user_id: number | null;
  ip: string;
  event_type: string;
  description: string;
  created_at: string;
};

export type LogFilters = {
  ip?: string;
  username?: string;
  from_date?: string;
  to_date?: string;
};
