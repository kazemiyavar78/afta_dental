// این ماژول نمونه است؛ باقی فیلدها/صفحات طبق همین الگو در فازهای بعدی اضافه می‌شوند.

import { z } from 'zod';

export const fundSchema = z.object({
  name: z.string().min(1, 'نام صندوق الزامی است'),
});

export type FundFormValues = z.infer<typeof fundSchema>;
