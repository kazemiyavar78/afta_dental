import { z } from 'zod';

/** اسکیمای فرم تست/ذخیره تعرفه (سازمان + سه مبلغ مرکز) */
export const tariffCalculateSchema = z.object({
  organization_id: z.number().min(1, 'انتخاب سازمان الزامی است'),
  technical_amount: z.number().min(0, 'نباید منفی باشد'),
  professional_center_amount: z.number().min(0, 'نباید منفی باشد'),
  consumption_center_amount: z.number().min(0, 'نباید منفی باشد'),
});

export type TariffCalculateFormValues = z.infer<typeof tariffCalculateSchema>;

/** اسکیمای فرم بازمحاسبه یک تعرفه */
export const tariffRecalculateSchema = z.object({
  technical_amount: z.number().min(0, 'نباید منفی باشد'),
  professional_center_amount: z.number().min(0, 'نباید منفی باشد'),
  consumption_center_amount: z.number().min(0, 'نباید منفی باشد'),
});

export type TariffRecalculateFormValues = z.infer<typeof tariffRecalculateSchema>;
