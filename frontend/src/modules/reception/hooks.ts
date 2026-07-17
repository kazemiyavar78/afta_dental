import { z } from 'zod';

/** اسکیمای حداقلی برای سازگاری با importهای قدیمی */
export const receptionSchema = z.object({
  reception_date: z.string().min(1, 'تاریخ الزامی است'),
});

export type ReceptionFormValues = z.infer<typeof receptionSchema>;
