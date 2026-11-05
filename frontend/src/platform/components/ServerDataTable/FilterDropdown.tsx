import { useEffect, useState } from 'react';
import { Button, Card, Flex, Input, InputNumber, Select, Typography } from 'antd';
import { JalaliDatePicker } from '@/platform/components/JalaliDatePicker/JalaliDatePicker';
import type { ServerColumn } from './types';

type FilterDropdownProps<T> = {
  column: ServerColumn<T>;
  filters: Record<string, string>;
  visible: boolean;
  onApply: (patch: Record<string, string>) => void;
  onClose: () => void;
};

/**
 * کلیدهای فیلتر یک ستون را برمی‌گرداند. بازه تاریخ و عدد دو کلید دارند.
 * @param column تعریف ستون
 */
export function filterKeysFor<T>(column: ServerColumn<T>): string[] {
  if (column.filter === 'jalaliDateRange' || column.filter === 'numberRange') {
    return [`${column.key}_from`, `${column.key}_to`];
  }
  if (column.filter === 'none') return [];
  return [column.key];
}

/**
 * فیلتر یک ستون را در سرستون نشان می‌دهد و تا زدن «اعمال» به جدول نمی‌فرستد.
 * تاریخ از JalaliDatePicker می‌آید و همان‌جا با تقویم میلادی به YYYY-MM-DD تبدیل شده است.
 */
export function FilterDropdown<T>({ column, filters, visible, onApply, onClose }: FilterDropdownProps<T>) {
  const [text, setText] = useState(filters[column.key] ?? '');
  const [selected, setSelected] = useState<string[]>(splitEnum(filters[column.key]));
  const [from, setFrom] = useState(filters[`${column.key}_from`] ?? '');
  const [to, setTo] = useState(filters[`${column.key}_to`] ?? '');

  const applied = JSON.stringify({
    text: filters[column.key] ?? '',
    from: filters[`${column.key}_from`] ?? '',
    to: filters[`${column.key}_to`] ?? '',
  });

  useEffect(() => {
    if (!visible) return;
    const current = JSON.parse(applied) as { text: string; from: string; to: string };
    setText(current.text);
    setSelected(splitEnum(current.text));
    setFrom(current.from);
    setTo(current.to);
  }, [visible, applied]);

  const apply = () => {
    if (column.filter === 'text') onApply({ [column.key]: text.trim() });
    if (column.filter === 'enum') onApply({ [column.key]: selected.join(',') });
    if (column.filter === 'jalaliDateRange' || column.filter === 'numberRange') {
      onApply({ [`${column.key}_from`]: from.trim(), [`${column.key}_to`]: to.trim() });
    }
    onClose();
  };

  const clear = () => {
    const patch: Record<string, string> = {};
    for (const key of filterKeysFor(column)) patch[key] = '';
    onApply(patch);
    onClose();
  };

  return (
    <Card size="small">
      <Flex vertical gap="small">
        {column.filter === 'text' && (
          <Input
            allowClear
            value={text}
            aria-label={`مقدار فیلتر ${column.title}`}
            onChange={(event) => setText(event.target.value)}
            onPressEnter={apply}
          />
        )}
        {column.filter === 'enum' && (
          <Select
            mode="multiple"
            allowClear
            showSearch
            optionFilterProp="label"
            options={column.options ?? []}
            value={selected}
            aria-label={`فیلتر ${column.title}`}
            popupMatchSelectWidth={false}
            getPopupContainer={(node) => node.parentElement ?? document.body}
            onChange={setSelected}
          />
        )}
        {(column.filter === 'jalaliDateRange' || column.filter === 'numberRange') && (
          <RangeInputs kind={column.filter} from={from} to={to} onFrom={setFrom} onTo={setTo} />
        )}
        <Flex gap="small">
          <Button type="primary" size="small" htmlType="button" onClick={apply}>
            اعمال
          </Button>
          <Button size="small" htmlType="button" onClick={clear}>
            پاک کردن
          </Button>
        </Flex>
      </Flex>
    </Card>
  );
}

type RangeInputsProps = {
  kind: 'jalaliDateRange' | 'numberRange';
  from: string;
  to: string;
  onFrom: (value: string) => void;
  onTo: (value: string) => void;
};

/**
 * دو سر بازه را می‌گیرد. تاریخ شمسی است و مقدار ذخیره‌شده میلادی YYYY-MM-DD است.
 */
function RangeInputs({ kind, from, to, onFrom, onTo }: RangeInputsProps) {
  if (kind === 'jalaliDateRange') {
    return (
      <Flex vertical gap="small">
        <Typography.Text type="secondary">از</Typography.Text>
        <JalaliDatePicker aria-label="از تاریخ" value={from} onChange={onFrom} />
        <Typography.Text type="secondary">تا</Typography.Text>
        <JalaliDatePicker aria-label="تا تاریخ" value={to} onChange={onTo} />
      </Flex>
    );
  }
  return (
    <Flex vertical gap="small">
      <InputNumber aria-label="از عدد" value={from === '' ? null : Number(from)} onChange={(value) => onFrom(numberText(value))} />
      <InputNumber aria-label="تا عدد" value={to === '' ? null : Number(to)} onChange={(value) => onTo(numberText(value))} />
    </Flex>
  );
}

/**
 * مقدار enum جداشده با ویرگول را به فهرست انتخاب برمی‌گرداند.
 * @param value متن فیلتر یا تهی
 */
function splitEnum(value: string | undefined): string[] {
  if (!value?.trim()) return [];
  return value.split(',').map((part) => part.trim()).filter(Boolean);
}

/**
 * عدد InputNumber را به متن فیلتر تبدیل می‌کند. تهی یعنی این سر بازه حذف شود.
 * @param value عدد انتخاب‌شده یا null
 */
function numberText(value: number | string | null): string {
  if (value == null || value === '') return '';
  return String(value);
}
