import { Button, Select, Space } from 'antd';
import type { SelectProps } from 'antd';

type SelectOption = { value: number; label: string };

type MultiSelectWithActionsProps = Omit<SelectProps, 'mode' | 'options' | 'onChange' | 'value'> & {
  options: SelectOption[];
  value?: number[];
  onChange?: (value: number[]) => void;
};

/**
 * Select چندانتخابی با دکمه‌های انتخاب همه، حذف همه و معکوس.
 */
export function MultiSelectWithActions({
  options,
  value,
  onChange,
  disabled,
  ...rest
}: MultiSelectWithActionsProps) {
  const selected = value ?? [];

  const handleSelectAll = () => {
    onChange?.(options.map((o) => o.value));
  };

  const handleClearAll = () => {
    onChange?.([]);
  };

  const handleInvert = () => {
    const set = new Set(selected);
    onChange?.(options.filter((o) => !set.has(o.value)).map((o) => o.value));
  };

  return (
    <div>
      <Space size={0} wrap style={{ marginBottom: 4 }}>
        <Button
          type="link"
          size="small"
          disabled={disabled || options.length === 0}
          onClick={handleSelectAll}
        >
          انتخاب همه
        </Button>
        <Button
          type="link"
          size="small"
          disabled={disabled || selected.length === 0}
          onClick={handleClearAll}
        >
          حذف همه
        </Button>
        <Button
          type="link"
          size="small"
          disabled={disabled || options.length === 0}
          onClick={handleInvert}
        >
          معکوس
        </Button>
      </Space>
      <Select
        mode="multiple"
        allowClear
        optionFilterProp="label"
        {...rest}
        disabled={disabled}
        options={options}
        value={value}
        onChange={(vals) => onChange?.(vals as number[])}
        style={{ width: '100%', ...(rest.style as object) }}
      />
    </div>
  );
}
