import { useRef, useState } from 'react';
import { Button, Modal, Space, Tag, Upload, message } from 'antd';
import { CheckOutlined, UploadOutlined } from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';
import { DataTable } from '@/platform/components/DataTable';
import { useApiQuery } from '@/platform/hooks/useApiQuery';
import { endReception, fetchPatientReceptions, uploadReceptionPhoto } from '../api';
import type { ReceptionDetail } from '../types';

type Props = {
  open: boolean;
  patientId: number | null;
  onClose: () => void;
  onEnded?: () => void;
};

/** جدول پذیرش‌های پرونده با امکان پایان پذیرش و آپلود عکس ضوابط */
export function PatientReceptionsModal({ open, patientId, onClose, onEnded }: Props) {
  const [endingId, setEndingId] = useState<number | null>(null);
  const [photoTarget, setPhotoTarget] = useState<{
    id: number;
    required: number;
    uploaded: number;
    descriptions: string[];
  } | null>(null);
  const fileInputRef = useRef<HTMLInputElement>(null);

  const { data = [], isLoading, refetch } = useApiQuery({
    queryKey: ['patient-receptions', patientId],
    queryFn: () => fetchPatientReceptions(patientId!),
    enabled: open && !!patientId,
  });

  /** تلاش برای پایان پذیرش و نمایش نیاز به عکس در صورت ضوابط */
  async function handleEnd(id: number) {
    setEndingId(id);
    try {
      const result = await endReception(id);
      if (result.success) {
        message.success(result.message);
        refetch();
        onEnded?.();
        return;
      }
      if (result.previous_reception_id) {
        message.warning(result.message);
        return;
      }
      if (result.required_photo_count > result.uploaded_photo_count) {
        setPhotoTarget({
          id,
          required: result.required_photo_count,
          uploaded: result.uploaded_photo_count,
          descriptions: result.regulation_descriptions ?? [],
        });
        message.warning(result.message);
        return;
      }
      message.warning(result.message);
    } catch (err: unknown) {
      const msg =
        (err as { response?: { data?: { error?: string } } })?.response?.data?.error ??
        'خطا در پایان پذیرش';
      message.error(msg);
    } finally {
      setEndingId(null);
    }
  }

  /** آپلود یک فایل عکس و تلاش مجدد پایان پذیرش در صورت تکمیل تعداد */
  async function handleUpload(file: File) {
    if (!photoTarget) return;
    try {
      await uploadReceptionPhoto(photoTarget.id, file);
      const nextUploaded = photoTarget.uploaded + 1;
      message.success('عکس آپلود شد');
      if (nextUploaded >= photoTarget.required) {
        setPhotoTarget(null);
        await handleEnd(photoTarget.id);
      } else {
        setPhotoTarget({ ...photoTarget, uploaded: nextUploaded });
      }
      refetch();
    } catch (err: unknown) {
      const msg =
        (err as { response?: { data?: { error?: string } } })?.response?.data?.error ??
        'خطا در آپلود عکس';
      message.error(msg);
    }
  }

  const columns: ColumnsType<ReceptionDetail> = [
    { title: 'شماره', dataIndex: 'id', key: 'id', width: 80 },
    { title: 'تاریخ پذیرش', dataIndex: 'reception_date', key: 'reception_date', width: 120 },
    {
      title: 'وضعیت',
      dataIndex: 'status',
      key: 'status',
      width: 100,
      render: (v: string) => (v === 'saved' ? 'ذخیره‌شده' : 'پیش‌نویس'),
    },
    {
      title: 'پایان پذیرش',
      dataIndex: 'reception_ended',
      key: 'reception_ended',
      width: 110,
      render: (v: boolean) =>
        v ? <Tag color="green">پایان‌یافته</Tag> : <Tag>باز</Tag>,
    },
    {
      title: 'عکس‌ها',
      dataIndex: 'photo_count',
      key: 'photo_count',
      width: 80,
    },
    {
      title: 'عملیات',
      key: 'actions',
      width: 160,
      render: (_, record) =>
        !record.reception_ended && record.status === 'saved' ? (
          <Button
            type="link"
            icon={<CheckOutlined />}
            loading={endingId === record.id}
            onClick={() => void handleEnd(record.id)}
          >
            پایان پذیرش
          </Button>
        ) : null,
    },
  ];

  return (
    <>
      <Modal
        title="پذیرش‌های پرونده"
        open={open}
        onCancel={onClose}
        footer={null}
        width={900}
        destroyOnHidden
        styles={{ body: { maxHeight: '70vh', overflowY: 'auto' } }}
      >
        <DataTable columns={columns} data={data} loading={isLoading} rowKey="id" pageSize={10} />
      </Modal>

      <Modal
        title="آپلود عکس دندان (ضوابط)"
        open={!!photoTarget}
        onCancel={() => setPhotoTarget(null)}
        footer={null}
        destroyOnHidden
      >
        {photoTarget && (
          <Space direction="vertical" style={{ width: '100%' }}>
            {(photoTarget.descriptions ?? []).map((d, i) => (
              <Tag key={i} color="orange">
                {d}
              </Tag>
            ))}
            <div>
              نیاز به {photoTarget.required} عکس — آپلود شده: {photoTarget.uploaded}
            </div>
            <input
              ref={fileInputRef}
              type="file"
              accept="image/*"
              style={{ display: 'none' }}
              onChange={(e) => {
                const f = e.target.files?.[0];
                if (f) void handleUpload(f);
                e.target.value = '';
              }}
            />
            <Upload
              showUploadList={false}
              beforeUpload={(file) => {
                void handleUpload(file);
                return false;
              }}
            >
              <Button icon={<UploadOutlined />}>انتخاب عکس</Button>
            </Upload>
          </Space>
        )}
      </Modal>
    </>
  );
}
