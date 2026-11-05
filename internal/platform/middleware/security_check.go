package middleware

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tpdenta/afta-reception/internal/platform/apperror"
	"github.com/tpdenta/afta-reception/internal/platform/security/audit"
)

// SecurityChecker اینترفیس بررسی SecurityCode کاربر.
type SecurityChecker interface {
	CheckCurrentUserSecurityCode(userID int) (bool, error)
}

// sessionCloser نشست‌های یک کاربر را با دلیل ممیزی می‌بندد.
// *session.Service همین متد را دارد.
type sessionCloser interface {
	DeleteAllUserSessions(ctx context.Context, userID int, actor audit.Actor, reason string) error
}

// SecurityCheckMiddleware قبل از API، SecurityCode کاربر را بررسی می‌کند.
// checker بررسی کد است و sessions حذف نشست‌ها. شکست کد، نشست‌ها را با دلیل security_code_failed می‌بندد.
func SecurityCheckMiddleware(checker SecurityChecker, sessions sessionCloser) gin.HandlerFunc {
	return func(c *gin.Context) {
		if publicPaths[c.Request.URL.Path] {
			c.Next()
			return
		}

		userID, exists := c.Get(ContextKeyUserID)
		if !exists {
			c.Next()
			return
		}

		uid, ok := userID.(int)
		if !ok {
			c.Next()
			return
		}

		valid, err := checker.CheckCurrentUserSecurityCode(uid)
		if err != nil {
			WriteError(c, err)
			c.Abort()
			return
		}

		if !valid {
			actor := audit.SystemActor()
			actor.IP = c.ClientIP()
			if c.Request != nil {
				actor.UserAgent = c.Request.UserAgent()
			}
			if err := sessions.DeleteAllUserSessions(c.Request.Context(), uid, actor, audit.ReasonSecurityCodeFailed); err != nil {
				WriteError(c, err)
				c.Abort()
				return
			}
			ClearSessionCookie(c, true)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": apperror.ErrIntegrity.UserMsg,
				"code":  apperror.ErrIntegrity.Code,
			})
			return
		}

		c.Next()
	}
}
