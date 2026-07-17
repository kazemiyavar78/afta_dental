import { useEffect, useRef } from 'react';
import { Col, Form, InputNumber, Row, Select, Typography } from 'antd';
import { useQuery } from '@tanstack/react-query';
import { JalaliDatePicker } from '@/platform/components/JalaliDatePicker/JalaliDatePicker';
import { fetchAssistants, fetchDoctors } from '../api';
import { useReceptionStore } from '../store/receptionStore';

/** فوکوس به کنترل داخل ظرف */
function focusField(wrap: HTMLElement | null) {
  wrap?.querySelector<HTMLElement>('input')?.focus();
}

/** انتخاب فشرده پزشک/دستیار به‌همراه تاریخ اعتبار و کد معرفی */
export function DoctorSelection() {
  const editing = useReceptionStore((s) => s.editing);
  const doctorId = useReceptionStore((s) => s.doctorId);
  const doctorMedicalCode = useReceptionStore((s) => s.doctorMedicalCode);
  const assistantId = useReceptionStore((s) => s.assistantId);
  const bookingDate = useReceptionStore((s) => s.bookingDate);
  const discount = useReceptionStore((s) => s.discount);
  const referralCode = useReceptionStore((s) => s.referralCode);
  const bookingDateFocusToken = useReceptionStore((s) => s.bookingDateFocusToken);
  const setDoctor = useReceptionStore((s) => s.setDoctor);
  const setAssistant = useReceptionStore((s) => s.setAssistant);
  const setBookingDate = useReceptionStore((s) => s.setBookingDate);
  const setDiscount = useReceptionStore((s) => s.setDiscount);
  const setReferralCode = useReceptionStore((s) => s.setReferralCode);
  const requestBookingDateFocus = useReceptionStore((s) => s.requestBookingDateFocus);

  const assistantWrapRef = useRef<HTMLDivElement>(null);
  const bookingWrapRef = useRef<HTMLDivElement>(null);
  const discountWrapRef = useRef<HTMLDivElement>(null);
  const referralWrapRef = useRef<HTMLDivElement>(null);

  const { data: doctorsData } = useQuery({
    queryKey: ['doctors'],
    queryFn: fetchDoctors,
  });
  const { data: assistantsData } = useQuery({
    queryKey: ['assistants'],
    queryFn: fetchAssistants,
  });

  const doctors = doctorsData ?? [];
  const assistants = assistantsData ?? [];
  const activeDoctors = doctors.filter((d) => d.is_active);

  useEffect(() => {
    if (bookingDateFocusToken > 0) {
      window.setTimeout(() => focusField(bookingWrapRef.current), 0);
    }
  }, [bookingDateFocusToken]);

  return (
    <Form layout="vertical" size="middle">
      <Row gutter={[12, 0]}>
        <Col xs={24} sm={12} md={6} lg={5}>
          <Form.Item
            label="پزشک"
            required
            extra={
              doctorMedicalCode ? (
                <Typography.Text type="secondary" style={{ fontSize: 12 }}>
                  کد نظام: {doctorMedicalCode}
                </Typography.Text>
              ) : null
            }
          >
            <Select
              showSearch
              optionFilterProp="label"
              disabled={!editing}
              value={doctorId ?? undefined}
              placeholder="انتخاب پزشک"
              options={activeDoctors.map((d) => ({
                value: d.id,
                label: `${d.name} ${d.family}${d.medical_code ? ` (${d.medical_code})` : ''}`,
              }))}
              onChange={(id) => {
                const d = activeDoctors.find((x) => x.id === id);
                if (d) {
                  setDoctor(d.id, `${d.name} ${d.family}`, d.medical_code);
                  window.setTimeout(() => focusField(assistantWrapRef.current), 0);
                }
              }}
              allowClear
              onClear={() => setDoctor(null, '', null)}
            />
          </Form.Item>
        </Col>
        <Col xs={24} sm={12} md={6} lg={5}>
          <div ref={assistantWrapRef}>
            <Form.Item label="دستیار">
              <Select
                showSearch
                allowClear
                optionFilterProp="label"
                disabled={!editing}
                value={assistantId ?? undefined}
                placeholder="اختیاری"
                options={assistants.map((a) => ({
                  value: a.id,
                  label: `${a.name} ${a.family}`,
                }))}
                onChange={(id) => {
                  if (id == null) {
                    setAssistant(null, '');
                    return;
                  }
                  const a = assistants.find((x) => x.id === id);
                  if (a) {
                    setAssistant(a.id, `${a.name} ${a.family}`);
                    requestBookingDateFocus();
                  }
                }}
              />
            </Form.Item>
          </div>
        </Col>
        <Col xs={24} sm={12} md={6} lg={5}>
          <div ref={bookingWrapRef}>
            <Form.Item label="تاریخ اعتبار دفترچه">
              <JalaliDatePicker
                value={bookingDate}
                disabled={!editing}
                style={{ width: '100%' }}
                onChange={(v) => {
                  setBookingDate(v || null);
                  window.setTimeout(() => focusField(discountWrapRef.current), 0);
                }}
              />
            </Form.Item>
          </div>
        </Col>
        <Col xs={24} sm={12} md={6} lg={4}>
          <div ref={discountWrapRef}>
            <Form.Item label="تخفیف">
              <InputNumber
                style={{ width: '100%' }}
                min={0}
                disabled={!editing}
                value={discount}
                onChange={(v) => setDiscount(Number(v) || 0)}
                onPressEnter={() => focusField(referralWrapRef.current)}
              />
            </Form.Item>
          </div>
        </Col>
        <Col xs={24} sm={12} md={6} lg={5}>
          <div ref={referralWrapRef}>
            <Form.Item label="کد معرفی‌نامه تکمیلی">
              <InputNumber
                style={{ width: '100%' }}
                disabled={!editing}
                value={referralCode ?? undefined}
                onChange={(v) => setReferralCode(v == null ? null : Number(v))}
              />
            </Form.Item>
          </div>
        </Col>
      </Row>
    </Form>
  );
}
