import { Button, Space } from 'antd';
import {
  FastBackwardOutlined,
  BackwardOutlined,
  ForwardOutlined,
  FastForwardOutlined,
  PlusOutlined,
} from '@ant-design/icons';
import { PermissionGuard } from '@/platform/auth/PermissionGuard';

type NavigationBarProps = {
  loading?: boolean;
  onFirst: () => void;
  onPrev: () => void;
  onNext: () => void;
  onLast: () => void;
  onNew: () => void;
};

/** نوار ناوبری پذیرش — فشرده با Space.Compact */
export function NavigationBar({
  loading,
  onFirst,
  onPrev,
  onNext,
  onLast,
  onNew,
}: NavigationBarProps) {
  return (
    <Space.Compact>
      <Button size="small" icon={<FastBackwardOutlined />} onClick={onFirst} loading={loading} />
      <Button size="small" icon={<BackwardOutlined />} onClick={onPrev} loading={loading} />
      <Button size="small" icon={<ForwardOutlined />} onClick={onNext} loading={loading} />
      <Button size="small" icon={<FastForwardOutlined />} onClick={onLast} loading={loading} />
      <PermissionGuard permission="reception.create">
        <Button size="small" type="primary" icon={<PlusOutlined />} onClick={onNew}>
          جدید
        </Button>
      </PermissionGuard>
    </Space.Compact>
  );
}
