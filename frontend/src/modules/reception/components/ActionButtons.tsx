import { Button, Space, Popconfirm, message } from 'antd';
import {
  SaveOutlined,
  EditOutlined,
  DeleteOutlined,
  PrinterOutlined,
  RollbackOutlined,
} from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import { PermissionGuard } from '@/platform/auth/PermissionGuard';
import { useAuth } from '@/platform/auth/useAuth';

type ActionButtonsProps = {
  saving?: boolean;
  canEdit: boolean;
  deleted: boolean;
  isNew: boolean;
  onSave: () => void;
  onEdit: () => void;
  onDelete: () => void;
  onRestore?: () => void;
};

/** دکمه‌های عملیاتی پذیرش — اندازه یکسان و فشرده */
export function ActionButtons({
  saving,
  canEdit,
  deleted,
  isNew,
  onSave,
  onEdit,
  onDelete,
  onRestore,
}: ActionButtonsProps) {
  const navigate = useNavigate();
  const { hasPermission } = useAuth();

  const canSave =
    !deleted &&
    canEdit &&
    ((isNew && hasPermission('reception.create')) ||
      (!isNew && (hasPermission('reception.update') || hasPermission('reception.create'))));

  return (
    <Space.Compact>
      {canSave && (
        <Button type="primary" size="small" icon={<SaveOutlined />} loading={saving} onClick={onSave}>
          ذخیره
        </Button>
      )}

      {!isNew && !deleted && (
        <PermissionGuard permission="reception.update">
          <Button size="small" icon={<EditOutlined />} onClick={onEdit}>
            ویرایش
          </Button>
        </PermissionGuard>
      )}

      {!isNew && !deleted && (
        <PermissionGuard permission="reception.delete">
          <Popconfirm title="حذف نرم این پذیرش؟" onConfirm={onDelete}>
            <Button size="small" danger icon={<DeleteOutlined />}>
              حذف
            </Button>
          </Popconfirm>
        </PermissionGuard>
      )}

      {deleted && (
        <PermissionGuard permission="reception.restore">
          <Button size="small" icon={<RollbackOutlined />} onClick={onRestore}>
            بازیابی
          </Button>
        </PermissionGuard>
      )}

      <PermissionGuard permission="reception.read">
        <Button
          size="small"
          icon={<PrinterOutlined />}
          onClick={() => {
            if (!hasPermission('reception.read')) {
              message.error('شما مجوز این عملیات را ندارید');
              return;
            }
            window.print();
          }}
        >
          پرینت
        </Button>
      </PermissionGuard>

      <Button size="small" onClick={() => navigate('/')}>
        بازگشت
      </Button>
    </Space.Compact>
  );
}
