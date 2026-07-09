// این ماژول نمونه است؛ باقی فیلدها/صفحات طبق همین الگو در فازهای بعدی اضافه می‌شوند.

import { z } from 'zod';

export const organizationSchema = z.object({
  name: z.string().min(1, 'نام سازمان الزامی است'),
});

export type OrganizationFormValues = z.infer<typeof organizationSchema>;
