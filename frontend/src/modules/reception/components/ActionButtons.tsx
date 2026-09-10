import { Button, Space, Popconfirm } from 'antd';
import {
  SaveOutlined,
  EditOutlined,
  DeleteOutlined,
  PrinterOutlined,
  RollbackOutlined,
} from '@ant-design/icons';
import { useQuery } from '@tanstack/react-query';
import { PermissionGuard } from '@/platform/auth/PermissionGuard';
import { useAuth } from '@/platform/auth/useAuth';
import { fetchOrganizations } from '@/modules/organization/api';
import { printReceptionFromStore } from '../receptionPrintActions';

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

/** دکمه‌های عملیاتی پذیرش — ذخیره / ویرایش / حذف / پرینت */
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
  const { hasPermission } = useAuth();
  const { data: organizationsData } = useQuery({
    queryKey: ['organizations'],
    queryFn: fetchOrganizations,
  });
  const organizations = organizationsData ?? [];

  const canSave =
    !deleted &&
    canEdit &&
    ((isNew && hasPermission('reception.create')) ||
      (!isNew && (hasPermission('reception.update') || hasPermission('reception.create'))));

  /** چاپ قبض ۷ سانتی پذیرش با اطلاعات صفحه جاری */
  function handlePrint() {
    printReceptionFromStore(organizations, hasPermission);
  }

  return (
    <Space.Compact>
      {canSave && (
        <Button type="primary" size="small" icon={<SaveOutlined />} loading={saving} onClick={onSave}>
          ذخیره
        </Button>
      )}

      {!isNew && !deleted && !canEdit && (
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
        <Button size="small" icon={<PrinterOutlined />} onClick={handlePrint}>
          پرینت
        </Button>
      </PermissionGuard>
    </Space.Compact>
  );
}
