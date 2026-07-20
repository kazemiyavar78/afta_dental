import { Modal, Tag } from 'antd';
import type { ColumnsType } from 'antd/es/table';
import { DataTable } from '@/platform/components/DataTable';
import { useApiQuery } from '@/platform/hooks/useApiQuery';
import { fetchPatientServiceHistory } from '../api';
import type { PatientServiceHistoryItem } from '../types';

type Props = {
  open: boolean;
  patientId: number | null;
  onClose: () => void;
};

/** مدال تاریخچه خدمات پرونده از اولین پذیرش تا الان */
export function PatientServicesHistoryModal({ open, patientId, onClose }: Props) {
  const { data = [], isLoading } = useApiQuery({
    queryKey: ['patient-services-history', patientId],
    queryFn: () => fetchPatientServiceHistory(patientId!),
    enabled: open && !!patientId,
  });

  const columns: ColumnsType<PatientServiceHistoryItem> = [
    { title: 'تاریخ پذیرش', dataIndex: 'reception_date', key: 'reception_date', width: 120 },
    {
      title: 'بیمه پایه',
      dataIndex: 'insurance_name',
      key: 'insurance_name',
      width: 140,
      render: (v: string) => v || '—',
    },
    {
      title: 'بیمه تکمیلی',
      dataIndex: 'additional_insurance_name',
      key: 'additional_insurance_name',
      width: 140,
      render: (v: string) => v || '—',
    },
    {
      title: 'صندوق',
      dataIndex: 'cash_amount',
      key: 'cash_amount',
      width: 120,
      render: (v: number) => `${Number(v).toLocaleString('fa-IR')} ریال`,
    },
    {
      title: 'لیست خدمات دریافت‌شده',
      dataIndex: 'service_names',
      key: 'service_names',
      render: (names: string[]) => (
        <div style={{ display: 'flex', flexWrap: 'wrap', gap: 4 }}>
          {(names ?? []).map((n, i) => (
            <Tag key={`${n}-${i}`}>{n}</Tag>
          ))}
        </div>
      ),
    },
  ];

  return (
    <Modal
      title="خدمات دریافت‌شده پرونده"
      open={open}
      onCancel={onClose}
      footer={null}
      width={1000}
      destroyOnHidden
      styles={{ body: { maxHeight: '70vh', overflowY: 'auto' } }}
    >
      <DataTable columns={columns} data={data} loading={isLoading} rowKey="reception_id" pageSize={10} />
    </Modal>
  );
}
