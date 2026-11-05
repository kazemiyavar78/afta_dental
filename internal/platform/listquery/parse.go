package listquery

import (
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/tpdenta/afta-reception/internal/platform/apperror"
)

// filterKey نام داخل filter[name] است. فقط شناسه ساده پذیرفته می‌شود تا کلید جعلی به SQL نرسد.
var filterKey = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

// Parse پارامترهای query را با spec اعتبارسنجی می‌کند.
// values همان Query درخواست است و spec تعریف ستون‌های همان API است.
// خروجی Query آماده Apply است. خطا یعنی VALIDATION_ERROR و هیچ ستونی نادیده گرفته نشده است.
func Parse(values url.Values, spec Spec) (Query, error) {
	if values == nil {
		values = url.Values{}
	}
	page, err := parsePage(values.Get("page"))
	if err != nil {
		return Query{}, err
	}
	size, err := parsePageSize(values.Get("page_size"))
	if err != nil {
		return Query{}, err
	}
	expr, desc, err := parseSort(values, spec)
	if err != nil {
		return Query{}, err
	}
	filters, applied, err := parseFilters(values, spec)
	if err != nil {
		return Query{}, err
	}
	return Query{
		Page:     page,
		PageSize: size,
		sortExpr: expr,
		sortDesc: desc,
		filters:  filters,
		applied:  applied,
	}, nil
}

// parsePage شماره صفحه را می‌خواند.
// raw مقدار page است. خالی یعنی صفحه ۱. خروجی شماره حداقل ۱ یا خطای اعتبارسنجی است.
func parsePage(raw string) (int, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return DefaultPage, nil
	}
	page, err := strconv.Atoi(raw)
	if err != nil || page < 1 {
		return 0, apperror.ErrValidation
	}
	return page, nil
}

// parsePageSize اندازه صفحه را می‌خواند.
// raw مقدار page_size است. خالی یعنی ۲۰. بزرگ‌تر از ۱۰۰ به ۱۰۰ بریده می‌شود.
// مقدار غیرعددی یا کمتر از ۱ خطای اعتبارسنجی است.
func parsePageSize(raw string) (int, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return DefaultPageSize, nil
	}
	size, err := strconv.Atoi(raw)
	if err != nil || size < 1 {
		return 0, apperror.ErrValidation
	}
	if size > MaxPageSize {
		return MaxPageSize, nil
	}
	return size, nil
}

// parseSort ستون و جهت مرتب‌سازی را از spec برمی‌دارد.
// values پارامترهای query و spec تعریف ستون‌هاست.
// خروجی عبارت SQL سرور و جهت نزولی است. ستون ناشناس یا غیرقابل مرتب‌سازی خطا است.
func parseSort(values url.Values, spec Spec) (string, bool, error) {
	name := strings.TrimSpace(values.Get("sort"))
	order := strings.ToLower(strings.TrimSpace(values.Get("order")))
	if order != "" && order != "asc" && order != "desc" {
		return "", false, apperror.ErrValidation
	}
	if name == "" {
		name = spec.DefaultSort
	}
	desc := spec.DefaultDesc
	if order != "" {
		desc = order == "desc"
	}
	if name == "" {
		return "", desc, nil
	}
	col, ok := spec.Columns[name]
	if !ok || !col.Sortable || col.Expr == "" {
		return "", false, apperror.ErrValidation
	}
	return col.Expr, desc, nil
}

// parseFilters فیلترهای filter[...] را به ترتیب پایدار می‌سازد.
// values پارامترها و spec ستون‌های مجاز است.
// خروجی شرط‌ها و نقشه ممیزی است. نام ناشناس خطا است و بی‌صدا حذف نمی‌شود.
func parseFilters(values url.Values, spec Spec) ([]filterClause, map[string]string, error) {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	clauses := make([]filterClause, 0)
	applied := make(map[string]string)
	for _, key := range keys {
		name, isFilter, bad := classifyFilterKey(key)
		if !isFilter {
			continue
		}
		if bad || len(values[key]) != 1 {
			return nil, nil, apperror.ErrValidation
		}
		clause, label, value, err := buildFilter(name, values[key][0], spec)
		if err != nil {
			return nil, nil, err
		}
		if clause == nil {
			continue
		}
		clauses = append(clauses, *clause)
		applied[label] = value
	}
	return clauses, applied, nil
}

// classifyFilterKey می‌گوید کلید query فیلتر است یا نه.
// key نام پارامتر است. isFilter برای پیشوند filter[ است.
// bad یعنی شکل کلید غلط است و باید رد شود، نه اینکه نادیده گرفته شود.
func classifyFilterKey(key string) (name string, isFilter bool, bad bool) {
	if !strings.HasPrefix(key, "filter[") {
		return "", false, false
	}
	if !strings.HasSuffix(key, "]") {
		return "", true, true
	}
	name = key[len("filter[") : len(key)-1]
	if !filterKey.MatchString(name) {
		return "", true, true
	}
	return name, true, false
}
