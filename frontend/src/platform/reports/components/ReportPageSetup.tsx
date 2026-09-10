import { Form, InputNumber, Modal, Select } from 'antd';
import type { ReportRuntimePageConfig, ReportMarginPreset, ReportOrientation, ReportPageSize } from '../types';
import {
  REPORT_MARGIN_PRESET_LABELS,
  REPORT_MARGIN_PRESETS,
  REPORT_PAGE_SIZE_LABELS,
} from '../constants';

type ReportPageSetupProps = {
  open: boolean;
  config: ReportRuntimePageConfig;
  onClose: () => void;
  onApply: (config: Partial<ReportRuntimePageConfig>) => void;
};

type FormValues = {
  size: ReportPageSize;
  orientation: ReportOrientation;
  marginPreset: ReportMarginPreset;
  margin: number;
  customWidth?: number;
  customHeight?: number;
};

/**
 * Modal تنظیمات صفحه — size, orientation, margin.
 */
export function ReportPageSetup({ open, config, onClose, onApply }: ReportPageSetupProps) {
  const [form] = Form.useForm<FormValues>();

  const handleOpen = () => {
    const preset = Object.entries(REPORT_MARGIN_PRESETS).find(
      ([, v]) => v === config.margin,
    )?.[0] as ReportMarginPreset | undefined;

    form.setFieldsValue({
      size: config.size,
      orientation: config.orientation,
      marginPreset: preset ?? 'custom',
      margin: config.margin,
      customWidth: config.customWidth,
      customHeight: config.customHeight,
    });
  };

  const handleFinish = (values: FormValues) => {
    const margin =
      values.marginPreset === 'custom'
        ? values.margin
        : REPORT_MARGIN_PRESETS[values.marginPreset as Exclude<ReportMarginPreset, 'custom'>];

    onApply({
      size: values.size,
      orientation: values.orientation,
      margin,
      marginPreset: values.marginPreset,
      customWidth: values.size === 'CUSTOM' ? values.customWidth : undefined,
      customHeight: values.size === 'CUSTOM' ? values.customHeight : undefined,
    });
    onClose();
  };

  return (
    <Modal
      title="تنظیمات صفحه"
      open={open}
      onCancel={onClose}
      onOk={() => form.submit()}
      afterOpenChange={(visible) => visible && handleOpen()}
      destroyOnClose
      width={480}
    >
      <Form form={form} layout="vertical" onFinish={handleFinish}>
        <Form.Item name="size" label="اندازه صفحه" rules={[{ required: true }]}>
          <Select
            options={Object.entries(REPORT_PAGE_SIZE_LABELS).map(([value, label]) => ({
              value,
              label,
            }))}
          />
        </Form.Item>

        <Form.Item name="orientation" label="جهت" rules={[{ required: true }]}>
          <Select
            options={[
              { value: 'portrait', label: 'عمودی (Portrait)' },
              { value: 'landscape', label: 'افقی (Landscape)' },
            ]}
          />
        </Form.Item>

        <Form.Item name="marginPreset" label="حاشیه" rules={[{ required: true }]}>
          <Select
            options={Object.entries(REPORT_MARGIN_PRESET_LABELS).map(([value, label]) => ({
              value,
              label,
            }))}
          />
        </Form.Item>

        <Form.Item noStyle shouldUpdate={(prev, cur) => prev.marginPreset !== cur.marginPreset}>
          {({ getFieldValue }) =>
            getFieldValue('marginPreset') === 'custom' ? (
              <Form.Item name="margin" label="حاشیه سفارشی (میلی‌متر)" rules={[{ required: true }]}>
                <InputNumber min={0} max={50} style={{ width: '100%' }} />
              </Form.Item>
            ) : null
          }
        </Form.Item>

        <Form.Item noStyle shouldUpdate={(prev, cur) => prev.size !== cur.size}>
          {({ getFieldValue }) =>
            getFieldValue('size') === 'CUSTOM' ? (
              <>
                <Form.Item name="customWidth" label="عرض (میلی‌متر)" rules={[{ required: true }]}>
                  <InputNumber min={50} max={500} style={{ width: '100%' }} />
                </Form.Item>
                <Form.Item name="customHeight" label="ارتفاع (میلی‌متر)" rules={[{ required: true }]}>
                  <InputNumber min={50} max={500} style={{ width: '100%' }} />
                </Form.Item>
              </>
            ) : null
          }
        </Form.Item>
      </Form>
    </Modal>
  );
}
