import { z } from 'zod';

// TODO: خواندن قوانین پیچیدگی رمز از endpoint تنظیمات امنیتی بک‌اند
const MIN_PASSWORD_LENGTH = 8;

/** الگوی شماره تلفن ایرانی */
const iranianPhoneRegex = /^09\d{9}$/;

/** اسکیمای ایجاد کاربر */
export const createUserSchema = z.object({
  username: z.string().min(3, 'نام کاربری حداقل ۳ کاراکتر'),
  password: z
    .string()
    .min(MIN_PASSWORD_LENGTH, `رمز عبور حداقل ${MIN_PASSWORD_LENGTH} کاراکتر`),
  address: z.string(),
  name: z.string().min(1, 'نام الزامی است'),
  family: z.string().min(1, 'نام خانوادگی الزامی است'),
  phone_number: z
    .string()
    .regex(iranianPhoneRegex, 'شماره تلفن باید ۱۱ رقم و با ۰۹ شروع شود')
    .or(z.literal('')),
  medical_code: z.string(),
  role_id: z.number({ error: 'نقش الزامی است' }).min(1, 'نقش الزامی است'),
});

/** اسکیمای ویرایش کاربر */
export const updateUserSchema = createUserSchema
  .omit({ password: true, username: true })
  .extend({
    is_active: z.boolean().optional(),
  });

export type CreateUserFormValues = z.infer<typeof createUserSchema>;
export type UpdateUserFormValues = z.infer<typeof updateUserSchema>;
