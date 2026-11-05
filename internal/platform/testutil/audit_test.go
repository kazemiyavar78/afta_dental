package testutil

import (
	"testing"

	"github.com/tpdenta/afta-reception/internal/platform/security/audit"
)

// TestFakeAuditRecorder_RecordsActorAndDescription فاعل و توضیح ذخیره‌شده را بررسی می‌کند.
func TestFakeAuditRecorder_RecordsActorAndDescription(t *testing.T) {
	rec := &FakeAuditRecorder{}
	actorID := 4

	if err := rec.LogEvent(&actorID, "10.0.0.8", audit.EventLogout, "خروج کاربر"); err != nil {
		t.Fatalf("LogEvent نباید خطا بدهد: %v", err)
	}
	actorID = 99
	if err := rec.LogEvent(nil, "10.0.0.9", audit.EventLoginFailed, "فاعل ناشناس"); err != nil {
		t.Fatalf("LogEvent با کاربر تهی نباید خطا بدهد: %v", err)
	}

	if len(rec.Events) != 2 {
		t.Fatalf("باید ۲ رویداد ذخیره شود، دریافت %d", len(rec.Events))
	}

	first := rec.Events[0]
	if first.UserID == nil || *first.UserID != 4 {
		t.Fatalf("شناسه فاعل باید ۴ بماند، دریافت %#v", first.UserID)
	}
	if first.IP != "10.0.0.8" || first.EventType != audit.EventLogout || first.Description != "خروج کاربر" {
		t.Fatalf("رویداد اول نادرست است: %+v", first)
	}

	second := rec.Events[1]
	if second.UserID != nil {
		t.Fatal("رویداد دوم نباید شناسه کاربر داشته باشد")
	}
	if second.EventType != audit.EventLoginFailed || second.Description != "فاعل ناشناس" {
		t.Fatalf("رویداد دوم نادرست است: %+v", second)
	}
}
