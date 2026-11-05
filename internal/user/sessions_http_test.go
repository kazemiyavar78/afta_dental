package user

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/tpdenta/afta-reception/internal/platform/listquery"
	"github.com/tpdenta/afta-reception/internal/platform/middleware"
	"github.com/tpdenta/afta-reception/internal/platform/security/audit"
	"github.com/tpdenta/afta-reception/internal/platform/security/session"
)

type fakeDirectory struct {
	page   listquery.Page[session.ListedSession]
	called bool
}

func (f *fakeDirectory) Search(context.Context, listquery.Query, int, bool) (listquery.Page[session.ListedSession], error) {
	f.called = true
	if f.page.Items == nil {
		f.page.Items = []session.ListedSession{}
	}
	return f.page, nil
}

func (f *fakeDirectory) DeleteSession(context.Context, uuid.UUID, audit.Actor, bool) error {
	return nil
}

func mountSessions(h *Handler, perms []string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set(middleware.ContextKeyUserID, 3)
		if perms != nil {
			c.Set(middleware.ContextKeyPermissions, perms)
		}
		c.Next()
	})
	RegisterRoutes(r.Group("/api"), h)
	return r
}

func getSessions(r http.Handler, path string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, path, nil))
	return w
}

// TestR1_05_UnknownSortReturnsValidationError مرتب‌سازی ناشناس نشست را ۴۰۰ می‌کند.
func TestR1_05_UnknownSortReturnsValidationError(t *testing.T) {
	dir := &fakeDirectory{}
	w := getSessions(mountSessions(&Handler{sessionSvc: dir}, nil), "/api/sessions?sort=drop_table")
	if w.Code != http.StatusBadRequest || !strings.Contains(w.Body.String(), "VALIDATION_ERROR") {
		t.Fatalf("مرتب‌سازی ناشناس: %d %s", w.Code, w.Body.String())
	}
	if dir.called || strings.Contains(w.Body.String(), "drop_table") {
		t.Fatal("ستون ناشناس نباید به پایگاه برسد یا در پاسخ بیاید")
	}
}

// TestR1_05_NonAdminCannotSortByUserID کاربر عادی ستون user_id را ندارد.
func TestR1_05_NonAdminCannotSortByUserID(t *testing.T) {
	dir := &fakeDirectory{}
	w := getSessions(mountSessions(&Handler{sessionSvc: dir}, []string{"users.read"}), "/api/sessions?sort=user_id")
	if w.Code != http.StatusBadRequest {
		t.Fatalf("user_id برای غیرمدیر: %d %s", w.Code, w.Body.String())
	}
	if dir.called {
		t.Fatal("نباید جستجو انجام شود")
	}
}

// TestR1_05_ListReturnsPageShape پاسخ نشست صفحه استاندارد است و برای مدیر user_id قابل مرتب‌سازی است.
func TestR1_05_ListReturnsPageShape(t *testing.T) {
	when := time.Date(2026, 4, 5, 6, 7, 8, 0, time.UTC)
	dir := &fakeDirectory{page: listquery.Page[session.ListedSession]{
		Items: []session.ListedSession{{
			ID: "6ba7b810-9dad-11d1-80b4-00c04fd430c8", Username: "ali", IP: "192.168.1.8",
			Browser: "Browser", CreationTime: when, UserID: 3,
		}},
		Total: 2, Page: 1, PageSize: 20,
	}}
	w := getSessions(mountSessions(&Handler{sessionSvc: dir}, []string{"users.listSessions"}), "/api/sessions?sort=user_id&order=asc")
	if w.Code != http.StatusOK {
		t.Fatalf("فهرست نشست: %d %s", w.Code, w.Body.String())
	}
	var body listquery.Page[SessionResponse]
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Total != 2 || body.Page != 1 || body.PageSize != 20 || len(body.Items) != 1 {
		t.Fatalf("شکل پاسخ نادرست است: %+v", body)
	}
	item := body.Items[0]
	if item.Username != "ali" || item.Ip != "192.168.1.8" || item.UserID != 3 || item.CreationTime == "" {
		t.Fatalf("ردیف نشست نادرست است: %+v", item)
	}
}
