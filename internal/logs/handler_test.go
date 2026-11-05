package logs

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tpdenta/afta-reception/internal/platform/listquery"
	"github.com/tpdenta/afta-reception/internal/platform/middleware"
	"github.com/tpdenta/afta-reception/internal/platform/security/audit"
	"github.com/tpdenta/afta-reception/internal/platform/testutil"
)

type fakeLogs struct {
	page listquery.Page[EventItem]
}

func (f fakeLogs) List(context.Context, listquery.Query, *int) (listquery.Page[EventItem], error) {
	if f.page.Items == nil {
		f.page = listquery.Page[EventItem]{
			Items:    []EventItem{{ID: 1, EventType: audit.EventLoginFailed, EventTypeLabel: "ورود ناموفق"}},
			Total:    1,
			Page:     1,
			PageSize: 20,
		}
	}
	return f.page, nil
}

func (fakeLogs) Meta(context.Context) (MetaResponse, error) {
	return MetaResponse{EventTypes: []Option{}, Reasons: []Option{}, TargetTypes: []Option{}}, nil
}

func (fakeLogs) Integrity(context.Context) (IntegrityResponse, error) {
	return IntegrityResponse{}, nil
}

func mountLogs(h *Handler, perms []string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set(middleware.ContextKeyUserID, 4)
		c.Set(middleware.ContextKeyUsername, "sara")
		c.Set(middleware.ContextKeyPermissions, perms)
		c.Next()
	})
	RegisterRoutes(r.Group("/api"), h)
	return r
}

func get(r http.Handler, path string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, path, nil))
	return w
}

// TestFAU_GEN1_AuditReadOnlyFirstPage صفحه اول جستجو را ثبت می‌کند و صفحه بعد را نه.
func TestFAU_GEN1_AuditReadOnlyFirstPage(t *testing.T) {
	rec := &testutil.FakeAuditRecorder{}
	h := NewHandler(fakeLogs{}, rec)
	r := mountLogs(h, []string{"logs.read"})

	second := get(r, "/api/logs?filter[ip]=203.0.113.9&page=2")
	if second.Code != http.StatusOK {
		t.Fatalf("فهرست: %d %s", second.Code, second.Body.String())
	}
	if len(rec.Events) != 0 {
		t.Fatal("صفحه دوم نباید AUDIT_READ بسازد")
	}

	first := get(r, "/api/logs?filter[ip]=203.0.113.9&filter[actor_username]=ali")
	if first.Code != http.StatusOK || len(rec.Events) != 1 {
		t.Fatalf("صفحه اول باید یک خواندن ثبت کند: code=%d events=%d", first.Code, len(rec.Events))
	}
	got := rec.Events[0].Event
	details, _ := got.Details.(map[string]string)
	if got.Type != audit.EventAuditRead || got.Actor.Username != "sara" || details["ip"] != "203.0.113.9" || details["actor_username"] != "ali" {
		t.Fatalf("خواندن ممیزی نادرست است: %+v %#v", got, details)
	}
	if !strings.Contains(first.Body.String(), `"items"`) {
		t.Fatalf("پاسخ فهرست استاندارد نیست: %s", first.Body.String())
	}
}

// TestR1_01_LogsWithoutPermissionAuditsDenial بدون logs.read پاسخ ۴۰۳ و AUDIT_READ_DENIED می‌دهد.
func TestR1_01_LogsWithoutPermissionAuditsDenial(t *testing.T) {
	rec := &testutil.FakeAuditRecorder{}
	middleware.SetAuthorizationAudit(rec)
	t.Cleanup(func() { middleware.SetAuthorizationAudit(nil) })

	r := mountLogs(NewHandler(fakeLogs{}, rec), []string{"patients.read"})
	for _, path := range []string{"/api/logs", "/api/logs/integrity"} {
		w := get(r, path)
		if w.Code != http.StatusForbidden || !strings.Contains(w.Body.String(), "FORBIDDEN") {
			t.Fatalf("%s باید ۴۰۳ باشد: %d %s", path, w.Code, w.Body.String())
		}
	}
	if len(rec.Events) != 2 {
		t.Fatalf("هر دو مسیر باید رد خواندن را ثبت کنند: %+v", rec.Events)
	}
	for _, ev := range rec.Events {
		if ev.Event.Type != audit.EventAuditReadDenied || ev.Event.Actor.Username != "sara" {
			t.Fatalf("رویداد رد خواندن نادرست است: %+v", ev.Event)
		}
	}
}

// TestR1_01_UnknownSortReturnsValidationError مرتب‌سازی ناشناس را ۴۰۰ می‌کند و مقدار را برنمی‌گرداند.
func TestR1_01_UnknownSortReturnsValidationError(t *testing.T) {
	r := mountLogs(NewHandler(fakeLogs{}, &testutil.FakeAuditRecorder{}), []string{"logs.read"})
	w := get(r, "/api/logs?sort=drop_table")
	if w.Code != http.StatusBadRequest || !strings.Contains(w.Body.String(), "VALIDATION_ERROR") {
		t.Fatalf("مرتب‌سازی ناشناس: %d %s", w.Code, w.Body.String())
	}
	if strings.Contains(w.Body.String(), "drop_table") {
		t.Fatal("مقدار sort نباید در پاسخ باشد")
	}
}

// TestR1_01_ListReturnsPageShape پاسخ فهرست را در قالب items و total و صفحه برمی‌گرداند.
func TestR1_01_ListReturnsPageShape(t *testing.T) {
	h := NewHandler(fakeLogs{page: listquery.Page[EventItem]{
		Items: []EventItem{{
			ID: 9, EventType: audit.EventLoginFailed, EventTypeLabel: "ورود ناموفق", IP: "192.168.1.4",
		}},
		Total: 4, Page: 1, PageSize: 20,
	}}, &testutil.FakeAuditRecorder{})
	w := get(mountLogs(h, []string{"logs.read"}), "/api/logs")
	if w.Code != http.StatusOK {
		t.Fatalf("فهرست: %d %s", w.Code, w.Body.String())
	}
	var body listquery.Page[EventItem]
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Total != 4 || body.Page != 1 || body.PageSize != 20 || len(body.Items) != 1 || body.Items[0].EventTypeLabel != "ورود ناموفق" {
		t.Fatalf("شکل پاسخ نادرست است: %+v", body)
	}
}

type stubChain struct {
	state audit.AuditChainState
}

func (s stubChain) LoadChainState() (audit.AuditChainState, error) {
	return s.state, nil
}

// TestR1_01_MetaReturnsPersianLabels کشوی فیلتر را با برچسب فارسی نوع، دلیل و هدف پر می‌کند.
func TestR1_01_MetaReturnsPersianLabels(t *testing.T) {
	h := NewHandler(NewService(nil, stubChain{}), &testutil.FakeAuditRecorder{})
	w := get(mountLogs(h, []string{"logs.read"}), "/api/logs/meta")
	if w.Code != http.StatusOK {
		t.Fatalf("meta: %d %s", w.Code, w.Body.String())
	}
	raw := w.Body.String()
	for _, label := range []string{"ورود ناموفق", "رمز نادرست", "کاربر"} {
		if !strings.Contains(raw, label) {
			t.Fatalf("برچسب %s در meta نیست: %s", label, raw)
		}
	}
}

// TestPP04_IntegrityRouteReturnsChainStatus نتیجه بررسی زنجیره را بدون هش برمی‌گرداند.
func TestPP04_IntegrityRouteReturnsChainStatus(t *testing.T) {
	checked := time.Date(2026, 3, 1, 8, 30, 0, 0, time.UTC)
	h := NewHandler(NewService(nil, stubChain{state: audit.AuditChainState{
		LastCheckAt: checked,
		LastResult:  audit.ChainResultBroken,
		BrokenAtID:  42,
	}}), &testutil.FakeAuditRecorder{})
	w := get(mountLogs(h, []string{"logs.read"}), "/api/logs/integrity")
	if w.Code != http.StatusOK {
		t.Fatalf("integrity: %d %s", w.Code, w.Body.String())
	}
	if strings.Contains(w.Body.String(), "IntegrityHash") || strings.Contains(w.Body.String(), "LastVerifiedHash") {
		t.Fatal("هش زنجیره نباید در پاسخ باشد")
	}
	var body IntegrityResponse
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.LastResult != audit.ChainResultBroken || body.BrokenAtID != 42 || body.LastCheckAt == nil || !body.LastCheckAt.Equal(checked) {
		t.Fatalf("وضعیت زنجیره نادرست است: %+v", body)
	}
	if body.OpenFindingsCount != 0 || body.OpenFindings == nil {
		t.Fatalf("بدون یافته، شمارش صفر و فهرست خالی است: %+v", body)
	}
}

type stubOpenFindings struct {
	rows []audit.IntegrityFinding
}

func (s stubOpenFindings) ListAllOpen(context.Context) ([]audit.IntegrityFinding, error) {
	return s.rows, nil
}

// TestPP04_IntegrityIncludesOpenFindings تعداد و فهرست نوع، نام و اولین زمان را برمی‌گرداند.
func TestPP04_IntegrityIncludesOpenFindings(t *testing.T) {
	seen := time.Date(2026, 3, 2, 9, 0, 0, 0, time.UTC)
	svc := NewService(nil, stubChain{state: audit.AuditChainState{
		LastCheckAt: seen,
		LastResult:  audit.ChainResultOK,
	}})
	svc.findings = stubOpenFindings{rows: []audit.IntegrityFinding{
		{EntityType: audit.TargetUser, EntityID: "4", EntityName: "mina", FirstSeenAt: seen},
		{EntityType: audit.TargetRole, EntityID: "2", EntityName: "reception", FirstSeenAt: seen.Add(time.Minute)},
	}}
	w := get(mountLogs(NewHandler(svc, &testutil.FakeAuditRecorder{}), []string{"logs.read"}), "/api/logs/integrity")
	if w.Code != http.StatusOK {
		t.Fatalf("integrity: %d %s", w.Code, w.Body.String())
	}
	if strings.Contains(w.Body.String(), "IntegrityHash") || strings.Contains(w.Body.String(), "RowHash") {
		t.Fatal("هش نباید در پاسخ یافته باشد")
	}
	var body IntegrityResponse
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.OpenFindingsCount != 2 || len(body.OpenFindings) != 2 {
		t.Fatalf("باید دو یافته باز باشد: %+v", body)
	}
	if body.OpenFindings[0].EntityType != audit.TargetUser || body.OpenFindings[0].EntityTypeLabel != "کاربر" || body.OpenFindings[0].EntityName != "mina" || !body.OpenFindings[0].FirstSeenAt.Equal(seen) {
		t.Fatalf("یافته کاربر نادرست است: %+v", body.OpenFindings[0])
	}
	if body.OpenFindings[1].EntityTypeLabel != "نقش" || body.OpenFindings[1].EntityName != "reception" {
		t.Fatalf("یافته نقش نادرست است: %+v", body.OpenFindings[1])
	}
}

// TestR1_01_MyLogsIsPaged لاگ خود کاربر را صفحه‌بندی می‌کند و صفحه بعد را ممیزی نمی‌کند.
func TestR1_01_MyLogsIsPaged(t *testing.T) {
	rec := &testutil.FakeAuditRecorder{}
	r := mountLogs(NewHandler(fakeLogs{}, rec), nil)
	second := get(r, "/api/logs/user?page=2")
	if second.Code != http.StatusOK || !strings.Contains(second.Body.String(), `"page_size"`) {
		t.Fatalf("لاگ کاربر: %d %s", second.Code, second.Body.String())
	}
	if len(rec.Events) != 0 {
		t.Fatal("صفحه دوم لاگ خود کاربر نباید AUDIT_READ بسازد")
	}
	first := get(r, "/api/logs/user")
	if first.Code != http.StatusOK || len(rec.Events) != 1 {
		t.Fatalf("صفحه اول لاگ خود کاربر: code=%d events=%d", first.Code, len(rec.Events))
	}
	details, _ := rec.Events[0].Event.Details.(map[string]string)
	if details["scope"] != "self" {
		t.Fatalf("محدوده self ثبت نشد: %#v", details)
	}
}
