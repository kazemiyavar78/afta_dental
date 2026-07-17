import { create } from 'zustand';
import {
  emptyPatient,
  emptyServiceLine,
  type PatientFormState,
  type ReceptionDetail,
  type ServiceLineState,
} from '../types';

type ReceptionStore = {
  /** شناسه پذیرش جاری؛ null یعنی فرم جدید */
  receptionId: number | null;
  status: string;
  deleted: boolean;
  editing: boolean;
  isNew: boolean;
  patient: PatientFormState;
  insuranceId: number | null;
  additionalInsuranceId: number | null;
  additionalInsurancePercentage: number | null;
  additionalInsuranceCoverage: number | null;
  doctorId: number | null;
  doctorName: string;
  doctorMedicalCode: string | null;
  assistantId: number | null;
  assistantName: string;
  bookingDate: string | null;
  receptionDate: string;
  description: string;
  discount: number;
  referralCode: number | null;
  services: ServiceLineState[];
  /** رفرنس برای فوکوس به بیمه پایه در خطای E-006 */
  insuranceFocusToken: number;
  /** کلید سطری که باید فوکوس انتخاب خدمت بگیرد */
  serviceFocusKey: string | null;
  /** کلید سطری که باید فوکوس فیلد تعداد بگیرد (مثلاً بعد از Enter روی توضیحات ردیف بالایی) */
  quantityFocusKey: string | null;
  /** توکن فوکوس به فیلد تاریخ اعتبار دفترچه پس از انتخاب دستیار */
  bookingDateFocusToken: number;

  resetNew: () => void;
  loadFromDetail: (detail: ReceptionDetail) => void;
  setPatient: (patch: Partial<PatientFormState>) => void;
  setInsuranceId: (id: number | null) => void;
  setAdditionalInsuranceId: (id: number | null) => void;
  setAdditionalInsurancePercentage: (v: number | null) => void;
  setAdditionalInsuranceCoverage: (v: number | null) => void;
  setDoctor: (id: number | null, name: string, medicalCode: string | null) => void;
  setAssistant: (id: number | null, name: string) => void;
  setBookingDate: (v: string | null) => void;
  setReceptionDate: (v: string) => void;
  setDescription: (v: string) => void;
  setDiscount: (v: number) => void;
  setReferralCode: (v: number | null) => void;
  setEditing: (v: boolean) => void;
  setServices: (services: ServiceLineState[]) => void;
  updateServiceLine: (key: string, patch: Partial<ServiceLineState>) => void;
  addServiceLine: () => void;
  removeServiceLine: (key: string) => void;
  requestInsuranceFocus: () => void;
  clearServiceFocus: () => void;
  /** درخواست فوکوس روی فیلد تعداد سطر مشخص */
  requestQuantityFocus: (key: string) => void;
  clearQuantityFocus: () => void;
  requestBookingDateFocus: () => void;
  hasOrganization: () => boolean;
};

const today = () => new Date().toISOString().slice(0, 10);

/** استور مرکزی وضعیت فرم پذیرش */
export const useReceptionStore = create<ReceptionStore>((set, get) => ({
  receptionId: null,
  status: 'draft',
  deleted: false,
  editing: true,
  isNew: true,
  patient: emptyPatient(),
  insuranceId: null,
  additionalInsuranceId: null,
  additionalInsurancePercentage: null,
  additionalInsuranceCoverage: null,
  doctorId: null,
  doctorName: '',
  doctorMedicalCode: null,
  assistantId: null,
  assistantName: '',
  bookingDate: null,
  receptionDate: today(),
  description: '',
  discount: 0,
  referralCode: null,
  services: [],
  insuranceFocusToken: 0,
  serviceFocusKey: null,
  quantityFocusKey: null,
  bookingDateFocusToken: 0,

  resetNew: () =>
    set({
      receptionId: null,
      status: 'draft',
      deleted: false,
      editing: true,
      isNew: true,
      patient: emptyPatient(),
      insuranceId: null,
      additionalInsuranceId: null,
      additionalInsurancePercentage: null,
      additionalInsuranceCoverage: null,
      doctorId: null,
      doctorName: '',
      doctorMedicalCode: null,
      assistantId: null,
      assistantName: '',
      bookingDate: null,
      receptionDate: today(),
      description: '',
      discount: 0,
      referralCode: null,
      services: [],
      serviceFocusKey: null,
      quantityFocusKey: null,
    }),

  loadFromDetail: (detail) => {
    if (detail.empty || !detail.id) {
      get().resetNew();
      return;
    }

    const patient = detail.patient
      ? {
          id: detail.patient.id,
          first_name: detail.patient.first_name,
          last_name: detail.patient.last_name,
          national_code: detail.patient.national_code,
          birth_date: detail.patient.birth_date,
          address: detail.patient.address,
          home_phone_number: detail.patient.home_phone_number,
          mobile_phone_number: detail.patient.mobile_phone_number,
          file_number: detail.patient.file_number,
          sex: detail.patient.sex,
          isExisting: true,
        }
      : emptyPatient();

    set({
      receptionId: detail.id,
      status: detail.status,
      deleted: detail.deleted,
      editing: false,
      isNew: false,
      patient,
      insuranceId: detail.insurance_id,
      additionalInsuranceId: detail.additional_insurance_id,
      additionalInsurancePercentage: detail.additional_insurance_percentage,
      additionalInsuranceCoverage: detail.additional_insurance_coverage,
      doctorId: detail.doctor_id,
      doctorName: detail.doctor_name ?? '',
      doctorMedicalCode: detail.doctor_medical_code ?? null,
      assistantId: detail.assistant_id,
      assistantName: detail.assistant_name ?? '',
      bookingDate: detail.booking_date,
      receptionDate: detail.reception_date || today(),
      description: detail.description ?? '',
      discount: detail.discount ?? 0,
      referralCode: detail.referral_code,
      services: (detail.services ?? []).map((s) => ({
        key: String(s.id || `${s.service_id}-${Math.random()}`),
        service_id: s.service_id,
        service_code: '',
        service_name: s.service_name,
        quantity: s.quantity,
        service_amount: s.service_amount,
        service_tariff: s.service_tariff,
        service_organization_share: s.service_organization_share,
        service_supplementary_insurance_share: s.service_supplementary_insurance_share,
        service_subsidy_share: s.service_subsidy_share,
        service_description: s.service_description,
        teeth_number: s.teeth_number,
        teeth_direction: s.teeth_direction,
        has_dental_direction: s.has_dental_direction,
        has_tooth: s.has_tooth,
      })),
    });
  },

  setPatient: (patch) => set({ patient: { ...get().patient, ...patch } }),
  setInsuranceId: (insuranceId) => set({ insuranceId }),
  setAdditionalInsuranceId: (additionalInsuranceId) => set({ additionalInsuranceId }),
  setAdditionalInsurancePercentage: (additionalInsurancePercentage) =>
    set({ additionalInsurancePercentage }),
  setAdditionalInsuranceCoverage: (additionalInsuranceCoverage) =>
    set({ additionalInsuranceCoverage }),
  setDoctor: (doctorId, doctorName, doctorMedicalCode) =>
    set({ doctorId, doctorName, doctorMedicalCode }),
  setAssistant: (assistantId, assistantName) => set({ assistantId, assistantName }),
  setBookingDate: (bookingDate) => set({ bookingDate }),
  setReceptionDate: (receptionDate) => set({ receptionDate }),
  setDescription: (description) => set({ description }),
  setDiscount: (discount) => set({ discount }),
  setReferralCode: (referralCode) => set({ referralCode }),
  setEditing: (editing) => set({ editing }),
  setServices: (services) => set({ services }),
  updateServiceLine: (key, patch) =>
    set({
      services: get().services.map((s) => (s.key === key ? { ...s, ...patch } : s)),
    }),
  addServiceLine: () => {
    const line = emptyServiceLine();
    set({ services: [...get().services, line], serviceFocusKey: line.key });
  },
  removeServiceLine: (key) =>
    set({ services: get().services.filter((s) => s.key !== key) }),
  requestInsuranceFocus: () =>
    set({ insuranceFocusToken: get().insuranceFocusToken + 1 }),
  clearServiceFocus: () => set({ serviceFocusKey: null }),
  requestQuantityFocus: (key) => set({ quantityFocusKey: key }),
  clearQuantityFocus: () => set({ quantityFocusKey: null }),
  requestBookingDateFocus: () =>
    set({ bookingDateFocusToken: get().bookingDateFocusToken + 1 }),
  hasOrganization: () => get().insuranceId != null || get().additionalInsuranceId != null,
}));
