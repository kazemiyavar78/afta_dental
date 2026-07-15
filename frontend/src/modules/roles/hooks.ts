import { z } from 'zod';

/** اسکیمای ایجاد/ویرایش نقش */
export const roleSchema = z.object({
  name: z.string().min(1, 'نام نقش الزامی است'),
  description: z.string(),
  permission_ids: z.array(z.number()),
});

export type RoleFormValues = z.infer<typeof roleSchema>;
