// این ماژول نمونه است؛ باقی فیلدها/صفحات طبق همین الگو در فازهای بعدی اضافه می‌شوند.

import { z } from 'zod';

export const receptionSchema = z.object({
  patient_name: z.string().min(1, 'نام بیمار الزامی است'),
  doctor_id: z.number().min(1, 'پزشک الزامی است'),
  reception_date: z.string().min(1, 'تاریخ الزامی است'),
});

export type ReceptionFormValues = z.infer<typeof receptionSchema>;
