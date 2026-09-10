import type { ReactNode } from 'react';
import { Button, Space, Tooltip } from 'antd';
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

type NavButtonProps = {
  title: string;
  shortcut?: string;
  icon: ReactNode;
  loading?: boolean;
  onClick: () => void;
  /** دکمه آخر — برای نمایش میانبر Esc */
  highlight?: boolean;
};

/** دکمه ناوبری با Tooltip و میانبر */
function NavButton({ title, shortcut, icon, loading, onClick, highlight }: NavButtonProps) {
  const label = shortcut ? `${title} (${shortcut})` : title;
  return (
    <Tooltip title={label}>
      <Button
        size="small"
        type="default"
        icon={icon}
        loading={loading}
        onClick={onClick}
        aria-label={label}
        className={highlight ? 'reception-nav-last' : undefined}
      />
    </Tooltip>
  );
}

/** نوار ناوبری پذیرش — با Tooltip میانبر و گروه‌بندی بصری */
export function NavigationBar({
  loading,
  onFirst,
  onPrev,
  onNext,
  onLast,
  onNew,
}: NavigationBarProps) {
  return (
    <Space size={4} align="center" wrap className="reception-nav-bar">
      <Space.Compact>
        <NavButton
          title="اولین پذیرش"
          icon={<FastBackwardOutlined />}
          loading={loading}
          onClick={onFirst}
        />
        <NavButton
          title="پذیرش قبلی"
          icon={<BackwardOutlined />}
          loading={loading}
          onClick={onPrev}
        />
        <NavButton
          title="پذیرش بعدی"
          icon={<ForwardOutlined />}
          loading={loading}
          onClick={onNext}
        />
        <NavButton
          title="آخرین پذیرش"
          shortcut="Esc"
          icon={<FastForwardOutlined />}
          loading={loading}
          onClick={onLast}
          highlight
        />
      </Space.Compact>

      <PermissionGuard permission="reception.create">
        <Tooltip title="پذیرش جدید">
          <Button size="small" type="primary" icon={<PlusOutlined />} onClick={onNew}>
            جدید
          </Button>
        </Tooltip>
      </PermissionGuard>

      <style>{`
        .reception-nav-bar .reception-nav-last {
          border-color: #91caff;
          background: #e6f4ff;
        }
        .reception-nav-bar .reception-nav-last:hover {
          border-color: #4096ff !important;
          background: #bae0ff !important;
        }
      `}</style>
    </Space>
  );
}
