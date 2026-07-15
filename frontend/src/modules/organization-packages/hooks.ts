import { z } from 'zod';

/** اسکیمای ایجاد/ویرایش بسته تعرفه سازمان */
export const organizationPackageSchema = z.object({
  package_name: z.string().min(1, 'نام بسته الزامی است').max(200, 'حداکثر ۲۰۰ کاراکتر'),
  package_description: z.string().max(200, 'حداکثر ۲۰۰ کاراکتر'),
  technical_coefficient: z.number().int().min(0, 'نباید منفی باشد'),
  technical_professional_coefficient: z.number().int().min(0, 'نباید منفی باشد'),
  consumption_coefficient: z.number().int().min(0, 'نباید منفی باشد'),
  subsidy_percentage: z.number().int().min(0, 'نباید منفی باشد').max(100, 'حداکثر ۱۰۰'),
  supplementary_percentage: z.number().int().min(0, 'نباید منفی باشد').max(100, 'حداکثر ۱۰۰'),
  organization_percentage: z.number().int().min(0, 'نباید منفی باشد').max(100, 'حداکثر ۱۰۰'),
});

export type OrganizationPackageFormValues = z.infer<typeof organizationPackageSchema>;
