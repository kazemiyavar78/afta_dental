import { useState } from 'react';
import { Button, Modal, Form, Input } from 'antd';
import { EditOutlined } from '@ant-design/icons';
import type { ColumnsType } from 'antd/es/table';
import { PageHeader } from '@/platform/components/PageHeader';
import { DataTable } from '@/platform/components/DataTable';
import { useApiQuery } from '@/platform/hooks/useApiQuery';
import { useApiMutation } from '@/platform/hooks/useApiMutation';
import { fetchSecuritySettings, updateSecuritySetting } from '../api';
import type { SecuritySetting } from '../types';

export function SettingsPage() {
  const [editing, setEditing] = useState<SecuritySetting | null>(null);
  const [form] = Form.useForm();

  const { data = [], isLoading, refetch } = useApiQuery({
    queryKey: ['security-settings'],
    queryFn: fetchSecuritySettings,
  });

  const mutation = useApiMutation({
    mutationFn: updateSecuritySetting,
    successMessage: 'تنظیم با موفقیت به‌روزرسانی شد',
    onSuccess: () => {
      setEditing(null);
      refetch();
    },
  });

  const columns: ColumnsType<SecuritySetting> = [
    { title: 'شناسه', dataIndex: 'id', key: 'id', hidden: true },
    { title: 'کلید تنظیم', dataIndex: 'name', key: 'name' },
    { title: 'مقدار فعلی', dataIndex: 'value', key: 'value' },
    {
      title: 'عملیات',
      key: 'action',
      render: (_, record) => (
        <Button
          icon={<EditOutlined />}
          onClick={() => {
            setEditing(record);
            form.setFieldsValue(record);
          }}
        >
          ویرایش
        </Button>
      ),
    },
  ];

  return (
    <>
      <PageHeader title="تنظیمات امنیتی" />
      <DataTable columns={columns} data={data} loading={isLoading} rowKey="name" />

      <Modal
        title={`ویرایش ${editing?.name ?? ''}`}
        open={!!editing}
        onCancel={() => setEditing(null)}
        onOk={() => form.submit()}
        confirmLoading={mutation.isPending}
      >
        <Form
          form={form}
          layout="vertical"
          onFinish={(values) => {
            if (!editing) return;
            mutation.mutate({ id: editing.id, name: editing.name, value: values.value });
          }}
        >
          <Form.Item name="value" label="مقدار جدید">
            <Input />
          </Form.Item>
        </Form>
      </Modal>
    </>
  );
}