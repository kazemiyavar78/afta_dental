import { useEffect, useRef, useState } from 'react';
import { Col, Form, Input, InputNumber, Row, Select, Typography, message } from 'antd';
import { useQuery } from '@tanstack/react-query';
import { fetchOrganizations } from '@/modules/organization/api';
import { fetchSpecialCodeByCode } from '@/modules/special-codes/api';
import { useReceptionStore } from '../store/receptionStore';

export type InsuranceRecalcOverrides = {
  insurance_id?: number | null;
  additional_insurance_id?: number | null;
  additional_insurance_percentage?: number | null;
  additional_insurance_coverage?: number | null;
  special_code_id?: number | null;
};

type InsuranceSelectionProps = {
  onInsuranceChanged: (overrides?: InsuranceRecalcOverrides) => void;
};

function focusField(wrap: HTMLElement | null) {
  wrap?.querySelector<HTMLElement>('input')?.focus();
}

/** انتخاب بیمه و کد خاص — فرم فشرده داخل کارت */
export function InsuranceSelection({ onInsuranceChanged }: InsuranceSelectionProps) {
  const editing = useReceptionStore((s) => s.editing);
  const insuranceId = useReceptionStore((s) => s.insuranceId);
  const additionalInsuranceId = useReceptionStore((s) => s.additionalInsuranceId);
  const percentage = useReceptionStore((s) => s.additionalInsurancePercentage);
  const coverage = useReceptionStore((s) => s.additionalInsuranceCoverage);
  const specialCodeValue = useReceptionStore((s) => s.specialCodeValue);
  const specialCodeName = useReceptionStore((s) => s.specialCodeName);
  const focusToken = useReceptionStore((s) => s.insuranceFocusToken);
  const setInsuranceId = useReceptionStore((s) => s.setInsuranceId);
  const setAdditionalInsuranceId = useReceptionStore((s) => s.setAdditionalInsuranceId);
  const setPercentage = useReceptionStore((s) => s.setAdditionalInsurancePercentage);
  const setCoverage = useReceptionStore((s) => s.setAdditionalInsuranceCoverage);
  const setSpecialCode = useReceptionStore((s) => s.setSpecialCode);

  const [codeDraft, setCodeDraft] = useState(specialCodeValue);
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
    setCodeDraft(specialCodeValue);
  }, [specialCodeValue]);

  useEffect(() => {
    if (focusToken > 0) {
      window.setTimeout(() => focusField(baseRef.current), 0);
    }
  }, [focusToken]);

  async function resolveSpecialCode(raw: string) {
    const code = raw.trim();
    if (!code || code === '0') {
      setSpecialCode(null, '', '');
      onInsuranceChanged({ special_code_id: null });
      return;
    }
    try {
      const sc = await fetchSpecialCodeByCode(code);
      if (!sc.is_active) {
        message.warning('کد خاص غیرفعال است');
        setSpecialCode(null, '', '');
        onInsuranceChanged({ special_code_id: null });
        return;
      }
      setSpecialCode(sc.id, sc.code, sc.name);
      onInsuranceChanged({ special_code_id: sc.id });
    } catch {
      message.error('کد خاص یافت نشد');
      setSpecialCode(null, '', '');
      onInsuranceChanged({ special_code_id: null });
    }
  }

  return (
    <Form layout="vertical" size="small">
      <Row gutter={[8, 0]}>
        <Col span={24}>
          <div ref={baseRef}>
            <Form.Item label="بیمه پایه" style={{ marginBottom: 8 }}>
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
        <Col span={24}>
          <div ref={suppRef}>
            <Form.Item label="بیمه تکمیلی" style={{ marginBottom: 8 }}>
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
        <Col span={12}>
          <div ref={percentageRef}>
            <Form.Item label="فرانشیز %" style={{ marginBottom: 8 }}>
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
        <Col span={12}>
          <div ref={coverageRef}>
            <Form.Item label="سقف تکمیلی" style={{ marginBottom: 8 }}>
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
        <Col span={12}>
          <Form.Item label="کد خاص" style={{ marginBottom: 8 }}>
            <Input
              disabled={!editing}
              value={codeDraft}
              placeholder="۰ = بدون"
              onChange={(e) => setCodeDraft(e.target.value)}
              onBlur={() => void resolveSpecialCode(codeDraft)}
              onPressEnter={() => void resolveSpecialCode(codeDraft)}
            />
          </Form.Item>
        </Col>
        <Col span={12}>
          <Form.Item label="نام کد خاص" style={{ marginBottom: 0 }}>
            <Typography.Text type={specialCodeName ? undefined : 'secondary'} ellipsis>
              {specialCodeName || '—'}
            </Typography.Text>
          </Form.Item>
        </Col>
      </Row>
    </Form>
  );
}
