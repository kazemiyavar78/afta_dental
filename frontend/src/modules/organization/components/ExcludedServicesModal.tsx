import { useEffect, useMemo, useState } from 'react';
import { Button, Checkbox, Input, Modal, Space, Typography, message } from 'antd';
import { useApiQuery } from '@/platform/hooks/useApiQuery';
import { useAuth } from '@/platform/auth/useAuth';
import { fetchServices } from '@/modules/services/api';
import {
  addExcludedServices,
  fetchExcludedServices,
  removeExcludedServices,
} from '../api';
import type { Organization } from '../types';

type ExcludedServicesModalProps = {
  organization: Organization | null;
  open: boolean;
  onClose: () => void;
};

/**
 * پاپ‌آپ انتخاب چندتایی خدماتی که سازمان پوشش تعرفه‌ای نمی‌دهد.
 */
export function ExcludedServicesModal({ organization, open, onClose }: ExcludedServicesModalProps) {
  const { hasPermission } = useAuth();
  const canAdd = hasPermission('excluded_services.add');
  const canRemove = hasPermission('excluded_services.remove');
  const canEdit = canAdd || canRemove;

  const [selectedIds, setSelectedIds] = useState<number[]>([]);
  const [initialIds, setInitialIds] = useState<number[]>([]);
  const [search, setSearch] = useState('');
  const [saving, setSaving] = useState(false);

  const { data: services = [], isLoading: loadingServices } = useApiQuery({
    queryKey: ['services'],
    queryFn: fetchServices,
    enabled: open,
  });

  const {
    data: excluded = [],
    isLoading: loadingExcluded,
    refetch: refetchExcluded,
  } = useApiQuery({
    queryKey: ['excluded-services', organization?.id],
    queryFn: () => fetchExcludedServices(organization!.id),
    enabled: open && organization != null,
  });

  useEffect(() => {
    if (!open) return;
    const ids = excluded.map((item) => item.service_id);
    setInitialIds(ids);
    setSelectedIds(ids);
    setSearch('');
  }, [excluded, open]);

  const filteredServices = useMemo(() => {
    const q = search.trim().toLowerCase();
    const list = [...services].sort((a, b) => a.service_code.localeCompare(b.service_code));
    if (!q) return list;
    return list.filter(
      (s) =>
        s.name.toLowerCase().includes(q) ||
        s.service_code.toLowerCase().includes(q),
    );
  }, [search, services]);

  /**
   * تغییر انتخاب یک خدمت با رعایت مجوز add/remove.
   * @param serviceId شناسه خدمت
   * @param checked وضعیت جدید
   */
  const toggleService = (serviceId: number, checked: boolean) => {
    const wasInitiallySelected = initialIds.includes(serviceId);
    if (checked && !wasInitiallySelected && !canAdd) return;
    if (!checked && wasInitiallySelected && !canRemove) return;

    setSelectedIds((prev) => {
      if (checked) return [...new Set([...prev, serviceId])];
      return prev.filter((id) => id !== serviceId);
    });
  };

  /**
   * ذخیره اختلاف انتخاب‌ها نسبت به وضعیت اولیه.
   */
  const handleSave = async () => {
    if (!organization || !canEdit) return;

    const initialSet = new Set(initialIds);
    const selectedSet = new Set(selectedIds);
    const toAdd = selectedIds.filter((id) => !initialSet.has(id));
    const toRemove = initialIds.filter((id) => !selectedSet.has(id));

    if (toAdd.length === 0 && toRemove.length === 0) {
      onClose();
      return;
    }

    setSaving(true);
    try {
      if (toAdd.length > 0) {
        await addExcludedServices({
          organization_id: organization.id,
          service_ids: toAdd,
        });
      }
      if (toRemove.length > 0) {
        await removeExcludedServices({
          organization_id: organization.id,
          service_ids: toRemove,
        });
      }
      message.success('خدمات خارج‌شده با موفقیت ذخیره شد');
      await refetchExcluded();
      onClose();
    } catch (err) {
      const appError = err as { message?: string; code?: string };
      if (appError?.code !== 'FORBIDDEN') {
        message.error(appError?.message ?? 'خطا در ذخیره خدمات خارج‌شده');
      }
    } finally {
      setSaving(false);
    }
  };

  const loading = loadingServices || loadingExcluded;

  return (
    <Modal
      title={
        organization
          ? `خدماتی که «${organization.name}» پوشش نمی‌دهد`
          : 'خدمات خارج‌شده'
      }
      open={open}
      onCancel={onClose}
      footer={null}
      destroyOnHidden
      width={640}
    >
      <Space direction="vertical" size={12} style={{ width: '100%' }}>
        <Input.Search
          allowClear
          placeholder="جستجو بر اساس کد یا نام خدمت"
          value={search}
          onChange={(e) => setSearch(e.target.value)}
        />
        {loading ? (
          <Typography.Text type="secondary">در حال بارگذاری خدمات...</Typography.Text>
        ) : filteredServices.length === 0 ? (
          <Typography.Text type="secondary">خدمتی یافت نشد.</Typography.Text>
        ) : (
          <div style={{ maxHeight: 360, overflow: 'auto', paddingInlineEnd: 4 }}>
            <Space direction="vertical" size={4} style={{ width: '100%' }}>
              {filteredServices.map((service) => {
                const checked = selectedIds.includes(service.id);
                const wasInitiallySelected = initialIds.includes(service.id);
                const disableCheck = !checked && !wasInitiallySelected && !canAdd;
                const disableUncheck = checked && wasInitiallySelected && !canRemove;
                return (
                  <Checkbox
                    key={service.id}
                    checked={checked}
                    disabled={disableCheck || disableUncheck || !canEdit}
                    onChange={(e) => toggleService(service.id, e.target.checked)}
                  >
                    <Typography.Text>
                      {service.service_code} — {service.name}
                    </Typography.Text>
                  </Checkbox>
                );
              })}
            </Space>
          </div>
        )}
        <Space>
          {canEdit && (
            <Button type="primary" loading={saving} onClick={handleSave}>
              ذخیره
            </Button>
          )}
          <Button onClick={onClose}>انصراف</Button>
        </Space>
      </Space>
    </Modal>
  );
}
