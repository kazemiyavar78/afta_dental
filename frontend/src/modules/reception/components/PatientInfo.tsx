import { useRef } from 'react';
import { Col, Form, Input, Radio, Row, Switch, message } from 'antd';
import type { InputRef } from 'antd/es/input';
import { JalaliDatePicker } from '@/platform/components/JalaliDatePicker/JalaliDatePicker';
import { fetchPatients } from '@/modules/patients/api';
import { useReceptionStore } from '../store/receptionStore';

/** فوکوس به اولین input داخل یک ظرف */
function focusInside(el: HTMLElement | null) {
  el?.querySelector<HTMLElement>('input, textarea, .ant-select-selector')?.focus();
}

/** بخش اطلاعات بیمار — فرم فشرده دو ستونه داخل کارت */
export function PatientInfo() {
  const patient = useReceptionStore((s) => s.patient) ?? {
    first_name: '',
    last_name: '',
    national_code: '',
    birth_date: '',
    address: null,
    home_phone_number: null,
    mobile_phone_number: null,
    file_number: '',
    sex: true,
    is_foreign_national: false,
    isExisting: false,
  };
  const editing = useReceptionStore((s) => s.editing);
  const isNew = useReceptionStore((s) => s.isNew);
  const setPatient = useReceptionStore((s) => s.setPatient);
  const searching = useRef(false);
  const sexWrapRef = useRef<HTMLDivElement>(null);
  const birthWrapRef = useRef<HTMLDivElement>(null);
  const addressRef = useRef<InputRef>(null);
  const firstNameRef = useRef<InputRef>(null);

  const identityLocked = patient.isExisting || (!isNew && !editing);
  const fieldsLocked = !editing || patient.isExisting;

  async function lookup(kind: 'national_code' | 'file_number', value: string) {
    const trimmed = value.trim();
    if (!trimmed || searching.current) return;
    searching.current = true;
    try {
      const list = await fetchPatients(
        kind === 'national_code' ? { national_code: trimmed } : { file_number: trimmed },
      );
      const found = list.find((p) =>
        kind === 'national_code' ? p.national_code === trimmed : p.file_number === trimmed,
      );
      if (found) {
        setPatient({
          id: found.id,
          first_name: found.first_name,
          last_name: found.last_name,
          national_code: found.national_code,
          birth_date: found.birth_date,
          address: found.address,
          home_phone_number: found.home_phone_number,
          mobile_phone_number: found.mobile_phone_number,
          file_number: found.file_number,
          sex: found.sex,
          is_foreign_national: found.is_foreign_national ?? false,
          isExisting: true,
        });
        message.success('اطلاعات بیمار بارگذاری شد');
      } else {
        setPatient({
          isExisting: false,
          id: undefined,
          ...(kind === 'national_code'
            ? { national_code: trimmed }
            : { file_number: trimmed }),
        });
        message.info('بیمار یافت نشد؛ می‌توانید اطلاعات جدید وارد کنید');
      }
    } catch {
      message.error('خطا در جستجوی بیمار');
    } finally {
      searching.current = false;
    }
  }

  function handleNationalCodeChange(raw: string) {
    if (patient.is_foreign_national) {
      setPatient({ national_code: raw.slice(0, 20) });
      return;
    }
    const digits = raw.replace(/\D/g, '').slice(0, 10);
    setPatient({ national_code: digits });
    if (digits.length === 10 && editing && !patient.isExisting) {
      void lookup('national_code', digits).then(() => {
        const latest = useReceptionStore.getState().patient;
        if (!latest.isExisting) {
          window.setTimeout(() => focusInside(sexWrapRef.current), 50);
        }
      });
    }
  }

  function handlePhoneChange(raw: string) {
    const digits = raw.replace(/\D/g, '').slice(0, 11);
    setPatient({ mobile_phone_number: digits || null });
    if (digits.length === 11 && !fieldsLocked) {
      window.setTimeout(() => addressRef.current?.focus(), 0);
    }
  }

  return (
    <Form layout="vertical" size="small" requiredMark="optional">
      <Row gutter={[8, 0]}>
        <Col span={12}>
          <Form.Item label="شماره پرونده" required style={{ marginBottom: 8 }}>
            <Input
              value={patient.file_number}
              disabled={identityLocked}
              onChange={(e) => setPatient({ file_number: e.target.value })}
              onBlur={(e) => {
                if (editing && !patient.isExisting) lookup('file_number', e.target.value);
              }}
            />
          </Form.Item>
        </Col>
        <Col span={12}>
          <Form.Item
            label={patient.is_foreign_national ? 'کد شناسایی' : 'کد ملی'}
            required={!patient.is_foreign_national}
            style={{ marginBottom: 8 }}
          >
            <Input
              value={patient.national_code}
              disabled={identityLocked}
              maxLength={patient.is_foreign_national ? 20 : 10}
              inputMode={patient.is_foreign_national ? 'text' : 'numeric'}
              placeholder={patient.is_foreign_national ? 'اختیاری' : '۱۰ رقم'}
              onChange={(e) => handleNationalCodeChange(e.target.value)}
              onBlur={(e) => {
                if (
                  editing &&
                  !patient.isExisting &&
                  !patient.is_foreign_national &&
                  e.target.value.replace(/\D/g, '').length === 10
                ) {
                  void lookup('national_code', e.target.value);
                }
              }}
            />
          </Form.Item>
        </Col>
        <Col span={12}>
          <Form.Item label="اتباع" style={{ marginBottom: 8 }}>
            <Switch
              size="small"
              disabled={identityLocked}
              checked={patient.is_foreign_national}
              checkedChildren="اتباع"
              unCheckedChildren="ایرانی"
              onChange={(v) => setPatient({ is_foreign_national: v, national_code: '' })}
            />
          </Form.Item>
        </Col>
        <Col span={12}>
          <div ref={sexWrapRef}>
            <Form.Item label="جنسیت" required style={{ marginBottom: 8 }}>
              <Radio.Group
                size="small"
                optionType="button"
                buttonStyle="solid"
                disabled={fieldsLocked}
                value={patient.sex ? 'male' : 'female'}
                options={[
                  { value: 'male', label: 'مرد' },
                  { value: 'female', label: 'زن' },
                ]}
                onChange={(e) => {
                  setPatient({ sex: e.target.value === 'male' });
                  window.setTimeout(() => focusInside(birthWrapRef.current), 0);
                }}
              />
            </Form.Item>
          </div>
        </Col>
        <Col span={12}>
          <div ref={birthWrapRef}>
            <Form.Item label="تاریخ تولد" required style={{ marginBottom: 8 }}>
              <JalaliDatePicker
                value={patient.birth_date || null}
                disabled={fieldsLocked}
                style={{ width: '100%' }}
                onChange={(v) => {
                  setPatient({ birth_date: v ?? '' });
                  window.setTimeout(() => firstNameRef.current?.focus(), 0);
                }}
              />
            </Form.Item>
          </div>
        </Col>
        <Col span={12}>
          <Form.Item label="نام" required style={{ marginBottom: 8 }}>
            <Input
              ref={firstNameRef}
              value={patient.first_name}
              disabled={fieldsLocked}
              onChange={(e) => setPatient({ first_name: e.target.value })}
            />
          </Form.Item>
        </Col>
        <Col span={12}>
          <Form.Item label="نام خانوادگی" required style={{ marginBottom: 8 }}>
            <Input
              value={patient.last_name}
              disabled={fieldsLocked}
              onChange={(e) => setPatient({ last_name: e.target.value })}
            />
          </Form.Item>
        </Col>
        <Col span={12}>
          <Form.Item label="تلفن" style={{ marginBottom: 8 }}>
            <Input
              value={patient.mobile_phone_number ?? ''}
              disabled={fieldsLocked}
              maxLength={11}
              inputMode="numeric"
              placeholder="09xxxxxxxxx"
              onChange={(e) => handlePhoneChange(e.target.value)}
            />
          </Form.Item>
        </Col>
        <Col span={24}>
          <Form.Item label="آدرس" style={{ marginBottom: 0 }}>
            <Input
              ref={addressRef}
              value={patient.address ?? ''}
              disabled={fieldsLocked}
              onChange={(e) => setPatient({ address: e.target.value || null })}
            />
          </Form.Item>
        </Col>
      </Row>
    </Form>
  );
}
