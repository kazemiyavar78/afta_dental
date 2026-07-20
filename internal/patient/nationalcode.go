package patient

import (
	"strconv"
	"strings"

	"github.com/tpdenta/afta-reception/internal/platform/apperror"
)

// IsValidIranianNationalCode اعتبارسنجی کد ملی ایرانی (۱۰ رقم + رقم کنترل) را انجام می‌دهد.
func IsValidIranianNationalCode(code string) bool {
	code = strings.TrimSpace(code)
	if len(code) != 10 {
		return false
	}
	for _, ch := range code {
		if ch < '0' || ch > '9' {
			return false
		}
	}
	// همه ارقام یکسان نامعتبرند
	allSame := true
	for i := 1; i < 10; i++ {
		if code[i] != code[0] {
			allSame = false
			break
		}
	}
	if allSame {
		return false
	}

	sum := 0
	for i := 0; i < 9; i++ {
		d, _ := strconv.Atoi(string(code[i]))
		sum += d * (10 - i)
	}
	remainder := sum % 11
	check, _ := strconv.Atoi(string(code[9]))
	if remainder < 2 {
		return check == remainder
	}
	return check == 11-remainder
}

// ValidateNationalCode بر اساس اتباع بودن، کد ملی را اعتبارسنجی می‌کند.
// ایرانی: کد ملی ۱۰ رقمی معتبر الزامی است.
// اتباع: کد می‌تواند خالی یا حداکثر ۲۰ کاراکتر باشد (بدون الگوریتم ایرانی).
func ValidateNationalCode(nationalCode string, isForeignNational bool) error {
	code := strings.TrimSpace(nationalCode)
	if isForeignNational {
		if len(code) > 20 {
			return apperror.New("E-002", "کد شناسایی اتباع حداکثر ۲۰ کاراکتر است.", "foreign id too long", 400)
		}
		return nil
	}
	if code == "" {
		return apperror.New("E-002", "کد ملی الزامی است.", "national code required", 400)
	}
	if !IsValidIranianNationalCode(code) {
		return apperror.New("E-002", "کد ملی ایرانی نامعتبر است.", "invalid iranian national code", 400)
	}
	return nil
}
