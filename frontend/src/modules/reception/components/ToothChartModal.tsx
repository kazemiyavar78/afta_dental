import { Modal } from 'antd';
import { ToothSelection } from '@/platform/components/tooths/tooth';
import type { ToothQuadrantSelection } from '@/platform/components/tooths/toothQuadrant';

type ToothChartModalProps = {
  open: boolean;
  selectedNumber: number | null;
  selectedDirection: number | null;
  onClose: () => void;
  onSelect: (selection: ToothQuadrantSelection) => void;
};

/** مودال انتخاب دندان با نمودار چهارربعی */
export function ToothChartModal({
  open,
  selectedNumber,
  selectedDirection,
  onClose,
  onSelect,
}: ToothChartModalProps) {
  return (
    <Modal
      title="انتخاب شماره دندان"
      open={open}
      onCancel={onClose}
      footer={null}
      width={560}
      destroyOnHidden
    >
      <ToothSelection
        selectedNumber={selectedNumber}
        selectedDirection={selectedDirection}
        onSelect={(selection) => {
          onSelect(selection);
          onClose();
        }}
      />
    </Modal>
  );
}
