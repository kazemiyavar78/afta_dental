import { useEffect, useMemo, useState } from 'react';
import {
  Button,
  Card,
  Col,
  Form,
  Input,
  Modal,
  Row,
  Select,
  Space,
  Tag,
  message,
} from 'antd';
import {
  DeleteOutlined,
  EditOutlined,
  PlusOutlined,
  PrinterOutlined,
  SearchOutlined,
  ClearOutlined,
} from '@ant-design/icons';
import { useForm, Controller } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import type { ColumnsType } from 'antd/es/table';
import dayjs from 'dayjs';
import { PageHeader } from '@/platform/components/PageHeader';
import { DataTable } from '@/platform/components/DataTable';
import { PermissionGuard } from '@/platform/auth/PermissionGuard';
import { confirmDialog } from '@/platform/components/ConfirmDialog';
import { JalaliDatePicker } from '@/platform/components/JalaliDatePicker/JalaliDatePicker';
import { useApiQuery } from '@/platform/hooks/useApiQuery';
import { useApiMutation } from '@/platform/hooks/useApiMutation';
import {
  createPatient,
  deletePatient,
  fetchPatients,
  updatePatient,
} from '../api';
import {
  patientSchema,
  patientSearchSchema,
  type PatientFormValues,
  type PatientSearchFormValues,
} from '../hooks';
import type { Patient, PatientSearchParams } from '../types';
import { printPatientsA4 } from '../printPatientsA4';

const emptyFormValues: PatientFormValues = {
  first_name: '',
  last_name: '',
  national_code: '',
  birth_date: '',
  address: '',
  home_phone_number: '',
  mobile_phone_number: '',
  file_number: '',
  sex: true,
};

const emptySearchValues: PatientSearchFormValues = {
  first_name: '',
  last_name: '',
  national_code: '',
  birth_date: '',
  address: '',
  home_phone_number: '',
  mobile_phone_number: '',
  file_number: '',
  sex: '',
};

/** مقدار اختیاری رشته را برای ارسال API پاکسازی می‌کند. */
function optionalText(value?: string | null): string | null {
  const trimmed = value?.trim();
  return trimmed ? trimmed : null;
}

/** فرم جستجو را به پارامتر API تبدیل می‌کند. */
function toSearchParams(values: PatientSearchFormValues): PatientSearchParams {
  const params: PatientSearchParams = {};
  if (values.first_name?.trim()) params.first_name = values.first_name.trim();
  if (values.last_name?.trim()) params.last_name = values.last_name.trim();
  if (values.national_code?.trim()) params.national_code = values.national_code.trim();
  if (values.birth_date?.trim()) params.birth_date = values.birth_date.trim();
  if (values.address?.trim()) params.address = values.address.trim();
  if (values.home_phone_number?.trim()) params.home_phone_number = values.home_phone_number.trim();
  if (values.mobile_phone_number?.trim()) params.mobile_phone_number = values.mobile_phone_number.trim();
  if (values.file_number?.trim()) params.file_number = values.file_number.trim();
  if (values.sex === 'true') params.sex = true;
  if (values.sex === 'false') params.sex = false;
  return params;
}

/** صفحه مدیریت بیماران (لیست، جستجو، ایجاد، ویرایش، حذف و چاپ A4 در یک صفحه) */
export function PatientsPage() {
  const [modalOpen, setModalOpen] = useState(false);
  const [editing, setEditing] = useState<Patient | null>(null);
  const [searchParams, setSearchParams] = useState<PatientSearchParams>({});

  const queryKey = useMemo(() => ['patients', searchParams], [searchParams]);

  const { data = [], isLoading, refetch } = useApiQuery({
    queryKey,
    queryFn: () => fetchPatients(searchParams),
  });

  const {
    control,
    handleSubmit,
    reset,
    setError,
    formState: { errors },
  } = useForm<PatientFormValues>({
    resolver: zodResolver(patientSchema),
    defaultValues: emptyFormValues,
  });

  const {
    control: searchControl,
    handleSubmit: handleSearchSubmit,
    reset: resetSearch,
  } = useForm<PatientSearchFormValues>({
    resolver: zodResolver(patientSearchSchema),
    defaultValues: emptySearchValues,
  });

  useEffect(() => {
    if (!modalOpen) return;
    if (editing) {
      reset({
        first_name: editing.first_name,
        last_name: editing.last_name,
        national_code: editing.national_code,
        birth_date: editing.birth_date,
        address: editing.address ?? '',
        home_phone_number: editing.home_phone_number ?? '',
        mobile_phone_number: editing.mobile_phone_number ?? '',
        file_number: editing.file_number,
        sex: editing.sex,
      });
    } else {
      reset(emptyFormValues);
    }
  }, [editing, modalOpen, reset]);

  const createMutation = useApiMutation({
    mutationFn: createPatient,
    successMessage: 'بیمار با موفقیت ایجاد شد',
    setError,
    onSuccess: () => {
      setModalOpen(false);
      refetch();
    },
  });

  const updateMutation = useApiMutation({
    mutationFn: (values: PatientFormValues) => updatePatient(editing!.id, {
      ...values,
      address: optionalText(values.address),
      home_phone_number: optionalText(values.home_phone_number),
      mobile_phone_number: optionalText(values.mobile_phone_number),
    }),
    successMessage: 'بیمار با موفقیت به‌روزرسانی شد',
    setError,
    onSuccess: () => {
      setModalOpen(false);
      setEditing(null);
      refetch();
    },
  });

  const deleteMutation = useApiMutation({
    mutationFn: deletePatient,
    successMessage: 'بیمار با موفقیت حذف شد',
    onSuccess: () => refetch(),
  });

  const openCreate = () => {
    setEditing(null);
    setModalOpen(true);
  };

  const openEdit = (record: Patient) => {
    setEditing(record);
    setModalOpen(true);
  };

  const closeModal = () => {
    setModalOpen(false);
    setEditing(null);
  };

  const onSearch = (values: PatientSearchFormValues) => {
    setSearchParams(toSearchParams(values));
  };

  const onClearSearch = () => {
    resetSearch(emptySearchValues);
    setSearchParams({});
  };

  const onPrint = () => {
    try {
      printPatientsA4(data, 'فهرست بیماران');
    } catch {
      message.error('امکان باز کردن پنجره چاپ وجود ندارد');
    }
  };

  const isPending = createMutation.isPending || updateMutation.isPending;

  const columns: ColumnsType<Patient> = [
    { title: 'شماره پرونده', dataIndex: 'file_number', key: 'file_number', width: 120 },
    { title: 'نام', dataIndex: 'first_name', key: 'first_name', width: 110 },
    { title: 'نام خانوادگی', dataIndex: 'last_name', key: 'last_name', width: 130 },
    { title: 'کد ملی', dataIndex: 'national_code', key: 'national_code', width: 120 },
    {
      title: 'تاریخ تولد',
      dataIndex: 'birth_date',
      key: 'birth_date',
      width: 120,
      render: (value: string) =>
        value ? dayjs(value).calendar('jalali').locale('fa').format('YYYY/MM/DD') : '—',
    },
    {
      title: 'جنسیت',
      dataIndex: 'sex',
      key: 'sex',
      width: 80,
      render: (sex: boolean) => <Tag color={sex ? 'blue' : 'magenta'}>{sex ? 'مرد' : 'زن'}</Tag>,
    },
    {
      title: 'موبایل',
      dataIndex: 'mobile_phone_number',
      key: 'mobile_phone_number',
      width: 120,
      render: (v: string | null) => v || '—',
    },
    {
      title: 'عملیات',
      key: 'actions',
      width: 200,
      fixed: 'left',
      render: (_, record) => (
        <>
          <PermissionGuard permission="patient.update">
            <Button type="link" icon={<EditOutlined />} onClick={() => openEdit(record)}>
              ویرایش
            </Button>
          </PermissionGuard>
          <PermissionGuard permission="patient.delete">
            <Button
              type="link"
              danger
              icon={<DeleteOutlined />}
              loading={deleteMutation.isPending}
              onClick={() =>
                confirmDialog({
                  title: 'حذف بیمار',
                  content: `آیا از حذف بیمار «${record.first_name} ${record.last_name}» مطمئن هستید؟`,
                  okType: 'danger',
                  onConfirm: () => deleteMutation.mutateAsync(record.id),
                })
              }
            >
              حذف
            </Button>
          </PermissionGuard>
        </>
      ),
    },
  ];

  return (
    <>
      <PageHeader
        title="بیماران"
        extra={
          <Space>
            <Button icon={<PrinterOutlined />} onClick={onPrint} disabled={data.length === 0}>
              چاپ A4
            </Button>
            <PermissionGuard permission="patient.create">
              <Button type="primary" icon={<PlusOutlined />} onClick={openCreate}>
                بیمار جدید
              </Button>
            </PermissionGuard>
          </Space>
        }
      />

      <Card style={{ marginBottom: 16 }} title="جستجو">
        <Form layout="vertical" onFinish={handleSearchSubmit(onSearch)}>
          <Row gutter={12}>
            <Col xs={24} sm={12} md={8} lg={6}>
              <Form.Item label="نام">
                <Controller name="first_name" control={searchControl} render={({ field }) => <Input {...field} allowClear />} />
              </Form.Item>
            </Col>
            <Col xs={24} sm={12} md={8} lg={6}>
              <Form.Item label="نام خانوادگی">
                <Controller name="last_name" control={searchControl} render={({ field }) => <Input {...field} allowClear />} />
              </Form.Item>
            </Col>
            <Col xs={24} sm={12} md={8} lg={6}>
              <Form.Item label="کد ملی">
                <Controller name="national_code" control={searchControl} render={({ field }) => <Input {...field} allowClear />} />
              </Form.Item>
            </Col>
            <Col xs={24} sm={12} md={8} lg={6}>
              <Form.Item label="شماره پرونده">
                <Controller name="file_number" control={searchControl} render={({ field }) => <Input {...field} allowClear />} />
              </Form.Item>
            </Col>
            <Col xs={24} sm={12} md={8} lg={6}>
              <Form.Item label="تاریخ تولد">
                <Controller
                  name="birth_date"
                  control={searchControl}
                  render={({ field }) => (
                    <JalaliDatePicker
                      style={{ width: '100%' }}
                      value={field.value}
                      onChange={field.onChange}
                      allowClear
                    />
                  )}
                />
              </Form.Item>
            </Col>
            <Col xs={24} sm={12} md={8} lg={6}>
              <Form.Item label="جنسیت">
                <Controller
                  name="sex"
                  control={searchControl}
                  render={({ field }) => (
                    <Select
                      {...field}
                      allowClear
                      options={[
                        { label: 'همه', value: '' },
                        { label: 'مرد', value: 'true' },
                        { label: 'زن', value: 'false' },
                      ]}
                    />
                  )}
                />
              </Form.Item>
            </Col>
            <Col xs={24} sm={12} md={8} lg={6}>
              <Form.Item label="موبایل">
                <Controller name="mobile_phone_number" control={searchControl} render={({ field }) => <Input {...field} allowClear />} />
              </Form.Item>
            </Col>
            <Col xs={24} sm={12} md={8} lg={6}>
              <Form.Item label="تلفن منزل">
                <Controller name="home_phone_number" control={searchControl} render={({ field }) => <Input {...field} allowClear />} />
              </Form.Item>
            </Col>
            <Col xs={24} sm={24} md={16} lg={12}>
              <Form.Item label="آدرس">
                <Controller name="address" control={searchControl} render={({ field }) => <Input {...field} allowClear />} />
              </Form.Item>
            </Col>
          </Row>
          <Space>
            <Button type="primary" htmlType="submit" icon={<SearchOutlined />}>
              جستجو
            </Button>
            <Button icon={<ClearOutlined />} onClick={onClearSearch}>
              پاک کردن فیلتر
            </Button>
          </Space>
        </Form>
      </Card>

      <DataTable columns={columns} data={data} loading={isLoading} rowKey="id" />

      <Modal
        title={editing ? `ویرایش بیمار: ${editing.first_name} ${editing.last_name}` : 'بیمار جدید'}
        open={modalOpen}
        onCancel={closeModal}
        footer={null}
        destroyOnHidden
        width={780}
      >
        <Form
          layout="vertical"
          onFinish={handleSubmit((values) => {
            const payload = {
              ...values,
              address: optionalText(values.address),
              home_phone_number: optionalText(values.home_phone_number),
              mobile_phone_number: optionalText(values.mobile_phone_number),
            };
            if (editing) {
              updateMutation.mutate(payload);
            } else {
              createMutation.mutate(payload);
            }
          })}
        >
          <Row gutter={12}>
            <Col span={12}>
              <Form.Item label="نام" validateStatus={errors.first_name ? 'error' : ''} help={errors.first_name?.message}>
                <Controller name="first_name" control={control} render={({ field }) => <Input {...field} />} />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item label="نام خانوادگی" validateStatus={errors.last_name ? 'error' : ''} help={errors.last_name?.message}>
                <Controller name="last_name" control={control} render={({ field }) => <Input {...field} />} />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item label="کد ملی" validateStatus={errors.national_code ? 'error' : ''} help={errors.national_code?.message}>
                <Controller name="national_code" control={control} render={({ field }) => <Input {...field} />} />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item label="شماره پرونده" validateStatus={errors.file_number ? 'error' : ''} help={errors.file_number?.message}>
                <Controller name="file_number" control={control} render={({ field }) => <Input {...field} />} />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item label="تاریخ تولد" validateStatus={errors.birth_date ? 'error' : ''} help={errors.birth_date?.message}>
                <Controller
                  name="birth_date"
                  control={control}
                  render={({ field }) => (
                    <JalaliDatePicker
                      style={{ width: '100%' }}
                      value={field.value}
                      onChange={field.onChange}
                    />
                  )}
                />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item label="جنسیت" validateStatus={errors.sex ? 'error' : ''} help={errors.sex?.message}>
                <Controller
                  name="sex"
                  control={control}
                  render={({ field }) => (
                    <Select
                      value={field.value}
                      onChange={field.onChange}
                      options={[
                        { label: 'مرد', value: true },
                        { label: 'زن', value: false },
                      ]}
                    />
                  )}
                />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item label="موبایل" validateStatus={errors.mobile_phone_number ? 'error' : ''} help={errors.mobile_phone_number?.message}>
                <Controller name="mobile_phone_number" control={control} render={({ field }) => <Input {...field} value={field.value ?? ''} />} />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item label="تلفن منزل" validateStatus={errors.home_phone_number ? 'error' : ''} help={errors.home_phone_number?.message}>
                <Controller name="home_phone_number" control={control} render={({ field }) => <Input {...field} value={field.value ?? ''} />} />
              </Form.Item>
            </Col>
            <Col span={24}>
              <Form.Item label="آدرس" validateStatus={errors.address ? 'error' : ''} help={errors.address?.message}>
                <Controller name="address" control={control} render={({ field }) => <Input.TextArea {...field} value={field.value ?? ''} rows={2} />} />
              </Form.Item>
            </Col>
          </Row>
          <Space>
            <Button type="primary" htmlType="submit" loading={isPending}>
              ذخیره
            </Button>
            <Button onClick={closeModal}>انصراف</Button>
          </Space>
        </Form>
      </Modal>
    </>
  );
}
