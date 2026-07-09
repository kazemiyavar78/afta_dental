// تایپ خطای استاندارد هم‌راستا با apperror بک‌اند

/** خطای استاندارد API */
export type AppError = {
  code: string;
  message: string;
  fields?: Record<string, string>;
};

/** پاسخ خطای بک‌اند */
type ApiErrorResponse = {
  error?: string;
  code?: string;
  fields?: Record<string, string>;
};

/**
 * تبدیل پاسخ خطای Axios به AppError یکدست.
 */
export function parseApiError(data: unknown, fallbackMessage = 'خطای ناشناخته'): AppError {
  if (data && typeof data === 'object') {
    const d = data as ApiErrorResponse;
    return {
      code: d.code ?? 'UNKNOWN',
      message: d.error ?? fallbackMessage,
      fields: d.fields,
    };
  }
  return { code: 'UNKNOWN', message: fallbackMessage };
}
