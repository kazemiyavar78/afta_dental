package apperror

import "fmt"

// AppError خطای یکدست برنامه با کد، پیام کاربرپسند فارسی و پیام لاگ داخلی.
type AppError struct {
	Code       string
	UserMsg    string
	LogMsg     string
	HTTPStatus int
}

// Error پیام لاگ داخلی را برمی‌گرداند.
func (e *AppError) Error() string {
	return fmt.Sprintf("[%s] %s", e.Code, e.LogMsg)
}

// New خطای جدید می‌سازد.
func New(code, userMsg, logMsg string, httpStatus int) *AppError {
	return &AppError{
		Code:       code,
		UserMsg:    userMsg,
		LogMsg:     logMsg,
		HTTPStatus: httpStatus,
	}
}

// خطاهای رایج
var (
	ErrNotFound      = New("NOT_FOUND", "مورد درخواستی یافت نشد.", "resource not found", 404)
	ErrUnauthorized  = New("UNAUTHORIZED", "احراز هویت انجام نشده است.", "unauthorized", 401)
	ErrForbidden     = New("FORBIDDEN", "دسترسی مجاز نیست.", "forbidden", 403)
	ErrBadRequest    = New("BAD_REQUEST", "درخواست نامعتبر است.", "bad request", 400)
	ErrInternal      = New("INTERNAL", "خطای داخلی سرور رخ داده است.", "internal server error", 500)
	ErrConflict      = New("CONFLICT", "تعارض در داده‌ها.", "conflict", 409)
	ErrInvalidCreds  = New("INVALID_CREDENTIALS", "نام کاربری یا رمز عبور اشتباه است.", "invalid credentials", 401)
	ErrAccountLocked = New("ACCOUNT_LOCKED", "حساب کاربری قفل شده است.", "account locked", 403)
	ErrIPLocked      = New("IP_LOCKED", "آدرس IP شما به دلیل تلاش‌های ناموفق قفل شده است.", "ip locked", 403)
	ErrWorkHours     = New("WORK_HOURS", "ورود خارج از ساعات کاری مجاز نیست.", "outside work hours", 403)
	ErrCSRF          = New("CSRF", "توکن CSRF نامعتبر است.", "csrf validation failed", 403)
	ErrIntegrity     = New("INTEGRITY", "یکپارچگی داده‌ها نقض شده است.", "integrity check failed", 500)
)
