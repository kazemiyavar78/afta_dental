package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/tpdenta/afta-reception/internal/platform/apperror"
	"github.com/tpdenta/afta-reception/internal/platform/security/audit"
)

// authorizationAudit ثبت رد خواندن لاگ است و در راه‌اندازی سرور ست می‌شود.
// nil یعنی این فرایند رویداد AUDIT_READ_DENIED نمی‌نویسد (تست‌هایی که آن را لازم ندارند).
var authorizationAudit audit.Recorder

// SetAuthorizationAudit ثبت‌کننده رد دسترسی به صفحه لاگ را نگه می‌دارد.
// rec همان Manager سرور است. خروجی ندارد. FAU_GEN.1: تلاش ناموفق خواندن ممیزی باید بماند.
func SetAuthorizationAudit(rec audit.Recorder) {
	authorizationAudit = rec
}

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

	forbid(c)
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
		forbid(c)
		return false
	}

	userRole, ok := roleName.(string)
	if !ok {
		forbid(c)
		return false
	}

	for _, r := range roleReq.Roles {
		if r == userRole {
			return true
		}
	}

	forbid(c)
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

// forbid پاسخ ۴۰۳ می‌دهد و اگر مسیر خواندن لاگ باشد AUDIT_READ_DENIED را ثبت می‌کند.
// c درخواست ردشده است. اگر ثبت ممیزی شکست بخورد پاسخ ۵۰۳ است تا خطا بی‌صدا نماند.
func forbid(c *gin.Context) {
	if isAuditLogPath(c) {
		if err := recordAuditReadDenied(c); err != nil {
			WriteError(c, err)
			c.Abort()
			return
		}
	}
	c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
		"error": apperror.ErrForbidden.UserMsg,
		"code":  apperror.ErrForbidden.Code,
	})
}

// isAuditLogPath می‌گوید درخواست خواندن فهرست لاگ امنیتی است.
// c درخواست است. خروجی true برای /api/logs و زیرمسیرهای آن است.
func isAuditLogPath(c *gin.Context) bool {
	if c == nil || c.Request == nil || c.Request.URL == nil {
		return false
	}
	path := c.Request.URL.Path
	return path == "/api/logs" || strings.HasPrefix(path, "/api/logs/")
}

// recordAuditReadDenied رد خواندن ممیزی را با فاعل context ثبت می‌کند.
// c درخواست است. اگر ثبت‌کننده تنظیم نشده باشد nil برمی‌گردد تا تست‌های دیگر نشکنند.
// خروجی خطای ممیزی است وقتی ثبت‌کننده هست و نوشتن شکست می‌خورد.
func recordAuditReadDenied(c *gin.Context) error {
	if authorizationAudit == nil {
		return nil
	}
	ctx := context.Background()
	if c != nil && c.Request != nil {
		ctx = c.Request.Context()
	}
	path := ""
	if c != nil && c.Request != nil && c.Request.URL != nil {
		path = c.Request.URL.Path
	}
	return audit.Unavailable(authorizationAudit.Record(ctx, audit.Event{
		Type:        audit.EventAuditReadDenied,
		Result:      audit.ResultFailure,
		Actor:       ActorFromContext(c),
		Target:      audit.Target{Name: path},
		Description: "رد خواندن ممیزی",
	}))
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
