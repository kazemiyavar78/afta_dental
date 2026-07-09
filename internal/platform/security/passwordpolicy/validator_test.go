package passwordpolicy

import (
	"testing"

	"github.com/tpdenta/afta-reception/internal/platform/security/settings"
)

// TestHasUpper بررسی وجود حرف بزرگ را تست می‌کند.
func TestHasUpper(t *testing.T) {
	if !hasUpper("Abc") {
		t.Fatal("باید حرف بزرگ تشخیص داده شود")
	}
	if hasUpper("abc") {
		t.Fatal("نباید حرف بزرگ تشخیص داده شود")
	}
}

// TestDefaultSettings مقادیر پیش‌فرض تنظیمات رمز را بررسی می‌کند.
func TestDefaultSettings(t *testing.T) {
	if settings.DefaultSettings[settings.PasswordMinLength] != "8" {
		t.Fatal("حداقل طول رمز باید ۸ باشد")
	}
}
