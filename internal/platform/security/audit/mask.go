package audit

import (
	"encoding/json"
	"strings"
)

// MaskTail همه نویسه‌ها جز keep نویسه آخر را ستاره می‌کند.
// value متن خام است و keep تعداد نویسه آشکار انتهاست. اگر متن کوتاه‌تر باشد همان برمی‌گردد.
// برای کد ملی و تلفن در جزئیات ممیزی، keep برابر ۴ است.
func MaskTail(value string, keep int) string {
	if keep < 0 {
		keep = 0
	}
	runes := []rune(value)
	if len(runes) <= keep {
		return value
	}
	return strings.Repeat("*", len(runes)-keep) + string(runes[len(runes)-keep:])
}

// maskedValue جایگزینی ثابت مقدار حساس است و هرگز خود مقدار را نگه نمی‌دارد.
const maskedValue = "***"

// sensitiveParts بخش‌هایی از نام کلید هستند که مقدارشان نباید در ممیزی بماند.
// pass جداگانه و فقط با برابری دقیق سنجیده می‌شود تا کلیدهایی مثل compass ماسک نشوند.
var sensitiveParts = []string{
	"password",
	"token",
	"otp",
	"csrf",
	"secret",
	"session_id",
	"security_code",
}

// MaskSensitive مقدار را برای ذخیره در جزئیات ممیزی بی‌خطر می‌کند.
// v هر ساختار قابل تبدیل به JSON است. خروجی همان شکل است و مقدار کلیدهای حساس *** شده.
// نقشه و آرایه تو در تو هم پیمایش می‌شوند. نام فیلد ساخت‌ها از برچسب json خوانده می‌شود.
// اگر v به JSON تبدیل نشود همان ورودی برمی‌گردد تا ثبت‌کننده خطای marshal را ببیند.
func MaskSensitive(v any) any {
	if v == nil {
		return nil
	}
	raw, err := json.Marshal(v)
	if err != nil {
		return v
	}
	var decoded any
	if err := json.Unmarshal(raw, &decoded); err != nil {
		return v
	}
	return maskJSON(decoded)
}

// maskJSON یک مقدار JSON عمومی را پیمایش می‌کند.
// v شیء، آرایه یا مقدار ساده پس از Unmarshal است. خروجی همان درخت با مقدارهای ماسک‌شده است.
func maskJSON(v any) any {
	switch val := v.(type) {
	case map[string]any:
		out := make(map[string]any, len(val))
		for key, child := range val {
			if isSensitiveKey(key) {
				out[key] = maskedValue
				continue
			}
			out[key] = maskJSON(child)
		}
		return out
	case []any:
		out := make([]any, len(val))
		for i, child := range val {
			out[i] = maskJSON(child)
		}
		return out
	default:
		return v
	}
}

// isSensitiveKey می‌گوید نام کلید نباید در ممیزی آشکار بماند.
// key نام فیلد JSON است. مقایسه بدون حساسیت به حروف است.
// pass فقط برابر دقیق است. بقیه اگر همان نام باشند یا داخل نام کلید آمده باشند حساس‌اند.
func isSensitiveKey(key string) bool {
	name := strings.ToLower(strings.TrimSpace(key))
	if name == "pass" {
		return true
	}
	for _, part := range sensitiveParts {
		if strings.Contains(name, part) {
			return true
		}
	}
	return false
}
