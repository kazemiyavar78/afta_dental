package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tpdenta/afta-reception/internal/platform/apperror"
)

// RoleRequirement metadata نقش مورد نیاز برای route.
type RoleRequirement struct {
	Roles []string
}

// AuthorizationMiddleware بررسی نقش/مجوز بر اساس metadata route.
func AuthorizationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		req, exists := c.Get("roleRequirement")
		if !exists {
			c.Next()
			return
		}

		roleReq, ok := req.(RoleRequirement)
		if !ok || len(roleReq.Roles) == 0 {
			c.Next()
			return
		}

		roleName, exists := c.Get(ContextKeyRoleName)
		if !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": apperror.ErrForbidden.UserMsg,
				"code":  apperror.ErrForbidden.Code,
			})
			return
		}

		userRole, ok := roleName.(string)
		if !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": apperror.ErrForbidden.UserMsg,
				"code":  apperror.ErrForbidden.Code,
			})
			return
		}

		for _, r := range roleReq.Roles {
			if r == userRole {
				c.Next()
				return
			}
		}

		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"error": apperror.ErrForbidden.UserMsg,
			"code":  apperror.ErrForbidden.Code,
		})
	}
}

// RequireRole middleware factory برای الزام نقش.
func RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("roleRequirement", RoleRequirement{Roles: roles})
		c.Next()
	}
}

// SetRoleMiddleware نقش کاربر را از دیتابیس در context قرار می‌دهد.
func SetRoleMiddleware(roleLookup func(userID int) (string, error)) gin.HandlerFunc {
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

		role, err := roleLookup(uid)
		if err == nil {
			c.Set(ContextKeyRoleName, role)
		}
		c.Next()
	}
}
