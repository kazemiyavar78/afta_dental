import { useEffect, useRef } from 'react';
import { Col, Form, Row, Select, Typography } from 'antd';
import { useQuery } from '@tanstack/react-query';
import { fetchAssistants, fetchDoctors } from '../api';
import { useReceptionStore } from '../store/receptionStore';

function focusField(wrap: HTMLElement | null) {
  wrap?.querySelector<HTMLElement>('input')?.focus();
}

/** انتخاب پزشک/دستیار و تاریخ‌ها — فرم فشرده داخل کارت */
export function DoctorSelection() {
  const editing = useReceptionStore((s) => s.editing);
  const doctorId = useReceptionStore((s) => s.doctorId);
  const doctorMedicalCode = useReceptionStore((s) => s.doctorMedicalCode);
  const assistantId = useReceptionStore((s) => s.assistantId);
  const bookingDateFocusToken = useReceptionStore((s) => s.bookingDateFocusToken);
  const setDoctor = useReceptionStore((s) => s.setDoctor);
  const setAssistant = useReceptionStore((s) => s.setAssistant);
  const requestBookingDateFocus = useReceptionStore((s) => s.requestBookingDateFocus);

  const assistantWrapRef = useRef<HTMLDivElement>(null);
  const bookingWrapRef = useRef<HTMLDivElement>(null);

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
    <Form layout="vertical" size="small">
      <Row gutter={[8, 0]}>
        <Col span={24}>
          <Form.Item
            label="پزشک"
            required
            style={{ marginBottom: 8 }}
            extra={
              doctorMedicalCode ? (
                <Typography.Text type="secondary" style={{ fontSize: 11 }}>
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
        <Col span={24}>
          <div ref={assistantWrapRef}>
            <Form.Item label="دستیار" style={{ marginBottom: 8 }}>
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
      </Row>
    </Form>
  );
}
