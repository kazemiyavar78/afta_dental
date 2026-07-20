import { z } from 'zod';
import { isValidIranianNationalCode } from '@/platform/utils/nationalCode';

/** اسکیمای ایجاد/ویرایش بیمار */
export const patientSchema = z
  .object({
    first_name: z.string().min(1, 'نام الزامی است'),
    last_name: z.string().min(1, 'نام خانوادگی الزامی است'),
    national_code: z.string(),
    birth_date: z.string().min(1, 'تاریخ تولد الزامی است'),
    address: z.string().optional().nullable(),
    home_phone_number: z.string().optional().nullable(),
    mobile_phone_number: z.string().optional().nullable(),
    file_number: z.string().min(1, 'شماره پرونده الزامی است'),
    sex: z.boolean(),
    is_foreign_national: z.boolean(),
  })
  .superRefine((data, ctx) => {
    const code = (data.national_code ?? '').trim();
    if (data.is_foreign_national) {
      if (code.length > 20) {
        ctx.addIssue({
          code: z.ZodIssueCode.custom,
          message: 'کد شناسایی اتباع حداکثر ۲۰ کاراکتر است',
          path: ['national_code'],
        });
      }
      return;
    }
    if (!code) {
      ctx.addIssue({
        code: z.ZodIssueCode.custom,
        message: 'کد ملی الزامی است',
        path: ['national_code'],
      });
      return;
    }
    if (!isValidIranianNationalCode(code)) {
      ctx.addIssue({
        code: z.ZodIssueCode.custom,
        message: 'کد ملی ایرانی نامعتبر است',
        path: ['national_code'],
      });
    }
  });

export type PatientFormValues = z.infer<typeof patientSchema>;

/** اسکیمای فیلتر جستجوی بیمار */
export const patientSearchSchema = z.object({
  first_name: z.string().optional(),
  last_name: z.string().optional(),
  national_code: z.string().optional(),
  birth_date: z.string().optional(),
  address: z.string().optional(),
  home_phone_number: z.string().optional(),
  mobile_phone_number: z.string().optional(),
  file_number: z.string().optional(),
  sex: z.enum(['', 'true', 'false']).optional(),
});

export type PatientSearchFormValues = z.infer<typeof patientSearchSchema>;
