package audit

import (
	"encoding/json"
	"reflect"
)

// Change مقدار قبل و بعد یک فیلد تغییریافته است.
// Before و After همان نوعی هستند که پس از JSON در جزئیات ممیزی می‌مانند.
type Change struct {
	Before any `json:"before"`
	After  any `json:"after"`
}

// Diff فیلدهای نام‌برده را بین دو مقدار مقایسه می‌کند و فقط فرق‌ها را برمی‌گرداند.
// before و after ساختار یا نقشه هستند. fields نام کلید JSON است، نه نام فیلد Go.
// خروجی nil است اگر fields خالی باشد یا هیچ فیلدی فرق نکرده باشد.
// مقایسه بعد از تبدیل به JSON است تا برچسب json ملاک باشد و نوع دو طرف یکی بماند.
func Diff(before, after any, fields []string) map[string]Change {
	if len(fields) == 0 {
		return nil
	}
	left := objectFields(before)
	right := objectFields(after)
	out := make(map[string]Change)
	for _, field := range fields {
		lv, lok := left[field]
		rv, rok := right[field]
		if !lok && !rok {
			continue
		}
		if lok && rok && reflect.DeepEqual(lv, rv) {
			continue
		}
		change := Change{}
		if lok {
			change.Before = lv
		}
		if rok {
			change.After = rv
		}
		out[field] = change
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// objectFields مقدار را به نقشه کلید JSON تبدیل می‌کند.
// v تهی یا غیرقابل تبدیل، نقشه خالی می‌دهد تا Diff به‌جای panic فیلد را غایب ببیند.
func objectFields(v any) map[string]any {
	if v == nil {
		return map[string]any{}
	}
	raw, err := json.Marshal(v)
	if err != nil {
		return map[string]any{}
	}
	var obj map[string]any
	if err := json.Unmarshal(raw, &obj); err != nil || obj == nil {
		return map[string]any{}
	}
	return obj
}
