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

/** نوار ناوبری پذیرش: اولین، قبلی، بعدی، آخرین، جدید */
export function NavigationBar({
  loading,
  onFirst,
  onPrev,
  onNext,
  onLast,
  onNew,
}: NavigationBarProps) {
  return (
    <Space wrap>
      <Button icon={<FastBackwardOutlined />} onClick={onFirst} loading={loading}>
        اولین
      </Button>
      <Button icon={<BackwardOutlined />} onClick={onPrev} loading={loading}>
        قبلی
      </Button>
      <Button icon={<ForwardOutlined />} onClick={onNext} loading={loading}>
        بعدی
      </Button>
      <Button icon={<FastForwardOutlined />} onClick={onLast} loading={loading}>
        آخرین
      </Button>
      <PermissionGuard permission="reception.create">
        <Button type="primary" icon={<PlusOutlined />} onClick={onNew}>
          جدید
        </Button>
      </PermissionGuard>
    </Space>
  );
}
