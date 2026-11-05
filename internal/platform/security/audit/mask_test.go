package audit

import "testing"

// TestMaskSensitive_ReplacesSecretKeys کلیدهای حساس تو در تو را با *** عوض می‌کند و فیلد عادی را نگه می‌دارد.
func TestMaskSensitive_ReplacesSecretKeys(t *testing.T) {
	in := map[string]any{
		"password":      "p1",
		"PASSWORD":      "p2",
		"password_hash": "p3",
		"pass":          "p4",
		"token":         "t",
		"access_token":  "at",
		"otp":           "123456",
		"csrf":          "c",
		"secret":        "s",
		"session_id":    "sid",
		"security_code": "sc",
		"username":      "ali",
		"compass":       "keep",
		"nested": map[string]any{
			"otp":  "999",
			"name": "آزاده",
		},
		"items": []any{
			map[string]any{"csrf": "header", "ok": true},
		},
	}

	got, ok := MaskSensitive(in).(map[string]any)
	if !ok {
		t.Fatalf("خروجی باید شیء باشد: %#v", MaskSensitive(in))
	}
	for _, key := range []string{"password", "PASSWORD", "password_hash", "pass", "token", "access_token", "otp", "csrf", "secret", "session_id", "security_code"} {
		if got[key] != "***" {
			t.Fatalf("کلید %s ماسک نشد: %#v", key, got[key])
		}
	}
	if got["username"] != "ali" || got["compass"] != "keep" {
		t.Fatalf("فیلد عادی نباید ماسک شود: %#v", got)
	}
	nested, ok := got["nested"].(map[string]any)
	if !ok || nested["otp"] != "***" || nested["name"] != "آزاده" {
		t.Fatalf("شیء تو در تو: %#v", got["nested"])
	}
	items, ok := got["items"].([]any)
	if !ok || len(items) != 1 {
		t.Fatalf("آرایه: %#v", got["items"])
	}
	item, ok := items[0].(map[string]any)
	if !ok || item["csrf"] != "***" || item["ok"] != true {
		t.Fatalf("عضو آرایه: %#v", items[0])
	}
}

// TestMaskSensitive_NilStaysNil ورودی تهی همان تهی می‌ماند تا جزئیات خالی ذخیره نشود.
func TestMaskSensitive_NilStaysNil(t *testing.T) {
	if MaskSensitive(nil) != nil {
		t.Fatal("nil باید nil بماند")
	}
}
