package middleware_test

import (
	"net/http"
	"testing"

	"github.com/tpdenta/afta-reception/internal/platform/middleware"
	"github.com/tpdenta/afta-reception/internal/platform/security/audit"
	"github.com/tpdenta/afta-reception/internal/platform/testutil"
)

// TestFAU_GEN1_AuditReadDenied رد دسترسی به /api/logs را ثبت می‌کند و مسیر دیگر را نه.
func TestFAU_GEN1_AuditReadDenied(t *testing.T) {
	rec := &testutil.FakeAuditRecorder{}
	middleware.SetAuthorizationAudit(rec)
	t.Cleanup(func() { middleware.SetAuthorizationAudit(nil) })

	logs := testutil.NewTestGinContext(http.MethodGet, "/api/logs", nil)
	logs.SetUserID(4)
	logs.SetUsername("sara")
	logs.Context.Set("permissionRequirement", middleware.PermissionRequirement{Permissions: []string{"logs.read"}})
	logs.SetPermissions([]string{"patients.read"})
	middleware.AuthorizationMiddleware()(logs.Context)
	if logs.Recorder.Code != http.StatusForbidden {
		t.Fatalf("وضعیت لاگ: %d", logs.Recorder.Code)
	}
	if len(rec.Events) != 1 || rec.Events[0].Event.Type != audit.EventAuditReadDenied || rec.Events[0].Event.Actor.Username != "sara" {
		t.Fatalf("رد خواندن ثبت نشد: %+v", rec.Events)
	}

	other := testutil.NewTestGinContext(http.MethodGet, "/api/patients", nil)
	other.SetUserID(4)
	other.SetUsername("sara")
	other.Context.Set("permissionRequirement", middleware.PermissionRequirement{Permissions: []string{"patients.read"}})
	other.SetPermissions(nil)
	middleware.AuthorizationMiddleware()(other.Context)
	if len(rec.Events) != 1 {
		t.Fatalf("رد مسیر دیگر نباید AUDIT_READ_DENIED بسازد: %d", len(rec.Events))
	}
}
