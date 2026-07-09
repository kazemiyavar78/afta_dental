// خواندن توکن CSRF از کوکی مرورگر (غیر HttpOnly)

const CSRF_COOKIE_NAME = 'csrf_token';

/**
 * مقدار کوکی csrf_token را از document.cookie می‌خواند.
 */
export function getCsrfTokenFromCookie(): string | null {
  const cookies = document.cookie.split(';');
  for (const raw of cookies) {
    const [name, ...rest] = raw.trim().split('=');
    if (name === CSRF_COOKIE_NAME) {
      return decodeURIComponent(rest.join('='));
    }
  }
  return null;
}
