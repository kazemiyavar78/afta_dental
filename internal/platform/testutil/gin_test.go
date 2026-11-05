package testutil

import (
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/tpdenta/afta-reception/internal/platform/middleware"
	"github.com/tpdenta/afta-reception/internal/platform/security/session"
)

// TestNewTestGinContext_SetsMiddlewareKeys کلیدهای context را با middleware یکی می‌کند.
func TestNewTestGinContext_SetsMiddlewareKeys(t *testing.T) {
	body := []byte(`{"name":"نمونه"}`)
	g := NewTestGinContext(http.MethodPost, "/api/patients", body)
	g.SetUserID(7)
	g.SetRoleName("Reception")
	perms := []string{"patient.read", "patient.create"}
	g.SetPermissions(perms)
	perms[0] = "changed-after-set"

	sess := &session.Session{
		ID:                 session.NewMSSQLUUID(),
		PersonnelAccountID: 7,
	}
	g.SetSession(sess)

	if g.Context.Request.Method != http.MethodPost {
		t.Fatalf("method باید POST باشد، دریافت %s", g.Context.Request.Method)
	}
	if g.Context.Request.URL.Path != "/api/patients" {
		t.Fatalf("path نادرست است: %s", g.Context.Request.URL.Path)
	}
	if g.Recorder == nil {
		t.Fatal("ResponseRecorder ساخته نشده است")
	}

	gotUser, ok := g.Context.Get(middleware.ContextKeyUserID)
	if !ok || gotUser != 7 {
		t.Fatalf("userID در context نیست: %v", gotUser)
	}
	gotRole, ok := g.Context.Get(middleware.ContextKeyRoleName)
	if !ok || gotRole != "Reception" {
		t.Fatalf("roleName در context نیست: %v", gotRole)
	}
	gotPerms, ok := g.Context.Get(middleware.ContextKeyPermissions)
	if !ok {
		t.Fatal("permissions در context نیست")
	}
	stored, ok := gotPerms.([]string)
	if !ok || len(stored) != 2 || stored[0] != "patient.read" || stored[1] != "patient.create" {
		t.Fatalf("permissions کپی نشده است: %#v", gotPerms)
	}

	gotSession, ok := g.Context.Get(middleware.ContextKeySession)
	if !ok || gotSession != sess {
		t.Fatal("session با کلید middleware ست نشده است")
	}
	gotSID, ok := g.Context.Get(middleware.ContextKeySessionID)
	if !ok {
		t.Fatal("sessionID در context نیست")
	}
	sid, ok := gotSID.(uuid.UUID)
	if !ok || sid != sess.ID.UUID() {
		t.Fatalf("sessionID باید uuid.UUID باشد: %#v", gotSID)
	}
}
