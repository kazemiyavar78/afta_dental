import { useCallback, useEffect, useState } from 'react';
import {
  Button,
  Card,
  Col,
  Divider,
  Dropdown,
  Flex,
  Form,
  Input,
  Row,
  Space,
  Tag,
  Typography,
  message,
} from 'antd';
import {
  DollarOutlined,
  DownOutlined,
  HistoryOutlined,
  UnorderedListOutlined,
} from '@ant-design/icons';
import { useLocation } from 'react-router-dom';
import { useAuth } from '@/platform/auth/useAuth';
import { PatientWalletLedgerModal } from '@/modules/wallet/components/PatientWalletLedgerModal';
import type { Patient } from '@/modules/patients/types';
import {
  calculateReceptionServices,
  createReception,
  deleteReception,
  navigateReception,
  restoreReception,
  updateReception,
} from '../api';
import { useReceptionStore } from '../store/receptionStore';
import { lineCashAmount, type UpsertReceptionPayload } from '../types';
import { NavigationBar } from '../components/NavigationBar';
import { PatientInfo } from '../components/PatientInfo';
import {
  InsuranceSelection,
  type InsuranceRecalcOverrides,
} from '../components/InsuranceSelection';
import { DoctorSelection } from '../components/DoctorSelection';
import { ServicesTable } from '../components/ServicesTable';
import { ActionButtons } from '../components/ActionButtons';
import { PatientServicesHistoryModal } from '../components/PatientServicesHistoryModal';
import { PatientReceptionsModal } from '../components/PatientReceptionsModal';

/** صفحه واحد فضای کاری پذیرش بیمار — Layout فشرده HIS/ERP */
export function ReceptionWorkspacePage() {
  const { hasPermission } = useAuth();
  const location = useLocation();
  const [navLoading, setNavLoading] = useState(false);
  const [saving, setSaving] = useState(false);
  const [ledgerOpen, setLedgerOpen] = useState(false);
  const [historyOpen, setHistoryOpen] = useState(false);
  const [receptionsOpen, setReceptionsOpen] = useState(false);

  const store = useReceptionStore();

  /** بارگذاری پذیرش از پاسخ API داخل استور */
  const applyDetail = useCallback((detail: Parameters<typeof store.loadFromDetail>[0]) => {
    if (detail.empty || !detail.id) {
      useReceptionStore.getState().resetNew();
      return;
    }
    useReceptionStore.getState().loadFromDetail(detail);
  }, []);

  /** بارگذاری آخرین پذیرش هنگام ورود؛ مسیر /new یا نبود پذیرش → فرم خالی */
  useEffect(() => {
    let cancelled = false;
    (async () => {
      if (location.pathname.endsWith('/new')) {
        useReceptionStore.getState().resetNew();
        return;
      }
      try {
        const last = await navigateReception('last');
        if (cancelled) return;
        if (last.empty) {
          useReceptionStore.getState().resetNew();
          return;
        }
        applyDetail(last);
      } catch {
        if (!cancelled) useReceptionStore.getState().resetNew();
      }
    })();
    return () => {
      cancelled = true;
    };
  }, [location.pathname, applyDetail]);

  /** ناوبری بین پذیرش‌ها */
  async function handleNav(dir: 'first' | 'prev' | 'next' | 'last') {
    setNavLoading(true);
    try {
      const detail = await navigateReception(dir, store.receptionId);
      if (detail.empty) {
        useReceptionStore.getState().resetNew();
        return;
      }
      applyDetail(detail);
    } catch {
      if (dir === 'first' || dir === 'last') {
        useReceptionStore.getState().resetNew();
      } else {
        message.info('پذیرش دیگری در این جهت وجود ندارد');
      }
    } finally {
      setNavLoading(false);
    }
  }

  /** محاسبه مجدد خدمات */
  const recalculate = useCallback(async (overrides?: InsuranceRecalcOverrides) => {
    const state = useReceptionStore.getState();
    const insuranceId =
      overrides && 'insurance_id' in overrides ? overrides.insurance_id ?? null : state.insuranceId;
    const additionalInsuranceId =
      overrides && 'additional_insurance_id' in overrides
        ? overrides.additional_insurance_id ?? null
        : state.additionalInsuranceId;
    const additionalInsuranceCoverage =
      overrides && 'additional_insurance_coverage' in overrides
        ? overrides.additional_insurance_coverage ?? null
        : state.additionalInsuranceCoverage;
    const additionalInsurancePercentage =
      overrides && 'additional_insurance_percentage' in overrides
        ? overrides.additional_insurance_percentage ?? null
        : state.additionalInsurancePercentage;
    const specialCodeId =
      overrides && 'special_code_id' in overrides
        ? overrides.special_code_id ?? null
        : state.specialCodeId;

    if (insuranceId == null && additionalInsuranceId == null) return;
    const lines = state.services.filter((s) => s.service_id > 0);
    if (lines.length === 0) return;
    try {
      const calculated = await calculateReceptionServices({
        insurance_id: insuranceId,
        additional_insurance_id: additionalInsuranceId,
        special_code_id: specialCodeId,
        additional_insurance_coverage: additionalInsuranceCoverage,
        additional_insurance_percentage: additionalInsurancePercentage,
        services: lines.map((s) => ({
          service_id: s.service_id,
          service_code: s.service_code,
          quantity: s.quantity,
          teeth_number: s.teeth_number,
          teeth_direction: s.teeth_direction,
          service_description: s.service_description,
        })),
      });
      const byId = new Map(calculated.map((c) => [c.service_id, c]));
      const latest = useReceptionStore.getState();
      latest.setServices(
        latest.services.map((s) => {
          const c = byId.get(s.service_id);
          if (!c) return s;
          return {
            ...s,
            service_name: c.service_name,
            quantity: c.quantity,
            service_amount: c.service_amount,
            service_tariff: c.service_tariff,
            service_organization_share: c.service_organization_share,
            service_supplementary_insurance_share: c.service_supplementary_insurance_share,
            service_subsidy_share: c.service_subsidy_share,
            has_dental_direction: c.has_dental_direction,
            has_tooth: c.has_tooth,
          };
        }),
      );
    } catch (err: unknown) {
      const msg =
        (err as { response?: { data?: { error?: string } } })?.response?.data?.error ??
        'خطا در محاسبه خدمات';
      message.error(msg);
    }
  }, []);

  /** ساخت payload ذخیره */
  function buildPayload(save: boolean): UpsertReceptionPayload {
    const s = useReceptionStore.getState();
    return {
      patient: {
        id: s.patient.id,
        first_name: s.patient.first_name,
        last_name: s.patient.last_name,
        national_code: s.patient.national_code,
        birth_date: s.patient.birth_date,
        address: s.patient.address,
        home_phone_number: s.patient.home_phone_number,
        mobile_phone_number: s.patient.mobile_phone_number,
        file_number: s.patient.file_number,
        sex: s.patient.sex,
        is_foreign_national: s.patient.is_foreign_national,
      },
      insurance_id: s.insuranceId,
      additional_insurance_id: s.additionalInsuranceId,
      special_code_id: s.specialCodeId,
      doctor_id: s.doctorId,
      assistant_id: s.assistantId,
      booking_date: s.bookingDate,
      reception_date: s.receptionDate,
      description: s.description,
      discount: s.discount,
      referral_code: s.referralCode,
      additional_insurance_coverage: s.additionalInsuranceCoverage,
      additional_insurance_percentage: s.additionalInsurancePercentage,
      services: s.services
        .filter((x) => x.service_id > 0)
        .map((x) => ({
          service_id: x.service_id,
          service_code: x.service_code,
          quantity: x.quantity,
          teeth_number: x.teeth_number,
          teeth_direction: x.teeth_direction,
          service_description: x.service_description,
        })),
      save,
    };
  }

  async function handleSave() {
    if (store.deleted) {
      message.error('پذیرش حذف شده و قابل ویرایش نیست');
      return;
    }
    if (!hasPermission(store.isNew ? 'reception.create' : 'reception.update') && !store.isNew) {
      if (!hasPermission('reception.create') && !hasPermission('reception.update')) {
        message.error('شما مجوز این عملیات را ندارید');
        return;
      }
    }
    const negativeLine = useReceptionStore
      .getState()
      .services.find((s) => s.service_id > 0 && lineCashAmount(s) < 0);
    if (negativeLine) {
      message.error('سهم صندوق یکی از خدمات منفی است؛ امکان ثبت وجود ندارد.');
      return;
    }
    setSaving(true);
    try {
      const payload = buildPayload(true);
      const detail =
        store.isNew || store.receptionId == null
          ? await createReception(payload)
          : await updateReception(store.receptionId, payload);
      applyDetail(detail);
      message.success('پذیرش ذخیره شد');
    } catch (err: unknown) {
      const msg =
        (err as { response?: { data?: { error?: string } } })?.response?.data?.error ??
        'خطا در ذخیره پذیرش';
      message.error(msg);
    } finally {
      setSaving(false);
    }
  }

  async function handleDelete() {
    if (store.receptionId == null) return;
    try {
      await deleteReception(store.receptionId);
      message.success('پذیرش حذف شد');
      try {
        const last = await navigateReception('last');
        applyDetail(last);
      } catch {
        store.resetNew();
      }
    } catch (err: unknown) {
      const msg =
        (err as { response?: { data?: { error?: string } } })?.response?.data?.error ??
        'خطا در حذف پذیرش';
      message.error(msg);
    }
  }

  async function handleRestore() {
    if (store.receptionId == null) return;
    try {
      const detail = await restoreReception(store.receptionId);
      applyDetail(detail);
      message.success('پذیرش بازیابی شد');
    } catch (err: unknown) {
      const msg =
        (err as { response?: { data?: { error?: string } } })?.response?.data?.error ??
        'خطا در بازیابی';
      message.error(msg);
    }
  }

  const walletPatient: Patient | null =
    store.patient.id != null
      ? {
          id: store.patient.id,
          first_name: store.patient.first_name,
          last_name: store.patient.last_name,
          national_code: store.patient.national_code,
          birth_date: store.patient.birth_date,
          address: store.patient.address,
          home_phone_number: store.patient.home_phone_number,
          mobile_phone_number: store.patient.mobile_phone_number,
          file_number: store.patient.file_number,
          sex: store.patient.sex,
          is_foreign_national: store.patient.is_foreign_national,
        }
      : null;

  const moreTools = [
    ...(hasPermission('wallet.read')
      ? [
          {
            key: 'ledger',
            icon: <DollarOutlined />,
            label: 'پرونده مالی بیمار',
            onClick: () => setLedgerOpen(true),
          },
        ]
      : []),
    {
      key: 'history',
      icon: <HistoryOutlined />,
      label: 'خدمات دریافت‌شده',
      onClick: () => setHistoryOpen(true),
    },
    {
      key: 'receptions',
      icon: <UnorderedListOutlined />,
      label: 'لیست پذیرش‌ها / پایان پذیرش',
      onClick: () => setReceptionsOpen(true),
    },
  ];

  return (
    <Flex vertical gap={8} className="reception-workspace">
      {/* Toolbar */}
      <Card size="small" styles={{ body: { padding: '8px 12px' } }}>
        <Flex wrap="wrap" gap={8} align="center" justify="space-between">
          <Space size={4} wrap>
            <Typography.Text strong>پذیرش</Typography.Text>
            <Tag>{store.isNew ? 'جدید' : `#${store.receptionId}`}</Tag>
            <Tag color={store.status === 'saved' ? 'green' : 'default'}>
              {store.status === 'saved' ? 'ذخیره‌شده' : 'پیش‌نویس'}
            </Tag>
            {store.deleted && <Tag color="red">حذف‌شده</Tag>}
            {store.receptionDate && (
              <Typography.Text type="secondary" style={{ fontSize: 12 }}>
                تاریخ: {store.receptionDate}
              </Typography.Text>
            )}
          </Space>

          <Flex wrap="wrap" gap={8} align="center">
            <NavigationBar
              loading={navLoading}
              onFirst={() => handleNav('first')}
              onPrev={() => handleNav('prev')}
              onNext={() => handleNav('next')}
              onLast={() => handleNav('last')}
              onNew={() => store.resetNew()}
            />
            <Divider type="vertical" style={{ height: 24, margin: 0 }} />
            <ActionButtons
              saving={saving}
              canEdit={store.editing}
              deleted={store.deleted}
              isNew={store.isNew}
              onSave={() => void handleSave()}
              onEdit={() => {
                if (store.deleted) {
                  message.error('پذیرش حذف شده و قابل ویرایش نیست');
                  return;
                }
                if (!hasPermission('reception.update')) {
                  message.error('شما مجوز این عملیات را ندارید');
                  return;
                }
                store.setEditing(true);
              }}
              onDelete={() => void handleDelete()}
              onRestore={() => void handleRestore()}
            />
            {!store.isNew && store.patient.id != null && (
              <Dropdown menu={{ items: moreTools }} trigger={['click']}>
                <Button size="small">
                  ابزارها <DownOutlined />
                </Button>
              </Dropdown>
            )}
          </Flex>
        </Flex>
      </Card>

      {/* بیمار | بیمه | پزشک */}
      <Row gutter={[8, 8]}>
        <Col xs={24} lg={8}>
          <Card title="اطلاعات بیمار" size="small" styles={{ body: { padding: 8 } }}>
            <PatientInfo />
          </Card>
        </Col>
        <Col xs={24} md={12} lg={8}>
          <Card title="بیمه و کد خاص" size="small" styles={{ body: { padding: 8 } }}>
            <InsuranceSelection onInsuranceChanged={(overrides) => void recalculate(overrides)} />
          </Card>
        </Col>
        <Col xs={24} md={12} lg={8}>
          <Card title="پزشک و پذیرش" size="small" styles={{ body: { padding: 8 } }}>
            <DoctorSelection />
            <Form layout="vertical" size="small" style={{ marginTop: 4 }}>
              <Form.Item label="توضیحات" style={{ marginBottom: 0 }}>
                <Input.TextArea
                  rows={1}
                  disabled={!store.editing}
                  value={store.description}
                  onChange={(e) => store.setDescription(e.target.value)}
                />
              </Form.Item>
            </Form>
          </Card>
        </Col>
      </Row>

      {/* جدول خدمات + جمع */}
      <Card size="small" styles={{ body: { padding: 8 } }} style={{ flex: 1, minHeight: 0 }}>
        <ServicesTable onRecalculate={() => void recalculate()} />
      </Card>

      <PatientWalletLedgerModal
        open={ledgerOpen}
        patient={walletPatient}
        onClose={() => setLedgerOpen(false)}
        showPaymentActions
      />
      <PatientServicesHistoryModal
        open={historyOpen}
        patientId={store.patient.id ?? null}
        onClose={() => setHistoryOpen(false)}
      />
      <PatientReceptionsModal
        open={receptionsOpen}
        patientId={store.patient.id ?? null}
        onClose={() => setReceptionsOpen(false)}
        onEnded={() => {
          if (store.receptionId != null) {
            void navigateReception('last').then(applyDetail).catch(() => undefined);
          }
        }}
      />
    </Flex>
  );
}
