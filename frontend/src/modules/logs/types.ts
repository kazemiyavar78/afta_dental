/** یک گزینه فیلتر که value کد سرور و label برچسب فارسی است. */
export type LogOption = {
  value: string;
  label: string;
};

/** فراداده کشوی فیلتر لاگ. */
export type LogMeta = {
  event_types: LogOption[];
  reasons: LogOption[];
  target_types: LogOption[];
};

/** وضعیت صحت زنجیره. last_result برابر ok یا broken است. */
export type LogIntegrity = {
  last_check_at: string | null;
  last_result: string;
  broken_at_id: number;
};

/** جزئیات ماسک‌شده یک رویداد. شیء است، نه متن JSON. */
export type LogDetails = Record<string, unknown>;

/** یک ردیف لاگ امنیتی. شناسه عددی کاربر در این پاسخ برای نمایش نیست. */
export type SecurityLog = {
  id: number;
  created_at: string;
  actor_username: string;
  target_name: string;
  target_type: string;
  event_type: string;
  event_type_label: string;
  result: string;
  reason: string;
  reason_label: string;
  ip: string;
  user_agent: string;
  description: string;
  details?: LogDetails;
};
