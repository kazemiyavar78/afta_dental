package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tpdenta/afta-reception/internal/platform/apperror"
)

const (
	ContextKeyPermissions = "permissions"
	ContextKeyIsAdmin     = "isAdmin"
)

// RoleRequirement metadata نقش مورد نیاز برای route.
type RoleRequirement struct {
	Roles []string
}

// PermissionRequirement metadata مجوز مورد نیاز برای route (هر کدام کافی است).
type PermissionRequirement struct {
	Permissions []string
}

// AuthorizationMiddleware بررسی نقش/مجوز بر اساس metadata route.
func AuthorizationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !authorizePermissions(c) {
			return
		}
		if !authorizeRoles(c) {
			return
		}
		c.Next()
	}
}

// authorizePermissions در صورت وجود الزام مجوز، دسترسی را بررسی می‌کند.
func authorizePermissions(c *gin.Context) bool {
	req, exists := c.Get("permissionRequirement")
	if !exists {
		return true
	}

	permReq, ok := req.(PermissionRequirement)
	if !ok || len(permReq.Permissions) == 0 {
		return true
	}

	if IsAdmin(c) {
		return true
	}

	userPerms := GetPermissions(c)
	for _, needed := range permReq.Permissions {
		if hasPermission(userPerms, needed) {
			return true
		}
	}

	c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
		"error": apperror.ErrForbidden.UserMsg,
		"code":  apperror.ErrForbidden.Code,
	})
	return false
}

// authorizeRoles در صورت وجود الزام نقش، دسترسی را بررسی می‌کند.
func authorizeRoles(c *gin.Context) bool {
	req, exists := c.Get("roleRequirement")
	if !exists {
		return true
	}

	roleReq, ok := req.(RoleRequirement)
	if !ok || len(roleReq.Roles) == 0 {
		return true
	}

	if IsAdmin(c) {
		return true
	}

	roleName, exists := c.Get(ContextKeyRoleName)
	if !exists {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"error": apperror.ErrForbidden.UserMsg,
			"code":  apperror.ErrForbidden.Code,
		})
		return false
	}

	userRole, ok := roleName.(string)
	if !ok {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"error": apperror.ErrForbidden.UserMsg,
			"code":  apperror.ErrForbidden.Code,
		})
		return false
	}

	for _, r := range roleReq.Roles {
		if r == userRole {
			return true
		}
	}

	c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
		"error": apperror.ErrForbidden.UserMsg,
		"code":  apperror.ErrForbidden.Code,
	})
	return false
}

// RequireRole middleware factory برای الزام نقش.
func RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("roleRequirement", RoleRequirement{Roles: roles})
		c.Next()
	}
}

// RequirePermission middleware factory برای الزام حداقل یکی از مجوزها.
func RequirePermission(permissions ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("permissionRequirement", PermissionRequirement{Permissions: permissions})
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

// SetPermissionsMiddleware مجوزهای کاربر و وضعیت ادمین را در context قرار می‌دهد.
func SetPermissionsMiddleware(lookup func(userID int) (permissions []string, isAdmin bool, err error)) gin.HandlerFunc {
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

		perms, isAdmin, err := lookup(uid)
		if err == nil {
			c.Set(ContextKeyPermissions, perms)
			c.Set(ContextKeyIsAdmin, isAdmin)
		}
		c.Next()
	}
}

// GetPermissions لیست مجوزهای کاربر را از context برمی‌گرداند.
func GetPermissions(c *gin.Context) []string {
	raw, exists := c.Get(ContextKeyPermissions)
	if !exists {
		return nil
	}
	perms, ok := raw.([]string)
	if !ok {
		return nil
	}
	return perms
}

// IsAdmin بررسی می‌کند کاربر تمام مجوزهای سیستم را دارد.
func IsAdmin(c *gin.Context) bool {
	raw, exists := c.Get(ContextKeyIsAdmin)
	if !exists {
		return false
	}
	isAdmin, ok := raw.(bool)
	return ok && isAdmin
}

// hasPermission وجود یک مجوز در لیست را بررسی می‌کند.
func hasPermission(permissions []string, permission string) bool {
	for _, p := range permissions {
		if p == permission {
			return true
		}
	}
	return false
}
