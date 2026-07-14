import axios, { type AxiosError, type InternalAxiosRequestConfig } from 'axios';
import { message } from 'antd';
import { API_BASE_URL } from '../config/env';
import { getCsrfTokenFromCookie } from './csrf';
import { parseApiError } from './errorTypes';

const MUTATING_METHODS = new Set(['post', 'put', 'patch', 'delete']);

/** هندلر خروج اجباری هنگام ۴۰۱ */
let unauthorizedHandler: (() => void) | null = null;

/**
 * ثبت تابع هدایت به لاگین — از AuthProvider فراخوانی می‌شود.
 */
export function setUnauthorizedHandler(handler: () => void): void {
  unauthorizedHandler = handler;
}

/** نمونه Axios با interceptor های امنیتی */
export const httpClient = axios.create({
  baseURL: API_BASE_URL.startsWith('http')
    ? API_BASE_URL
    : `${API_BASE_URL}`,
  withCredentials: true,
  headers: {
    'Content-Type': 'application/json',
  },
});

httpClient.interceptors.request.use((config: InternalAxiosRequestConfig) => {
  const method = (config.method ?? 'get').toLowerCase();
  if (MUTATING_METHODS.has(method)) {
    const token = getCsrfTokenFromCookie();
    if (token) {
      config.headers.set('X-CSRF-Token', token);
    }
  }
  return config;
});

httpClient.interceptors.response.use(
  (response) => response,
  (error: AxiosError) => {
    const status = error.response?.status;
    const currentPath = window.location.pathname;

    if (status === 401) {
      if (currentPath !== '/login') {
        unauthorizedHandler?.();
      }
      return Promise.reject(parseApiError(error.response?.data, 'احراز هویت انجام نشده است.'));
    }

    if (status === 403) {
      const appError = parseApiError(error.response?.data, 'دسترسی مجاز نیست.');
      message.error(appError.message);
      return Promise.reject(appError);
    }

    const appError = parseApiError(
      error.response?.data,
      error.message || 'خطا در ارتباط با سرور',
    );
    return Promise.reject(appError);
  },
);
