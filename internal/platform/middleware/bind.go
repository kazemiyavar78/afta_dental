package middleware

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/tpdenta/afta-reception/internal/platform/apperror"
)

// quotedValue مقدار داخل نقل‌قول را از متن خطای bind حذف می‌کند.
// strconv و Gin مقدار ارسالی را داخل نقل‌قول می‌گذارند و آن مقدار نباید لاگ شود.
var quotedValue = regexp.MustCompile(`"(?:\\.|[^"\\])*"|'(?:\\.|[^'\\])*'`)

// BindJSON بدنه JSON را در dest می‌ریزد.
// c زمینه درخواست است و dest باید اشاره‌گر به struct مقصد باشد.
// اگر ورودی معتبر باشد true برمی‌گرداند و dest پر می‌شود.
// اگر نامعتبر باشد پاسخ ۴۰۰ با VALIDATION_ERROR می‌نویسد و false برمی‌گرداند (D-08).
func BindJSON(c *gin.Context, dest any) bool {
	return bindRequest(c, c.ShouldBindJSON, dest)
}

// BindQuery پارامترهای query را در dest می‌ریزد.
// c زمینه درخواست است و dest باید اشاره‌گر به struct مقصد باشد.
// اگر ورودی معتبر باشد true برمی‌گرداند و dest پر می‌شود.
// اگر نامعتبر باشد پاسخ ۴۰۰ با VALIDATION_ERROR می‌نویسد و false برمی‌گرداند (D-08).
func BindQuery(c *gin.Context, dest any) bool {
	return bindRequest(c, c.ShouldBindQuery, dest)
}

// bindRequest خطای Gin را به خطای اعتبارسنجی یکدست تبدیل می‌کند.
// bind همان ShouldBindJSON یا ShouldBindQuery است و dest مقصد پر شدن است.
// true یعنی خطا نبود. false یعنی پاسخ خطا نوشته شده و handler باید برگردد.
func bindRequest(c *gin.Context, bind func(any) error, dest any) bool {
	if err := bind(dest); err != nil {
		log.Printf("bind input failed: %s", bindLogMessage(err))
		WriteError(c, apperror.ErrValidation)
		return false
	}
	return true
}

// bindLogMessage متن لاگ سرور را از خطای bind می‌سازد.
// err خطای Gin یا validator است. خروجی نام فیلد و نوع قاعده است، بدون مقدار ارسالی.
func bindLogMessage(err error) string {
	var slices binding.SliceValidationError
	if errors.As(err, &slices) {
		parts := make([]string, 0, len(slices))
		for _, item := range slices {
			if item == nil {
				continue
			}
			parts = append(parts, bindLogMessage(item))
		}
		return strings.Join(parts, "; ")
	}

	var fields validator.ValidationErrors
	if errors.As(err, &fields) {
		parts := make([]string, 0, len(fields))
		for _, field := range fields {
			parts = append(parts, field.Field()+" tag="+field.Tag())
		}
		return strings.Join(parts, "; ")
	}

	var typeErr *json.UnmarshalTypeError
	if errors.As(err, &typeErr) {
		name := typeErr.Field
		if name == "" {
			name = "(root)"
		}
		// Value اینجا نوع JSON است (string، number) نه محتوای ارسالی.
		return "field=" + name + " json_type=" + typeErr.Value
	}

	if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
		return "empty or truncated body"
	}

	return redactBindText(err.Error())
}

// redactBindText نقل‌قول‌ها را از متن خطا حذف و طول را محدود می‌کند.
// text متن خام خطا است. خروجی برای لاگ سرور است و مقدار فیلد را ندارد.
func redactBindText(text string) string {
	cleaned := quotedValue.ReplaceAllString(text, "[redacted]")
	cleaned = strings.Join(strings.Fields(cleaned), " ")
	const maxLen = 180
	if len(cleaned) > maxLen {
		return cleaned[:maxLen]
	}
	return cleaned
}
