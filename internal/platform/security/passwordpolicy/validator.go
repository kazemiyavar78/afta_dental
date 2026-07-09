package passwordpolicy

import (
	"strings"
	"unicode"

	"github.com/tpdenta/afta-reception/internal/platform/security/settings"
)

// Validator اعتبارسنج پیچیدگی رمز عبور.
type Validator struct {
	settings *settings.Service
}

// NewValidator نمونه Validator می‌سازد.
func NewValidator(settingsSvc *settings.Service) *Validator {
	return &Validator{settings: settingsSvc}
}

// IsValidPassword رمز را بر اساس تنظیمات امنیتی بررسی می‌کند.
func (v *Validator) IsValidPassword(password string) (bool, []string) {
	var errors []string

	minLen, _ := v.settings.GetSettingInt(settings.PasswordMinLength)
	if minLen <= 0 {
		minLen = 8
	}
	if len(password) < minLen {
		errors = append(errors, "رمز عبور باید حداقل "+itoa(minLen)+" کاراکتر باشد.")
	}

	reqUpper, _ := v.settings.GetSettingValue(settings.PasswordRequireUppercase)
	if reqUpper == "true" && !hasUpper(password) {
		errors = append(errors, "رمز عبور باید حداقل یک حرف بزرگ انگلیسی داشته باشد.")
	}

	reqLower, _ := v.settings.GetSettingValue(settings.PasswordRequireLowercase)
	if reqLower == "true" && !hasLower(password) {
		errors = append(errors, "رمز عبور باید حداقل یک حرف کوچک انگلیسی داشته باشد.")
	}

	reqDigit, _ := v.settings.GetSettingValue(settings.PasswordRequireDigit)
	if reqDigit == "true" && !hasDigit(password) {
		errors = append(errors, "رمز عبور باید حداقل یک رقم داشته باشد.")
	}

	reqSpecial, _ := v.settings.GetSettingValue(settings.PasswordRequireSpecial)
	if reqSpecial == "true" && !hasSpecial(password) {
		errors = append(errors, "رمز عبور باید حداقل یک کاراکتر ویژه داشته باشد.")
	}

	return len(errors) == 0, errors
}

func hasUpper(s string) bool {
	for _, r := range s {
		if unicode.IsUpper(r) {
			return true
		}
	}
	return false
}

func hasLower(s string) bool {
	for _, r := range s {
		if unicode.IsLower(r) {
			return true
		}
	}
	return false
}

func hasDigit(s string) bool {
	for _, r := range s {
		if unicode.IsDigit(r) {
			return true
		}
	}
	return false
}

func hasSpecial(s string) bool {
	special := "!@#$%^&*()_+-=[]{}|;':\",./<>?"
	for _, r := range s {
		if strings.ContainsRune(special, r) {
			return true
		}
	}
	return false
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var digits []byte
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	return string(digits)
}
