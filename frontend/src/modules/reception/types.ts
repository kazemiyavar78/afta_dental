/** انواع داده ماژول پذیرش بیمار */

export type PatientFormState = {
  id?: number;
  first_name: string;
  last_name: string;
  national_code: string;
  birth_date: string;
  address: string | null;
  home_phone_number: string | null;
  mobile_phone_number: string | null;
  file_number: string;
  sex: boolean;
  is_foreign_national: boolean;
  /** بیمار از دیتابیس بارگذاری شده و فیلدهای هویتی قفل‌اند */
  isExisting: boolean;
};

export type ServiceLineState = {
  key: string;
  service_id: number;
  service_code: string;
  service_name: string;
  quantity: number;
  service_amount: number;
  service_tariff: number;
  service_organization_share: number;
  service_supplementary_insurance_share: number;
  service_subsidy_share: number;
  service_description: string;
  teeth_number: number | null;
  teeth_direction: number | null;
  has_dental_direction: boolean;
  has_tooth: boolean;
};

export type ReceptionDetail = {
  id: number;
  patient_id: number;
  patient?: PatientFormState & { id: number };
  insurance_id: number | null;
  additional_insurance_id: number | null;
  special_code_id: number | null;
  special_code_name?: string;
  special_code_value?: string;
  doctor_id: number | null;
  assistant_id: number | null;
  doctor_name?: string;
  doctor_medical_code?: string | null;
  assistant_name?: string;
  booking_date: string | null;
  reception_date: string;
  status: string;
  description: string;
  discount: number;
  referral_code: number | null;
  additional_insurance_coverage: number | null;
  additional_insurance_percentage: number | null;
  reception_ended: boolean;
  photo_count: number;
  registered_by_id: number | null;
  services: Array<{
    id: number;
    service_id: number;
    service_name: string;
    quantity: number;
    service_amount: number;
    service_tariff: number;
    service_organization_share: number;
    service_supplementary_insurance_share: number;
    service_subsidy_share: number;
    service_description: string;
    teeth_number: number | null;
    teeth_direction: number | null;
    has_dental_direction: boolean;
    has_tooth: boolean;
  }>;
  deleted: boolean;
  /** اگر true باشد هیچ پذیرشی وجود ندارد و فرم خالی باید نمایش داده شود */
  empty?: boolean;
};

export type UpsertReceptionPayload = {
  patient: {
    id?: number;
    first_name: string;
    last_name: string;
    national_code: string;
    birth_date: string;
    address?: string | null;
    home_phone_number?: string | null;
    mobile_phone_number?: string | null;
    file_number: string;
    sex: boolean;
    is_foreign_national: boolean;
  };
  insurance_id: number | null;
  additional_insurance_id: number | null;
  special_code_id: number | null;
  doctor_id: number | null;
  assistant_id: number | null;
  booking_date: string | null;
  reception_date: string;
  description: string;
  discount: number;
  referral_code: number | null;
  additional_insurance_coverage: number | null;
  additional_insurance_percentage: number | null;
  services: Array<{
    service_id: number;
    service_code?: string;
    quantity: number;
    teeth_number: number | null;
    teeth_direction: number | null;
    service_description: string;
  }>;
  save: boolean;
};

export type CalculatePayload = {
  insurance_id: number | null;
  additional_insurance_id: number | null;
  special_code_id: number | null;
  additional_insurance_coverage: number | null;
  additional_insurance_percentage: number | null;
  services: Array<{
    service_id: number;
    service_code?: string;
    quantity: number;
    teeth_number: number | null;
    teeth_direction: number | null;
    service_description: string;
  }>;
};

export type CalculatedServiceLine = {
  service_id: number;
  service_code: string;
  service_name: string;
  quantity: number;
  service_amount: number;
  service_tariff: number;
  service_organization_share: number;
  service_supplementary_insurance_share: number;
  service_subsidy_share: number;
  service_description: string;
  teeth_number: number | null;
  teeth_direction: number | null;
  has_dental_direction: boolean;
  has_tooth: boolean;
};

export type DoctorUser = {
  id: number;
  name: string;
  family: string;
  medical_code: string | null;
  is_active: boolean;
  user_type: string;
};

export type EndReceptionResult = {
  success: boolean;
  reception_ended: boolean;
  previous_reception_id?: number | null;
  regulation_descriptions?: string[];
  required_photo_count: number;
  uploaded_photo_count: number;
  message: string;
};

export type PatientServiceHistoryItem = {
  reception_id: number;
  reception_date: string;
  insurance_name: string;
  additional_insurance_name: string;
  cash_amount: number;
  service_names: string[];
};

/** سهم صندوق یک سطر: نرخ − سهم سازمان − سهم تکمیلی − یارانه */
export function lineCashAmount(line: {
  service_amount: number;
  service_organization_share: number;
  service_supplementary_insurance_share: number;
  service_subsidy_share: number;
}): number {
  return (
    Number(line.service_amount ?? 0) -
    Number(line.service_organization_share ?? 0) -
    Number(line.service_supplementary_insurance_share ?? 0) -
    Number(line.service_subsidy_share ?? 0)
  );
}

/** حالت خالی فرم پذیرش جدید */
export function emptyPatient(): PatientFormState {
  return {
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
}

/** سطر خدمت خالی */
export function emptyServiceLine(): ServiceLineState {
  return {
    key: `${Date.now()}-${Math.random()}`,
    service_id: 0,
    service_code: '',
    service_name: '',
    quantity: 1,
    service_amount: 0,
    service_tariff: 0,
    service_organization_share: 0,
    service_supplementary_insurance_share: 0,
    service_subsidy_share: 0,
    service_description: '',
    teeth_number: null,
    teeth_direction: null,
    has_dental_direction: false,
    has_tooth: false,
  };
}
