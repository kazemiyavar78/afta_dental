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
			userID, _ := c.Get(ContextKeyUserID)
			uid, _ := userID.(int)
			_ = auditMgr.LogEvent(&uid, c.ClientIP(), audit.EventCSRFViolation,
				"توکن CSRF نامعتبر یا ناقص")

			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": apperror.ErrCSRF.UserMsg,
				"code":  apperror.ErrCSRF.Code,
			})
			return
		}

		if !verifyCSRFToken(cookieToken, c, key) {
			userID, _ := c.Get(ContextKeyUserID)
			uid, _ := userID.(int)
			_ = auditMgr.LogEvent(&uid, c.ClientIP(), audit.EventCSRFViolation,
				"توکن CSRF با نشست مطابقت ندارد")

			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": apperror.ErrCSRF.UserMsg,
				"code":  apperror.ErrCSRF.Code,
			})
			return
		}

		c.Next()
	}
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
