package listquery

import (
	"strconv"
	"strings"
	"time"

	"github.com/tpdenta/afta-reception/internal/platform/apperror"
)

// buildFilter یک کلید فیلتر را به شرط پارامتری تبدیل می‌کند.
// name نام داخل کروشه، raw مقدار ارسالی و spec تعریف ستون‌هاست.
// clause تهی و خطای تهی یعنی مقدار خالی و شرطی لازم نیست.
// خطا یعنی ستون یا مقدار پذیرفته نیست. label و value برای ممیزی‌اند.
func buildFilter(name, raw string, spec Spec) (*filterClause, string, string, error) {
	if col, ok := spec.Columns[name]; ok {
		return buildColumn(name, raw, col)
	}
	base, edge, ok := splitRange(name)
	if !ok {
		return nil, "", "", apperror.ErrValidation
	}
	col, exists := spec.Columns[base]
	if !exists || !col.Filterable || col.Expr == "" {
		return nil, "", "", apperror.ErrValidation
	}
	if col.Type != KindNumber && col.Type != KindDate {
		return nil, "", "", apperror.ErrValidation
	}
	return buildBound(name, edge, raw, col)
}

// buildColumn فیلتر مستقیم یک ستون غیر بازه‌ای را می‌سازد.
// name نام ستون، raw مقدار و col تعریف سرور است.
// عدد و تاریخ اینجا رد می‌شوند چون فقط با _from و _to فیلتر می‌شوند.
func buildColumn(name, raw string, col Column) (*filterClause, string, string, error) {
	if !col.Filterable || col.Expr == "" {
		return nil, "", "", apperror.ErrValidation
	}
	if col.Type == KindNumber || col.Type == KindDate {
		return nil, "", "", apperror.ErrValidation
	}
	value := columnValue(raw, col)
	if value == "" {
		return nil, "", "", nil
	}
	switch col.Type {
	case KindText:
		return &filterClause{
			sql:  col.Expr + " LIKE ? ESCAPE '\\'",
			args: []any{"%" + escapeLike(value) + "%"},
		}, name, value, nil
	case KindExact:
		return &filterClause{sql: col.Expr + " = ?", args: []any{value}}, name, value, nil
	case KindEnum:
		return buildEnum(name, value, col)
	case KindBool:
		return buildBool(name, value, col)
	default:
		return nil, "", "", apperror.ErrValidation
	}
}

// columnValue مقدار فیلتر را قبل از مقایسه نرمال می‌کند.
// متن فقط وقتی Trim روشن باشد فاصله‌های کنار را از دست می‌دهد. بقیه نوع‌ها همیشه trim می‌شوند.
func columnValue(raw string, col Column) string {
	if col.Type == KindText && !col.Trim {
		return raw
	}
	return strings.TrimSpace(raw)
}

// buildEnum شرط IN را از مقدارهای جداشده با ویرگول می‌سازد.
// name نام ستون، value متن trim‌شده و col عبارت SQL سرور است.
// جزء خالی بین ویرگول‌ها خطا است. هر جزء یک پارامتر جداست.
func buildEnum(name, value string, col Column) (*filterClause, string, string, error) {
	parts := strings.Split(value, ",")
	args := make([]any, 0, len(parts))
	clean := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			return nil, "", "", apperror.ErrValidation
		}
		args = append(args, part)
		clean = append(clean, part)
	}
	marks := strings.TrimRight(strings.Repeat("?,", len(args)), ",")
	return &filterClause{sql: col.Expr + " IN (" + marks + ")", args: args}, name, strings.Join(clean, ","), nil
}

// buildBool برابری true یا false را می‌سازد.
// name نام ستون و value متن trim‌شده است. col عبارت SQL است.
// فقط همین دو واژه، بدون توجه به حروف بزرگ، پذیرفته می‌شود.
func buildBool(name, value string, col Column) (*filterClause, string, string, error) {
	var flag bool
	switch strings.ToLower(value) {
	case "true":
		flag = true
	case "false":
		flag = false
	default:
		return nil, "", "", apperror.ErrValidation
	}
	return &filterClause{sql: col.Expr + " = ?", args: []any{flag}}, name, strconv.FormatBool(flag), nil
}

// splitRange پسوند _from یا _to را از نام فیلتر جدا می‌کند.
// name نام داخل کروشه است. ok نادرست یعنی این نام بازه نیست.
func splitRange(name string) (base, edge string, ok bool) {
	switch {
	case strings.HasSuffix(name, "_from"):
		return strings.TrimSuffix(name, "_from"), "from", true
	case strings.HasSuffix(name, "_to"):
		return strings.TrimSuffix(name, "_to"), "to", true
	default:
		return "", "", false
	}
}

// buildBound یک سر بازه عدد یا تاریخ را می‌سازد.
// key نام کامل فیلتر، edge مقدار from یا to، raw متن کاربر و col تعریف ستون است.
// تاریخِ to تا شروع روز بعد، یعنی پایان همان روز، مقایسه می‌شود.
func buildBound(key, edge, raw string, col Column) (*filterClause, string, string, error) {
	if col.Type == KindDate {
		return buildDateBound(key, edge, raw, col)
	}
	return buildNumberBound(key, edge, raw, col)
}

// buildDateBound یک سر بازه تاریخ میلادی را می‌سازد.
// raw فقط YYYY-MM-DD است. from از نیمه شب UTC همان روز است و to پیش از نیمه شب روز بعد.
func buildDateBound(key, edge, raw string, col Column) (*filterClause, string, string, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return nil, "", "", nil
	}
	day, err := time.ParseInLocation("2006-01-02", value, time.UTC)
	if err != nil {
		return nil, "", "", apperror.ErrValidation
	}
	if edge == "from" {
		return &filterClause{sql: col.Expr + " >= ?", args: []any{day}}, key, value, nil
	}
	return &filterClause{sql: col.Expr + " < ?", args: []any{day.AddDate(0, 0, 1)}}, key, value, nil
}

// buildNumberBound یک سر بازه عدد صحیح را می‌سازد.
// raw باید عدد صحیح باشد. from با >= و to با <= مقایسه می‌شود تا هر دو سر شامل خود مقدار باشند.
func buildNumberBound(key, edge, raw string, col Column) (*filterClause, string, string, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return nil, "", "", nil
	}
	number, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return nil, "", "", apperror.ErrValidation
	}
	op := ">="
	if edge == "to" {
		op = "<="
	}
	return &filterClause{sql: col.Expr + " " + op + " ?", args: []any{number}}, key, value, nil
}

// escapeLike نویسه‌های خاص LIKE را بی‌اثر می‌کند.
// % و _ و خود نویسه فرار با پیشوند \ می‌آیند چون کوئری ESCAPE '\' دارد.
// [ در SQL Server شروع کلاس نویسه است و به شکل [[] لفظی می‌شود.
func escapeLike(value string) string {
	var b strings.Builder
	b.Grow(len(value))
	for _, r := range value {
		switch r {
		case '\\', '%', '_':
			b.WriteByte('\\')
			b.WriteRune(r)
		case '[':
			b.WriteString("[[]")
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}
