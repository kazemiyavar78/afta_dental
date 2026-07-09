package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tpdenta/afta-reception/internal/platform/apperror"
	"github.com/tpdenta/afta-reception/internal/platform/security/audit"
)

// SecurityChecker اینترفیس بررسی SecurityCode کاربر.
type SecurityChecker interface {
	CheckCurrentUserSecurityCode(userID int) (bool, error)
}

// SecurityCheckMiddleware قبل از API، SecurityCode کاربر را بررسی می‌کند.
func SecurityCheckMiddleware(checker SecurityChecker, sessionSvc interface {
	DeleteAllUserSessions(userID int) error
}, auditMgr *audit.Manager) gin.HandlerFunc {
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
			_ = sessionSvc.DeleteAllUserSessions(uid)
			ClearSessionCookie(c, true)
			_ = auditMgr.LogEvent(&uid, c.ClientIP(), audit.EventDataTampering,
				"SecurityCode کاربر نامعتبر — نشست پایان یافت")

			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": apperror.ErrIntegrity.UserMsg,
				"code":  apperror.ErrIntegrity.Code,
			})
			return
		}

		c.Next()
	}
}