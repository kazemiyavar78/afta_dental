import { useCallback, useEffect, useState } from 'react';
import { Card, Col, Form, Input, Row, Space, Typography, message } from 'antd';
import { useLocation } from 'react-router-dom';
import { PageHeader } from '@/platform/components/PageHeader';
import { useAuth } from '@/platform/auth/useAuth';
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

/** صفحه واحد فضای کاری پذیرش بیمار */
export function ReceptionWorkspacePage() {
  const { hasPermission } = useAuth();
  const location = useLocation();
  const [navLoading, setNavLoading] = useState(false);
  const [saving, setSaving] = useState(false);

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

  /** ناوبری بین پذیرش‌ها؛ در صورت خالی بودن لیست، فرم جدید باز می‌شود */
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

  /** محاسبه مجدد همه سطرهای دارای خدمت؛ overrides مقادیر تازه بیمه را قبل از خواندن استور اعمال می‌کند */
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

    if (insuranceId == null && additionalInsuranceId == null) return;
    const lines = state.services.filter((s) => s.service_id > 0);
    if (lines.length === 0) return;
    try {
      const calculated = await calculateReceptionServices({
        insurance_id: insuranceId,
        additional_insurance_id: additionalInsuranceId,
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

  /** ساخت payload ذخیره از وضعیت استور */
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
      },
      insurance_id: s.insuranceId,
      additional_insurance_id: s.additionalInsuranceId,
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

  /** ذخیره پذیرش (ایجاد یا ویرایش) */
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

  /** حذف نرم */
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

  /** بازیابی پذیرش حذف‌شده */
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

  return (
    <div className="reception-workspace">
      <PageHeader
        title="پذیرش بیمار"
        extra={
          <Space>
            <Typography.Text type="secondary">
              وضعیت: {store.status}
              {store.deleted ? ' (حذف‌شده)' : ''}
              {store.receptionId != null ? ` — #${store.receptionId}` : ' — جدید'}
            </Typography.Text>
          </Space>
        }
      />

      <Card size="small" styles={{ body: { padding: '10px 16px' } }} style={{ marginBottom: 12 }}>
        <div
          style={{
            display: 'flex',
            flexWrap: 'wrap',
            gap: 12,
            alignItems: 'center',
            justifyContent: 'space-between',
          }}
        >
          <NavigationBar
            loading={navLoading}
            onFirst={() => handleNav('first')}
            onPrev={() => handleNav('prev')}
            onNext={() => handleNav('next')}
            onLast={() => handleNav('last')}
            onNew={() => store.resetNew()}
          />
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
        </div>
      </Card>

      <Card title="اطلاعات بیمار" size="small" style={{ marginBottom: 12 }}>
        <PatientInfo />
      </Card>

      <Card title="بیمه‌ها" size="small" style={{ marginBottom: 12 }}>
        <InsuranceSelection onInsuranceChanged={(overrides) => void recalculate(overrides)} />
      </Card>

      <Card title="پزشک و پذیرش" size="small" style={{ marginBottom: 12 }}>
        <DoctorSelection />
        <Form layout="vertical" size="middle">
          <Row gutter={[12, 0]}>
            <Col xs={24}>
              <Form.Item label="توضیحات">
                <Input.TextArea
                  rows={1}
                  disabled={!store.editing}
                  value={store.description}
                  onChange={(e) => store.setDescription(e.target.value)}
                />
              </Form.Item>
            </Col>
          </Row>
        </Form>
      </Card>

      <Card size="small" style={{ marginBottom: 12 }}>
        <ServicesTable onRecalculate={() => void recalculate()} />
      </Card>

      <style>{`
        .reception-workspace .ant-card-head {
          min-height: 40px;
          padding: 0 12px;
        }
        .reception-workspace .ant-card-head-title {
          font-size: 14px;
          padding: 8px 0;
        }
        .reception-workspace .ant-form-item {
          margin-bottom: 10px;
        }
        .reception-workspace .ant-form-item-label {
          padding-bottom: 2px;
        }
        .reception-workspace .ant-form-item-label > label {
          font-size: 13px;
        }
      `}</style>
    </div>
  );
}
