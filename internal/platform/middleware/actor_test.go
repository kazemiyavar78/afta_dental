package middleware_test

import (
	"net/http"
	"testing"

	"github.com/tpdenta/afta-reception/internal/platform/middleware"
	"github.com/tpdenta/afta-reception/internal/platform/testutil"
)

// TestActorFromContext_ReadsUserIPAndAgent شناسه، نام، IP و User-Agent را از context تست می‌خواند.
func TestActorFromContext_ReadsUserIPAndAgent(t *testing.T) {
	g := testutil.NewTestGinContext(http.MethodGet, "/api/logs", nil)
	g.SetUserID(9)
	g.SetUsername("sara")
	g.Context.Request.RemoteAddr = "198.51.100.20:5555"
	g.Context.Request.Header.Set("User-Agent", "UnitTest")

	actor := middleware.ActorFromContext(g.Context)
	if actor.UserID == nil || *actor.UserID != 9 {
		t.Fatalf("شناسه فاعل: %v", actor.UserID)
	}
	if actor.Username != "sara" {
		t.Fatalf("نام کاربری: %s", actor.Username)
	}
	if actor.UserAgent != "UnitTest" {
		t.Fatalf("User-Agent: %s", actor.UserAgent)
	}
	if actor.IP != "198.51.100.20" {
		t.Fatalf("IP: %s", actor.IP)
	}
}

// TestD14_ActorFromContextWithoutUserIsNil نبودن کاربر اشاره‌گر به صفر نمی‌سازد.
func TestD14_ActorFromContextWithoutUserIsNil(t *testing.T) {
	g := testutil.NewTestGinContext(http.MethodPost, "/api/users", nil)
	actor := middleware.ActorFromContext(g.Context)
	if actor.UserID != nil {
		t.Fatalf("بدون کاربر UserID باید تهی باشد: %d", *actor.UserID)
	}
	if actor.Username != "" {
		t.Fatalf("بدون نام، نام باید خالی باشد: %s", actor.Username)
	}
}

// TestD14_ActorFromContextIgnoresNonIntUser مقدار غیر int را مثل کاربر غایب می‌بیند.
func TestD14_ActorFromContextIgnoresNonIntUser(t *testing.T) {
	g := testutil.NewTestGinContext(http.MethodPost, "/api/users", nil)
	g.Context.Set(middleware.ContextKeyUserID, "1")
	actor := middleware.ActorFromContext(g.Context)
	if actor.UserID != nil {
		t.Fatal("شناسه غیر int نباید ذخیره شود")
	}
}
