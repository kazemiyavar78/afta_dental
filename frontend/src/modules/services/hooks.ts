import { z } from 'zod';

/** اسکیمای ایجاد/ویرایش خدمت */
export const serviceSchema = z.object({
  service_code: z.string().min(1, 'کد خدمت الزامی است').max(20, 'حداکثر ۲۰ کاراکتر'),
  name: z.string().min(1, 'نام خدمت الزامی است').max(200, 'حداکثر ۲۰۰ کاراکتر'),
  technical_coefficient: z.number().min(0, 'نباید منفی باشد'),
  professional_coefficient: z.number().min(0, 'نباید منفی باشد'),
  consumption_coefficient: z.number().min(0, 'نباید منفی باشد'),
  service_rate: z.number().min(0, 'نباید منفی باشد'),
  service_tariff: z.number().min(0, 'نباید منفی باشد'),
  international_code: z.string().max(20, 'حداکثر ۲۰ کاراکتر'),
  default_count: z.number().int('باید عدد صحیح باشد').min(0, 'نباید منفی باشد'),
  maximum_count: z.number().int('باید عدد صحیح باشد').min(0, 'نباید منفی باشد'),
  service_features: z.enum(['', '#', '*', '#*']),
  is_active: z.boolean(),
  is_dental_direction: z.boolean(),
  allow_multiple_use: z.boolean(),
});


export type ServiceFormValues = z.infer<typeof serviceSchema>;
