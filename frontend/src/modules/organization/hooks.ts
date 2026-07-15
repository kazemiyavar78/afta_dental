// این ماژول نمونه است؛ باقی فیلدها/صفحات طبق همین الگو در فازهای بعدی اضافه می‌شوند.

import { z } from 'zod';

/** اسکیمای ایجاد/ویرایش سازمان */
export const organizationSchema = z.object({
  name: z.string().min(1, 'نام سازمان الزامی است'),
  is_takmili: z.boolean(),
  is_active: z.boolean(),
  package_id: z.number().min(1, 'انتخاب بسته تعرفه الزامی است'),
});

export type OrganizationFormValues = z.infer<typeof organizationSchema>;
