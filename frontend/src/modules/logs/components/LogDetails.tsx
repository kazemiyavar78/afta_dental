import { List, Table, Typography } from 'antd';
import type { LogDetails as LogDetailsValue } from '../types';

const HIDDEN_KEYS = new Set(['id', 'user_id', 'actor_user_id', 'actor_id', 'target_id']);

const FIELD_LABELS: Record<string, string> = {
  name: 'نام',
  family: 'نام خانوادگی',
  phone: 'تلفن',
  address: 'آدرس',
  medical_code: 'کد نظام پزشکی',
  role: 'نقش',
  user_type: 'نوع کاربر',
  is_active: 'فعال',
  is_locked: 'قفل',
  description: 'توضیحات',
  key: 'کلید',
  session_prefix: 'پیشوند نشست',
  lock_minutes: 'مدت قفل (دقیقه)',
  scope: 'محدوده',
  ip: 'IP',
  file_number: 'شماره پرونده',
};

type DiffRow = {
  field: string;
  before: string;
  after: string;
};

type ChangeValue = {
  before?: unknown;
  after?: unknown;
};

/**
 * جزئیات یک رویداد را بدون JSON خام نشان می‌دهد.
 * اختلاف فیلدها جدول «فیلد | قبل | بعد» است و مجوزها فهرست اضافه‌شده و حذف‌شده.
 * @param details شیء جزئیات یا تهی
 */
export function LogDetails({ details }: { details?: LogDetailsValue }) {
  const object = asObject(details);
  if (!object) return <Typography.Text type="secondary">جزئیاتی ثبت نشده است</Typography.Text>;

  const rows = diffRows(object);
  const added = stringList(object.added_permissions);
  const removed = stringList(object.removed_permissions);
  const permissions = stringList(object.permissions);
  if (rows.length === 0 && added.length === 0 && removed.length === 0 && permissions.length === 0) {
    return <Typography.Text type="secondary">جزئیاتی ثبت نشده است</Typography.Text>;
  }

  return (
    <>
      {rows.length > 0 && (
        <Table<DiffRow>
          size="small"
          pagination={false}
          rowKey="field"
          dataSource={rows}
          columns={[
            { title: 'فیلد', dataIndex: 'field', key: 'field' },
            { title: 'قبل', dataIndex: 'before', key: 'before' },
            { title: 'بعد', dataIndex: 'after', key: 'after' },
          ]}
        />
      )}
      <PermissionList title="اضافه‌شده" items={added} />
      <PermissionList title="حذف‌شده" items={removed} />
      <PermissionList title="مجوزها" items={permissions} />
    </>
  );
}

/**
 * فهرست نام مجوز را فقط وقتی خالی نیست نشان می‌دهد.
 * @param title برچسب فارسی فهرست
 * @param items نام مجوزها
 */
function PermissionList({ title, items }: { title: string; items: string[] }) {
  if (items.length === 0) return null;
  return (
    <List
      size="small"
      header={<Typography.Text strong>{title}</Typography.Text>}
      dataSource={items}
      renderItem={(item) => <List.Item>{item}</List.Item>}
    />
  );
}

/**
 * جزئیات را به شیء تبدیل می‌کند. رشته JSON نامعتبر یا غیرشیء نادیده گرفته می‌شود تا خام نشان داده نشود.
 * @param details مقدار آمده از API
 */
function asObject(details: unknown): LogDetailsValue | null {
  if (!details) return null;
  if (typeof details === 'string') {
    try {
      return asObject(JSON.parse(details) as unknown);
    } catch {
      return null;
    }
  }
  if (typeof details === 'object' && !Array.isArray(details)) return details as LogDetailsValue;
  return null;
}

/**
 * ردیف‌های اختلاف را از شیء جزئیات می‌سازد. کلید شناسه عددی کاربر حذف می‌شود.
 * @param details شیء جزئیات
 */
function diffRows(details: LogDetailsValue): DiffRow[] {
  if (isFlatChange(details)) {
    const label = typeof details.key === 'string' && details.key ? fieldLabel(details.key) : 'مقدار';
    return [{ field: label, before: formatValue(details.before), after: formatValue(details.after) }];
  }
  const rows: DiffRow[] = [];
  for (const [key, value] of Object.entries(details)) {
    if (HIDDEN_KEYS.has(key) || isPermissionKey(key)) continue;
    if (isChange(value)) {
      rows.push({ field: fieldLabel(key), before: formatValue(value.before), after: formatValue(value.after) });
      continue;
    }
    rows.push({ field: fieldLabel(key), before: '—', after: formatValue(value) });
  }
  return rows;
}

/**
 * می‌گوید این شیء همان شکل تنظیم است: کلید اختیاری به‌علاوه before و after ساده.
 * @param details شیء جزئیات
 */
function isFlatChange(details: LogDetailsValue): boolean {
  const keys = Object.keys(details);
  if (!keys.every((key) => key === 'key' || key === 'before' || key === 'after')) return false;
  return 'before' in details || 'after' in details;
}

/**
 * می‌گوید مقدار یک اختلاف فیلد است، نه یک شیء تودرتوی دیگر.
 * @param value مقدار یک کلید جزئیات
 */
function isChange(value: unknown): value is ChangeValue {
  if (!value || typeof value !== 'object' || Array.isArray(value)) return false;
  const keys = Object.keys(value);
  return keys.length > 0 && keys.every((key) => key === 'before' || key === 'after');
}

/**
 * کلیدهای مربوط به مجوز را از جدول اختلاف جدا می‌کند.
 * @param key نام فیلد جزئیات
 */
function isPermissionKey(key: string): boolean {
  return key === 'added_permissions' || key === 'removed_permissions' || key === 'permissions';
}

/**
 * برچسب فارسی فیلد را برمی‌گرداند. کلید ناشناس همان نام ذخیره‌شده می‌ماند.
 * @param key نام فیلد
 */
function fieldLabel(key: string): string {
  return FIELD_LABELS[key] ?? key;
}

/**
 * مقدار را به متن خوانا تبدیل می‌کند و آکولاد JSON تولید نمی‌کند.
 * @param value مقدار قبل یا بعد
 */
function formatValue(value: unknown): string {
  if (value == null || value === '') return '—';
  if (typeof value === 'boolean') return value ? 'بله' : 'خیر';
  if (typeof value === 'number' || typeof value === 'string') return String(value);
  if (Array.isArray(value)) {
    const parts = value.map((item) => formatValue(item)).filter((item) => item !== '—');
    return parts.length ? parts.join('، ') : '—';
  }
  if (typeof value === 'object') {
    const parts = Object.entries(value as Record<string, unknown>)
      .filter(([key]) => !HIDDEN_KEYS.has(key))
      .map(([key, item]) => `${fieldLabel(key)}: ${formatValue(item)}`);
    return parts.length ? parts.join('، ') : '—';
  }
  return '—';
}

/**
 * فهرست رشته‌ای مجوز را از مقدار جزئیات می‌خواند.
 * @param value آرایه نام یا مقدار دیگر
 */
function stringList(value: unknown): string[] {
  if (!Array.isArray(value)) return [];
  return value.filter((item): item is string => typeof item === 'string' && item.trim() !== '');
}
