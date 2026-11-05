package middleware_test

import (
	"net/http"
	"testing"

	"github.com/tpdenta/afta-reception/internal/platform/apperror"
	"github.com/tpdenta/afta-reception/internal/platform/middleware"
	"github.com/tpdenta/afta-reception/internal/platform/security/audit"
	"github.com/tpdenta/afta-reception/internal/platform/security/loginguard"
	"github.com/tpdenta/afta-reception/internal/platform/testutil"
)

type lockGuard struct {
	blocks *loginguard.BlockedLoginLog
}

func (g lockGuard) CheckIPLock(string) error { return apperror.ErrIPLocked }

func (g lockGuard) AllowBlockedLoginLog(ip string) bool { return g.blocks.Allow(ip) }

// TestR1_18_MiddlewareLoginBlockedOncePerMinute میان‌افزار قفل IP هم حداکثر یک LOGIN_BLOCKED در دقیقه می‌نویسد.
func TestR1_18_MiddlewareLoginBlockedOncePerMinute(t *testing.T) {
	rec := &testutil.FakeAuditRecorder{}
	guard := lockGuard{blocks: loginguard.NewBlockedLoginLog()}
	handler := middleware.LoginRateLimiterMiddleware(guard, rec)
	for i := 0; i < 2; i++ {
		g := testutil.NewTestGinContext(http.MethodPost, "/api/login", nil)
		g.Context.Request.RemoteAddr = "203.0.113.50:1234"
		handler(g.Context)
		if g.Recorder.Code != http.StatusForbidden {
			t.Fatalf("وضعیت باید ۴۰۳ باشد: %d", g.Recorder.Code)
		}
	}
	if len(rec.Events) != 1 {
		t.Fatalf("باید یک LOGIN_BLOCKED باشد، %d ثبت شد", len(rec.Events))
	}
	got := rec.Events[0].Event
	if got.Type != audit.EventLoginBlocked || got.Reason != audit.ReasonIPLocked || got.Actor.IP != "203.0.113.50" {
		t.Fatalf("رویداد میان‌افزار نادرست است: %+v", got)
	}
}
