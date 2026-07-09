import { z } from 'zod';

/** اعتبارسنجی فرم لاگین */
export const loginSchema = z.object({
  username: z.string().min(1, 'نام کاربری الزامی است'),
  password: z.string().min(1, 'رمز عبور الزامی است'),
});

export type LoginFormValues = z.infer<typeof loginSchema>;

export type LoginResponse = {
  user: {
    id: number;
    username: string;
    name: string;
    family: string;
    role_name: string;
  };
  session_id: string;
};
