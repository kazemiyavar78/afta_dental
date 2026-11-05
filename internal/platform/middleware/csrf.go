package middleware

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/tpdenta/afta-reception/internal/platform/apperror"
	"github.com/tpdenta/afta-reception/internal/platform/security/audit"
)

const CSRFCookieName = "csrf_token"
const CSRFHeaderName = "X-CSRF-Token"

// CSRFMiddleware محافظت Double-Submit Cookie برای متدهای mutating.
func CSRFMiddleware(csrfKey string, auditMgr *audit.Manager) gin.HandlerFunc {
	key := []byte(csrfKey)

	return func(c *gin.Context) {
		method := strings.ToUpper(c.Request.Method)
		if method == http.MethodGet || method == http.MethodHead || method == http.MethodOptions {
			// برای درخواست‌های خواندنی، توکن CSRF را در کوکی ست کن
			setCSRFCookieIfSession(c, key)
			c.Next()
			return
		}

		if publicPaths[c.Request.URL.Path] {
			c.Next()
			return
		}

		cookieToken, err := c.Cookie(CSRFCookieName)
		headerToken := c.GetHeader(CSRFHeaderName)

		if err != nil || cookieToken == "" || headerToken == "" || cookieToken != headerToken {
			if !rejectCSRF(c, auditMgr, "توکن CSRF نامعتبر یا ناقص") {
				return
			}
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": apperror.ErrCSRF.UserMsg,
				"code":  apperror.ErrCSRF.Code,
			})
			return
		}

		if !verifyCSRFToken(cookieToken, c, key) {
			if !rejectCSRF(c, auditMgr, "توکن CSRF با نشست مطابقت ندارد") {
				return
			}
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": apperror.ErrCSRF.UserMsg,
				"code":  apperror.ErrCSRF.Code,
			})
			return
		}

		c.Next()
	}
}

// rejectCSRF نقض CSRF را ثبت می‌کند و اگر ثبت شکست بخورد پاسخ ۵۰۳ می‌نویسد.
// c درخواست است، auditMgr ثبت‌کننده است و description توضیح فارسی ردیف است.
// خروجی true یعنی ثبت شد و صداکننده باید ۴۰۳ بنویسد. false یعنی پاسخ قبلاً نوشته شده.
// D-14 / FAU_GEN.2: اگر کاربر در context نباشد UserID تهی می‌ماند، نه اشاره‌گر به صفر.
// FAU_STG.4 / D-13: شکست ثبت ممیزی نباید نادیده گرفته شود.
func rejectCSRF(c *gin.Context, auditMgr *audit.Manager, description string) bool {
	err := recordCSRFFailure(c, auditMgr, description)
	if err == nil {
		return true
	}
	c.Abort()
	WriteError(c, err)
	return false
}

// recordCSRFFailure رویداد نقض CSRF را با ثبت‌کننده مرکزی می‌نویسد.
// c درخواست است، auditMgr ثبت‌کننده است و description توضیح فارسی است. خروجی خطای ممیزی است.
func recordCSRFFailure(c *gin.Context, auditMgr *audit.Manager, description string) error {
	if auditMgr == nil {
		return audit.Unavailable(apperror.ErrInternal)
	}
	return audit.Unavailable(auditMgr.Record(c.Request.Context(), audit.Event{
		Type:        audit.EventCSRFViolation,
		Result:      audit.ResultFailure,
		Actor:       ActorFromContext(c),
		Description: description,
	}))
}

func setCSRFCookieIfSession(c *gin.Context, key []byte) {
	sessionID, exists := c.Get(ContextKeySessionID)
	if !exists {
		return
	}

	sid, ok := sessionID.(interface{ String() string })
	if !ok {
		return
	}

	token := generateCSRFToken(sid.String(), key)
	secure := c.Request.TLS != nil || c.GetHeader("X-Forwarded-Proto") == "https"
	c.SetSameSite(http.SameSiteStrictMode)
	c.SetCookie(CSRFCookieName, token, 0, "/", "", secure, false)
}

func generateCSRFToken(sessionID string, key []byte) string {
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(sessionID))
	return hex.EncodeToString(mac.Sum(nil))
}

func verifyCSRFToken(token string, c *gin.Context, key []byte) bool {
	sessionID, exists := c.Get(ContextKeySessionID)
	if !exists {
		return false
	}

	sid, ok := sessionID.(interface{ String() string })
	if !ok {
		return false
	}

	expected := generateCSRFToken(sid.String(), key)
	return hmac.Equal([]byte(token), []byte(expected))
}

// SetCSRFCookieOnLogin پس از لاگین توکن CSRF را ست می‌کند.
func SetCSRFCookieOnLogin(c *gin.Context, sessionID string, csrfKey string, secure bool) {
	token := generateCSRFToken(sessionID, []byte(csrfKey))
	c.SetSameSite(http.SameSiteStrictMode)
	c.SetCookie(CSRFCookieName, token, 0, "/", "", secure, false)
}
