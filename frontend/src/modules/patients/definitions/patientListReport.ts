import type { ReportDefinition } from '@/platform/reports';
import { formatDate } from '@/platform/reports';
import type { Patient } from '../types';

/** برچسب جنسیت */
function sexLabel(sex: boolean): string {
  return sex ? 'مرد' : 'زن';
}

/** ReportDefinition فهرست بیماران */
export const patientListReportDefinition: ReportDefinition<Patient> = {
  id: 'patient-list',
  title: 'فهرست بیماران',
  page: { size: 'A4', orientation: 'portrait', margin: 12 },
  columns: [
    { key: 'file_number', title: 'شماره پرونده', width: 9, format: 'text' },
    { key: 'first_name', title: 'نام', width: 9, format: 'text' },
    { key: 'last_name', title: 'نام خانوادگی', width: 11, format: 'text' },
    { key: 'national_code', title: 'کد ملی', width: 10, format: 'text' },
    {
      key: 'birth_date',
      title: 'تاریخ تولد',
      width: 9,
      format: 'date',
      formatter: (v) => formatDate(String(v)),
      printFormatter: (v) => formatDate(String(v)),
    },
    {
      key: 'sex',
      title: 'جنسیت',
      width: 6,
      formatter: (v) => sexLabel(Boolean(v)),
      printFormatter: (v) => sexLabel(Boolean(v)),
      excelFormatter: (v) => sexLabel(Boolean(v)),
    },
    { key: 'mobile_phone_number', title: 'موبایل', width: 10, format: 'text' },
    { key: 'home_phone_number', title: 'تلفن منزل', width: 10, format: 'text' },
    { key: 'address', title: 'آدرس', width: 22, align: 'right', format: 'text' },
  ],
  header: { showGeneratedAt: true },
  footer: { text: 'صفحات به‌صورت خودکار بر اساس اندازه A4 شکسته می‌شوند.' },
  toolbar: { print: true, pdf: true, excel: true, pageSetup: true },
  export: { sheetName: 'بیماران' },
};
