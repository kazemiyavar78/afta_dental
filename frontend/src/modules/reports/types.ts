/** فیلترهای گزارش درآمد به تفکیک سازمان */
export type RevenueByOrganizationFilters = {
  from_date?: string;
  to_date?: string;
  from_time?: string;
  to_time?: string;
  base_organization_ids?: number[];
  supplementary_organization_ids?: number[];
  detail_supplementary?: boolean;
  only_supplementary?: boolean;
};

/** یک سطر تجمیعی گزارش درآمد */
export type RevenueByOrganizationRow = {
  row_label: string;
  base_organization_id: number;
  base_organization_name: string;
  supplementary_organization_id?: number | null;
  supplementary_organization_name?: string;
  service_amount: number;
  service_tariff: number;
  organization_share: number;
  supplementary_share: number;
  subsidy_share: number;
  payable_amount: number;
  reception_count: number;
};

/** متادیتای فیلتر اعمال‌شده */
export type RevenueByOrganizationMeta = {
  from_date: string;
  to_date: string;
  from_time: string;
  to_time: string;
  detail_supplementary: boolean;
  has_supplementary_filter: boolean;
  only_supplementary: boolean;
  free_organization_name?: string;
};

/** پاسخ API گزارش درآمد به تفکیک سازمان */
export type RevenueByOrganizationResponse = {
  rows: RevenueByOrganizationRow[];
  summary: RevenueByOrganizationRow;
  meta: RevenueByOrganizationMeta;
};

/** مبلغ را با جداکننده فارسی نمایش می‌دهد */
export { formatCurrency as formatMoney } from '@/platform/reports';

/** نوع گزارش کارکرد پزشکان */
export type DoctorPerformanceReportType =
  | 'by_organization'
  | 'by_service'
  | 'patient_list'
  | 'revenue_by_doctors';

/** فیلترهای مشترک گزارش کارکرد پزشکان */
export type DoctorPerformanceBaseFilters = {
  from_date?: string;
  to_date?: string;
  from_time?: string;
  to_time?: string;
  doctor_ids?: number[];
  separate_doctors?: boolean;
};

/** فیلترهای گزارش کارکرد به تفکیک سازمان */
export type DoctorPerformanceByOrganizationFilters = DoctorPerformanceBaseFilters & {
  base_organization_ids?: number[];
  supplementary_organization_ids?: number[];
  detail_supplementary?: boolean;
  only_supplementary?: boolean;
};

/** فیلترهای گزارش کارکرد به تفکیک خدمت */
export type DoctorPerformanceByServiceFilters = DoctorPerformanceBaseFilters & {
  service_ids?: number[];
};

/** جمع مالی مشترک گزارش کارکرد پزشکان */
export type DoctorPerformanceSummaryRow = {
  service_amount: number;
  service_tariff: number;
  organization_share: number;
  supplementary_share: number;
  subsidy_share: number;
  payable_amount: number;
  reception_count?: number;
  row_label?: string;
};

/** سطر گزارش کارکرد به تفکیک سازمان */
export type DoctorPerformanceByOrganizationRow = {
  row_label: string;
  base_organization_id: number;
  base_organization_name: string;
  supplementary_organization_id?: number | null;
  supplementary_organization_name?: string;
  service_amount: number;
  service_tariff: number;
  organization_share: number;
  supplementary_share: number;
  subsidy_share: number;
  payable_amount: number;
  reception_count: number;
};

/** سطر گزارش کارکرد به تفکیک خدمت */
export type DoctorPerformanceByServiceRow = {
  service_id: number;
  service_name: string;
  service_amount: number;
  service_tariff: number;
  organization_share: number;
  supplementary_share: number;
  subsidy_share: number;
  payable_amount: number;
  reception_count: number;
};

/** سطر لیست بیماران پزشکان */
export type DoctorPerformancePatientRow = {
  patient_id: number;
  patient_name: string;
  national_code: string;
  service_amount: number;
  service_tariff: number;
  organization_share: number;
  supplementary_share: number;
  subsidy_share: number;
  payable_amount: number;
  services_received: string;
};

/** سطر گزارش درآمد به تفکیک پزشک */
export type RevenueByDoctorsRow = {
  doctor_id: number;
  medical_code: string;
  doctor_name: string;
  service_amount: number;
  service_tariff: number;
  organization_share: number;
  supplementary_share: number;
  subsidy_share: number;
  payable_amount: number;
  reception_count: number;
};

/** متادیتای گزارش کارکرد پزشکان */
export type DoctorPerformanceMeta = {
  report_type: DoctorPerformanceReportType;
  from_date: string;
  to_date: string;
  from_time: string;
  to_time: string;
  doctor_names: string[];
  separate_doctors?: boolean;
  detail_supplementary?: boolean;
  has_supplementary_filter?: boolean;
  only_supplementary?: boolean;
  free_organization_name?: string;
  service_names?: string[];
};

/** بخش گزارش مربوط به یک پزشک */
export type DoctorPerformanceDoctorSection<
  T extends
    | DoctorPerformanceByOrganizationRow
    | DoctorPerformanceByServiceRow
    | DoctorPerformancePatientRow,
> = {
  doctor_id: number;
  doctor_name: string;
  rows: T[];
  summary: DoctorPerformanceSummaryRow;
};

/** پاسخ API گزارش کارکرد به تفکیک سازمان */
export type DoctorPerformanceByOrganizationResponse = {
  rows: DoctorPerformanceByOrganizationRow[];
  summary: DoctorPerformanceSummaryRow;
  sections?: DoctorPerformanceDoctorSection<DoctorPerformanceByOrganizationRow>[];
  meta: DoctorPerformanceMeta;
};

/** پاسخ API گزارش کارکرد به تفکیک خدمت */
export type DoctorPerformanceByServiceResponse = {
  rows: DoctorPerformanceByServiceRow[];
  summary: DoctorPerformanceSummaryRow;
  sections?: DoctorPerformanceDoctorSection<DoctorPerformanceByServiceRow>[];
  meta: DoctorPerformanceMeta;
};

/** پاسخ API لیست بیماران پزشکان */
export type DoctorPerformancePatientListResponse = {
  rows: DoctorPerformancePatientRow[];
  summary: DoctorPerformanceSummaryRow;
  sections?: DoctorPerformanceDoctorSection<DoctorPerformancePatientRow>[];
  meta: DoctorPerformanceMeta;
};

/** پاسخ API گزارش درآمد به تفکیک پزشکان */
export type RevenueByDoctorsResponse = {
  rows: RevenueByDoctorsRow[];
  summary: DoctorPerformanceSummaryRow;
  meta: DoctorPerformanceMeta;
};

/** نوع گزارش کارکرد کاربران */
export type UserPerformanceReportType = 'by_reception' | 'by_transaction';

/** فیلترهای گزارش کارکرد کاربران */
export type UserPerformanceFilters = {
  from_date?: string;
  to_date?: string;
  from_time?: string;
  to_time?: string;
  user_ids?: number[];
  separate_by_patient?: boolean;
};

/** سطر گزارش کارکرد بر اساس پذیرش */
export type UserPerformanceReceptionRow = {
  user_id: number;
  user_name: string;
  service_amount: number;
  service_tariff: number;
  organization_share: number;
  supplementary_share: number;
  subsidy_share: number;
  payable_amount: number;
  reception_count: number;
};

/** سطر ریز پذیرش در گزارش کارکرد کاربران */
export type UserPerformanceReceptionDetailRow = {
  reception_id: number;
  patient_name: string;
  national_code: string;
  service_amount: number;
  service_tariff: number;
  organization_share: number;
  supplementary_share: number;
  subsidy_share: number;
  payable_amount: number;
  services_received: string;
};

/** سطر گزارش کارکرد بر اساس تراکنش مالی */
export type UserPerformanceTransactionRow = {
  user_id: number;
  user_name: string;
  total_received: number;
  total_paid: number;
  received_tx_count: number;
  paid_tx_count: number;
  total_tx_count: number;
};

/** سطر ریز تراکنش مالی در گزارش کارکرد کاربران */
export type UserPerformanceTransactionDetailRow = {
  transaction_id: number;
  file_number: string;
  patient_name: string;
  national_code: string;
  received: number;
  paid: number;
};

/** جمع گزارش پذیرش کاربران */
export type UserPerformanceReceptionSummary = {
  service_amount: number;
  service_tariff: number;
  organization_share: number;
  supplementary_share: number;
  subsidy_share: number;
  payable_amount: number;
  reception_count?: number;
  row_label?: string;
};

/** جمع گزارش تراکنش کاربران */
export type UserPerformanceTransactionSummary = {
  total_received: number;
  total_paid: number;
  received_tx_count?: number;
  paid_tx_count?: number;
  total_tx_count?: number;
  row_label?: string;
};

/** متادیتای گزارش کارکرد کاربران */
export type UserPerformanceMeta = {
  report_type: UserPerformanceReportType;
  from_date: string;
  to_date: string;
  from_time: string;
  to_time: string;
  user_names: string[];
  separate_by_patient?: boolean;
};

/** بخش گزارش مربوط به یک کاربر — پذیرش */
export type UserPerformanceReceptionSection = {
  user_id: number;
  user_name: string;
  rows: UserPerformanceReceptionDetailRow[];
  summary: UserPerformanceReceptionSummary;
};

/** بخش گزارش مربوط به یک کاربر — تراکنش */
export type UserPerformanceTransactionSection = {
  user_id: number;
  user_name: string;
  rows: UserPerformanceTransactionDetailRow[];
  summary: UserPerformanceTransactionSummary;
};

/** پاسخ API گزارش کارکرد بر اساس پذیرش */
export type UserPerformanceByReceptionResponse = {
  rows: UserPerformanceReceptionRow[];
  summary: UserPerformanceReceptionSummary;
  sections?: UserPerformanceReceptionSection[];
  meta: UserPerformanceMeta;
};

/** پاسخ API گزارش کارکرد بر اساس تراکنش مالی */
export type UserPerformanceByTransactionResponse = {
  rows: UserPerformanceTransactionRow[];
  summary: UserPerformanceTransactionSummary;
  sections?: UserPerformanceTransactionSection[];
  meta: UserPerformanceMeta;
};
