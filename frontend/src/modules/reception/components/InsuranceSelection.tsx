import { useEffect, useRef } from 'react';
import { Col, Form, InputNumber, Row, Select } from 'antd';
import { useQuery } from '@tanstack/react-query';
import { fetchOrganizations } from '@/modules/organization/api';
import { useReceptionStore } from '../store/receptionStore';

export type InsuranceRecalcOverrides = {
  insurance_id?: number | null;
  additional_insurance_id?: number | null;
  additional_insurance_percentage?: number | null;
  additional_insurance_coverage?: number | null;
};

type InsuranceSelectionProps = {
  /** پس از تغییر سازمان/درصد/سقف، محاسبه مجدد خدمات با مقادیر تازه */
  onInsuranceChanged: (overrides?: InsuranceRecalcOverrides) => void;
};

/** فوکوس به کنترل داخل ظرف */
function focusField(wrap: HTMLElement | null) {
  wrap?.querySelector<HTMLElement>('input')?.focus();
}

/** انتخاب بیمه پایه و تکمیلی (حداقل یکی الزامی) */
export function InsuranceSelection({ onInsuranceChanged }: InsuranceSelectionProps) {
  const editing = useReceptionStore((s) => s.editing);
  const insuranceId = useReceptionStore((s) => s.insuranceId);
  const additionalInsuranceId = useReceptionStore((s) => s.additionalInsuranceId);
  const percentage = useReceptionStore((s) => s.additionalInsurancePercentage);
  const coverage = useReceptionStore((s) => s.additionalInsuranceCoverage);
  const focusToken = useReceptionStore((s) => s.insuranceFocusToken);
  const setInsuranceId = useReceptionStore((s) => s.setInsuranceId);
  const setAdditionalInsuranceId = useReceptionStore((s) => s.setAdditionalInsuranceId);
  const setPercentage = useReceptionStore((s) => s.setAdditionalInsurancePercentage);
  const setCoverage = useReceptionStore((s) => s.setAdditionalInsuranceCoverage);

  const baseRef = useRef<HTMLDivElement>(null);
  const suppRef = useRef<HTMLDivElement>(null);
  const percentageRef = useRef<HTMLDivElement>(null);
  const coverageRef = useRef<HTMLDivElement>(null);

  const { data: organizationsData } = useQuery({
    queryKey: ['organizations'],
    queryFn: fetchOrganizations,
  });
  const organizations = organizationsData ?? [];

  const baseOptions = organizations
    .filter((o) => !o.is_takmili)
    .map((o) => ({ value: o.id, label: o.name }));
  const suppOptions = organizations
    .filter((o) => o.is_takmili)
    .map((o) => ({ value: o.id, label: o.name }));

  useEffect(() => {
    if (focusToken > 0) {
      window.setTimeout(() => focusField(baseRef.current), 0);
    }
  }, [focusToken]);

  return (
    <Form layout="vertical" size="middle">
      <Row gutter={[12, 0]}>
        <Col xs={24} sm={12} md={6}>
          <div ref={baseRef}>
            <Form.Item label="بیمه پایه">
              <Select
                allowClear
                showSearch
                optionFilterProp="label"
                disabled={!editing}
                value={insuranceId ?? undefined}
                options={baseOptions}
                onChange={(v) => {
                  const next = v ?? null;
                  setInsuranceId(next);
                  onInsuranceChanged({ insurance_id: next });
                  if (next != null) {
                    window.setTimeout(() => focusField(suppRef.current), 0);
                  }
                }}
              />
            </Form.Item>
          </div>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <div ref={suppRef}>
            <Form.Item label="بیمه تکمیلی">
              <Select
                allowClear
                showSearch
                optionFilterProp="label"
                disabled={!editing}
                value={additionalInsuranceId ?? undefined}
                options={suppOptions}
                onChange={(v) => {
                  const next = v ?? null;
                  setAdditionalInsuranceId(next);
                  if (next == null) {
                    setPercentage(null);
                    setCoverage(null);
                    onInsuranceChanged({
                      additional_insurance_id: null,
                      additional_insurance_percentage: null,
                      additional_insurance_coverage: null,
                    });
                    return;
                  }
                  onInsuranceChanged({ additional_insurance_id: next });
                  window.setTimeout(() => focusField(percentageRef.current), 0);
                }}
              />
            </Form.Item>
          </div>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <div ref={percentageRef}>
            <Form.Item label="درصد بیمه تکمیلی">
              <InputNumber
                style={{ width: '100%' }}
                min={0}
                max={100}
                disabled={!editing || additionalInsuranceId == null}
                value={percentage ?? undefined}
                onChange={(v) => {
                  const next = v == null ? null : Number(v);
                  setPercentage(next);
                  onInsuranceChanged({ additional_insurance_percentage: next });
                  if (next === 100) {
                    window.setTimeout(() => focusField(coverageRef.current), 0);
                  }
                }}
              />
            </Form.Item>
          </div>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <div ref={coverageRef}>
            <Form.Item label="سقف تعهد بیمه تکمیلی">
              <InputNumber
                style={{ width: '100%' }}
                min={0}
                disabled={!editing || additionalInsuranceId == null}
                value={coverage ?? undefined}
                onChange={(v) => {
                  const next = v == null ? null : Number(v);
                  setCoverage(next);
                  onInsuranceChanged({ additional_insurance_coverage: next });
                }}
              />
            </Form.Item>
          </div>
        </Col>
      </Row>
    </Form>
  );
}
