package logs

import (
	"net/http"
	"testing"

	"github.com/tpdenta/afta-reception/internal/platform/security/audit"
	"github.com/tpdenta/afta-reception/internal/platform/testutil"
)

type fakeFinder struct{}

func (fakeFinder) FindEvents(audit.EventFilter) ([]audit.SecurityEvent, error) {
	return []audit.SecurityEvent{{ID: 1, EventType: audit.EventLoginFailed}}, nil
}

// TestFAU_GEN1_AuditReadOnlyFirstPage صفحه اول جستجو را ثبت می‌کند و صفحه بعد را نه.
func TestFAU_GEN1_AuditReadOnlyFirstPage(t *testing.T) {
	rec := &testutil.FakeAuditRecorder{}
	h := NewHandler(fakeFinder{}, rec)

	second := testutil.NewTestGinContext(http.MethodGet, "/api/logs?ip=203.0.113.9&page=2", nil)
	second.SetUserID(3)
	second.SetUsername("sara")
	h.List(second.Context)
	if second.Recorder.Code != http.StatusOK {
		t.Fatalf("فهرست: %d %s", second.Recorder.Code, second.Recorder.Body.String())
	}
	if len(rec.Events) != 0 {
		t.Fatal("صفحه دوم نباید AUDIT_READ بسازد")
	}

	first := testutil.NewTestGinContext(http.MethodGet, "/api/logs?ip=203.0.113.9&username=ali", nil)
	first.SetUserID(3)
	first.SetUsername("sara")
	h.List(first.Context)
	if first.Recorder.Code != http.StatusOK || len(rec.Events) != 1 {
		t.Fatalf("صفحه اول باید یک خواندن ثبت کند: code=%d events=%d", first.Recorder.Code, len(rec.Events))
	}
	got := rec.Events[0].Event
	details, _ := got.Details.(map[string]string)
	if got.Type != audit.EventAuditRead || got.Actor.Username != "sara" || details["ip"] != "203.0.113.9" || details["username"] != "ali" {
		t.Fatalf("خواندن ممیزی نادرست است: %+v %#v", got, details)
	}
	raw := first.Recorder.Body.String()
	if details["ip"] == "" || len(raw) == 0 {
		t.Fatal("پاسخ فهرست خالی است")
	}
}
