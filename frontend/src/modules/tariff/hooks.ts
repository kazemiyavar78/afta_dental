// این ماژول نمونه است؛ باقی فیلدها/صفحات طبق همین الگو در فازهای بعدی اضافه می‌شوند.

import { z } from 'zod';

export const tariffSchema = z.object({
  name: z.string().min(1, 'نام تعرفه الزامی است'),
  amount: z.number().min(0, 'مبلغ باید مثبت باشد'),
});

export type TariffFormValues = z.infer<typeof tariffSchema>;
