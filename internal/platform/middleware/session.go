package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/tpdenta/afta-reception/internal/platform/apperror"
	"github.com/tpdenta/afta-reception/internal/platform/security/session"
)

const (
	ContextKeySession   = "session"
	ContextKeyUserID    = "userID"
	ContextKeyRoleName  = "roleName"
	ContextKeySessionID = "sessionID"
)

// publicPaths مسیرهایی که نیاز به نشست ندارند.
var publicPaths = map[string]bool{
	"/api/login":  true,
	"/api/health": true,
}

// SessionMiddleware بررسی نشست از کوکی session_id.
func SessionMiddleware(sessionSvc *session.Service, secureCookies bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		if publicPaths[c.Request.URL.Path] {
			c.Next()
			return
		}

		cookie, err := c.Cookie(session.SessionCookieName)
		if err != nil || cookie == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": apperror.ErrUnauthorized.UserMsg,
				"code":  apperror.ErrUnauthorized.Code,
			})
			return
		}

		sessionID, err := uuid.Parse(cookie)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": apperror.ErrUnauthorized.UserMsg,
				"code":  apperror.ErrUnauthorized.Code,
			})
			return
		}

		sess, err := sessionSvc.GetSession(sessionID)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": apperror.ErrUnauthorized.UserMsg,
				"code":  apperror.ErrUnauthorized.Code,
			})
			return
		}

		c.Set(ContextKeySession, sess)
		c.Set(ContextKeyUserID, sess.PersonnelAccountID)
		c.Set(ContextKeySessionID, sessionID)
		c.Next()
	}
}

// SetSessionCookie کوکی نشست را تنظیم می‌کند.
func SetSessionCookie(c *gin.Context, sessionID uuid.UUID, secure bool) {
	c.SetSameSite(http.SameSiteStrictMode)
	c.SetCookie(session.SessionCookieName, sessionID.String(), 0, "/", "", secure, true)
}

// ClearSessionCookie کوکی نشست را پاک می‌کند.
func ClearSessionCookie(c *gin.Context, secure bool) {
	c.SetSameSite(http.SameSiteStrictMode)
	c.SetCookie(session.SessionCookieName, "", -1, "/", "", secure, true)
}
