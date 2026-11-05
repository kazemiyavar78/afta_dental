// Package testutil helperهای مشترک تست handler و ممیزی است.
// اتصال پایگاه فقط در dbtest.go و با build tag برابر integration ساخته می‌شود
// تا go test معمولی به SQL Server وصل نشود.
package testutil

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/tpdenta/afta-reception/internal/platform/middleware"
	"github.com/tpdenta/afta-reception/internal/platform/security/session"
)

// GinTest درخواست httptest و context گین برای فراخوانی مستقیم handler است.
type GinTest struct {
	Context  *gin.Context
	Recorder *httptest.ResponseRecorder
}

var ginTestModeOnce sync.Once

// NewTestGinContext یک context گین و ResponseRecorder می‌سازد.
// method و path درخواست HTTP هستند. body بدنه خام است و nil یعنی بدون بدنه.
// خروجی GinTest است تا تست کاربر، نقش، مجوز و نشست را با کلیدهای middleware ست کند.
func NewTestGinContext(method, path string, body []byte) *GinTest {
	ginTestModeOnce.Do(func() {
		gin.SetMode(gin.TestMode)
	})
	if method == "" {
		method = http.MethodGet
	}
	if path == "" {
		path = "/"
	}

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)

	var reader io.Reader
	if body != nil {
		reader = bytes.NewReader(body)
	}
	req := httptest.NewRequest(method, path, reader)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	ctx.Request = req

	return &GinTest{Context: ctx, Recorder: recorder}
}

// SetUserID شناسه کاربر را با کلید ContextKeyUserID در context می‌گذارد.
// userID همان نوع int است که SessionMiddleware از نشست می‌خواند.
func (g *GinTest) SetUserID(userID int) {
	g.Context.Set(middleware.ContextKeyUserID, userID)
}

// SetUsername نام کاربری را با کلید ContextKeyUsername در context می‌گذارد.
// username همان نامی است که SessionMiddleware پس از خواندن کاربر ذخیره می‌کند.
func (g *GinTest) SetUsername(username string) {
	g.Context.Set(middleware.ContextKeyUsername, username)
}

// SetRoleName نام نقش را با کلید ContextKeyRoleName در context می‌گذارد.
// roleName همان رشته‌ای است که SetRoleMiddleware ذخیره می‌کند.
func (g *GinTest) SetRoleName(roleName string) {
	g.Context.Set(middleware.ContextKeyRoleName, roleName)
}

// SetPermissions فهرست مجوز را با کلید ContextKeyPermissions در context می‌گذارد.
// permissions کپی می‌شود تا تغییر بعدی اسلایسِ تست، context را عوض نکند.
func (g *GinTest) SetPermissions(permissions []string) {
	if permissions == nil {
		g.Context.Set(middleware.ContextKeyPermissions, []string(nil))
		return
	}
	copied := make([]string, len(permissions))
	copy(copied, permissions)
	g.Context.Set(middleware.ContextKeyPermissions, copied)
}

// SetSession نشست را با کلید ContextKeySession در context می‌گذارد.
// اگر sess تهی نباشد، شناسه آن هم با ContextKeySessionID از نوع uuid.UUID ست می‌شود،
// مثل SessionMiddleware. شناسه کاربر را عوض نمی‌کند.
func (g *GinTest) SetSession(sess *session.Session) {
	g.Context.Set(middleware.ContextKeySession, sess)
	if sess == nil {
		return
	}
	g.Context.Set(middleware.ContextKeySessionID, sess.ID.UUID())
}
