package apperror

import "testing"

// TestErrAuditUnavailable_StatusAndMessage کد، وضعیت HTTP و پیام فارسی شکست ممیزی را ثابت می‌کند.
func TestErrAuditUnavailable_StatusAndMessage(t *testing.T) {
	err := ErrAuditUnavailable
	if err.Code != "AUDIT_UNAVAILABLE" {
		t.Fatalf("کد: %s", err.Code)
	}
	if err.HTTPStatus != 503 {
		t.Fatalf("وضعیت: %d", err.HTTPStatus)
	}
	if err.UserMsg != "ثبت رویداد امنیتی ممکن نشد؛ عملیات انجام نشد." {
		t.Fatalf("پیام: %s", err.UserMsg)
	}
}
