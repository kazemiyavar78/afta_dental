import { httpClient } from '@/platform/api/httpClient';
import type {
  RevenueByOrganizationFilters,
  RevenueByOrganizationResponse,
  DoctorPerformanceBaseFilters,
  DoctorPerformanceByOrganizationFilters,
  DoctorPerformanceByOrganizationResponse,
  DoctorPerformanceByServiceFilters,
  DoctorPerformanceByServiceResponse,
  DoctorPerformancePatientListResponse,
  RevenueByDoctorsResponse,
  UserPerformanceByReceptionResponse,
  UserPerformanceByTransactionResponse,
  UserPerformanceFilters,
} from './types';

/**
 * دریافت گزارش درآمد به تفکیک سازمان با فیلترهای انتخاب‌شده.
 */
export async function fetchRevenueByOrganization(
  filters: RevenueByOrganizationFilters,
): Promise<RevenueByOrganizationResponse> {
  const params: Record<string, string | boolean> = {};

  if (filters.from_date) params.from_date = filters.from_date;
  if (filters.to_date) params.to_date = filters.to_date;
  if (filters.from_time) params.from_time = filters.from_time;
  if (filters.to_time) params.to_time = filters.to_time;
  if (filters.base_organization_ids?.length) {
    params.base_organization_ids = filters.base_organization_ids.join(',');
  }
  if (filters.supplementary_organization_ids?.length) {
    params.supplementary_organization_ids = filters.supplementary_organization_ids.join(',');
  }
  if (filters.detail_supplementary) {
    params.detail_supplementary = true;
  }
  if (filters.only_supplementary) {
    params.only_supplementary = true;
  }

  const { data } = await httpClient.get<RevenueByOrganizationResponse>(
    '/reports/revenue-by-organization',
    { params },
  );
  return data;
}

/** پارامترهای مشترک query گزارش کارکرد پزشکان را می‌سازد */
function buildDoctorPerformanceParams(
  filters: DoctorPerformanceBaseFilters,
): Record<string, string | boolean> {
  const params: Record<string, string | boolean> = {};
  if (filters.from_date) params.from_date = filters.from_date;
  if (filters.to_date) params.to_date = filters.to_date;
  if (filters.from_time) params.from_time = filters.from_time;
  if (filters.to_time) params.to_time = filters.to_time;
  if (filters.doctor_ids?.length) {
    params.doctor_ids = filters.doctor_ids.join(',');
  }
  if (filters.separate_doctors) {
    params.separate_doctors = true;
  }
  return params;
}

/**
 * دریافت گزارش کارکرد پزشکان به تفکیک سازمان.
 */
export async function fetchDoctorPerformanceByOrganization(
  filters: DoctorPerformanceByOrganizationFilters,
): Promise<DoctorPerformanceByOrganizationResponse> {
  const params: Record<string, string | boolean> = {
    ...buildDoctorPerformanceParams(filters),
  };
  if (filters.base_organization_ids?.length) {
    params.base_organization_ids = filters.base_organization_ids.join(',');
  }
  if (filters.supplementary_organization_ids?.length) {
    params.supplementary_organization_ids = filters.supplementary_organization_ids.join(',');
  }
  if (filters.detail_supplementary) params.detail_supplementary = true;
  if (filters.only_supplementary) params.only_supplementary = true;

  const { data } = await httpClient.get<DoctorPerformanceByOrganizationResponse>(
    '/reports/doctor-performance/by-organization',
    { params },
  );
  return data;
}

/**
 * دریافت گزارش کارکرد پزشکان به تفکیک خدمت.
 */
export async function fetchDoctorPerformanceByService(
  filters: DoctorPerformanceByServiceFilters,
): Promise<DoctorPerformanceByServiceResponse> {
  const params: Record<string, string | boolean> = {
    ...buildDoctorPerformanceParams(filters),
  };
  if (filters.service_ids?.length) {
    params.service_ids = filters.service_ids.join(',');
  }

  const { data } = await httpClient.get<DoctorPerformanceByServiceResponse>(
    '/reports/doctor-performance/by-service',
    { params },
  );
  return data;
}

/**
 * دریافت لیست بیماران پذیرش‌شده توسط پزشکان.
 */
export async function fetchDoctorPerformancePatientList(
  filters: DoctorPerformanceBaseFilters,
): Promise<DoctorPerformancePatientListResponse> {
  const { data } = await httpClient.get<DoctorPerformancePatientListResponse>(
    '/reports/doctor-performance/patient-list',
    { params: buildDoctorPerformanceParams(filters) },
  );
  return data;
}

/**
 * دریافت گزارش درآمد به تفکیک پزشکان.
 */
export async function fetchRevenueByDoctors(
  filters: DoctorPerformanceBaseFilters,
): Promise<RevenueByDoctorsResponse> {
  const { data } = await httpClient.get<RevenueByDoctorsResponse>(
    '/reports/doctor-performance/revenue-by-doctors',
    { params: buildDoctorPerformanceParams(filters) },
  );
  return data;
}

/** پارامترهای مشترک query گزارش کارکرد کاربران را می‌سازد */
function buildUserPerformanceParams(
  filters: UserPerformanceFilters,
): Record<string, string | boolean> {
  const params: Record<string, string | boolean> = {};
  if (filters.from_date) params.from_date = filters.from_date;
  if (filters.to_date) params.to_date = filters.to_date;
  if (filters.from_time) params.from_time = filters.from_time;
  if (filters.to_time) params.to_time = filters.to_time;
  if (filters.user_ids?.length) {
    params.user_ids = filters.user_ids.join(',');
  }
  if (filters.separate_by_patient) {
    params.separate_by_patient = true;
  }
  return params;
}

/**
 * دریافت گزارش کارکرد کاربران بر اساس پذیرش انجام‌شده.
 */
export async function fetchUserPerformanceByReception(
  filters: UserPerformanceFilters,
): Promise<UserPerformanceByReceptionResponse> {
  const { data } = await httpClient.get<UserPerformanceByReceptionResponse>(
    '/reports/user-performance/by-reception',
    { params: buildUserPerformanceParams(filters) },
  );
  return data;
}

/**
 * دریافت گزارش کارکرد کاربران بر اساس مبلغ دریافتی/پرداختی.
 */
export async function fetchUserPerformanceByTransaction(
  filters: UserPerformanceFilters,
): Promise<UserPerformanceByTransactionResponse> {
  const { data } = await httpClient.get<UserPerformanceByTransactionResponse>(
    '/reports/user-performance/by-transaction',
    { params: buildUserPerformanceParams(filters) },
  );
  return data;
}
