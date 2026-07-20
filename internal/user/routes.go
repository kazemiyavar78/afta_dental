package user

import (
	"github.com/gin-gonic/gin"
	"github.com/tpdenta/afta-reception/internal/platform/middleware"
)

// RegisterRoutes مسیرهای ماژول کاربران را ثبت می‌کند.
func RegisterRoutes(r *gin.RouterGroup, h *Handler) {
	// عمومی
	r.POST("/login", h.Login)

	// نیازمند احراز هویت
	auth := r.Group("")
	{
		auth.POST("/logout", h.Logout)
		auth.GET("/me", h.GetMe)
		auth.POST("/change-password", h.ChangePassword)

		// نقش‌های قابل انتصاب (برای فرم کاربر)
		auth.GET("/roles/assignable", middleware.RequirePermission("users.create", "users.update"), middleware.AuthorizationMiddleware(), h.ListAssignableRoles)

		// مدیریت نقش‌ها
		auth.GET("/roles", middleware.RequirePermission("roles.read"), middleware.AuthorizationMiddleware(), h.ListRoles)
		auth.POST("/roles", middleware.RequirePermission("roles.create"), middleware.AuthorizationMiddleware(), h.CreateRole)
		auth.GET("/roles/:id", middleware.RequirePermission("roles.read"), middleware.AuthorizationMiddleware(), h.GetRole)
		auth.PUT("/roles/:id", middleware.RequirePermission("roles.update"), middleware.AuthorizationMiddleware(), h.UpdateRole)
		auth.DELETE("/roles/:id", middleware.RequirePermission("roles.delete"), middleware.AuthorizationMiddleware(), h.DeleteRole)
		auth.GET("/permissions", middleware.RequirePermission("roles.read"), middleware.AuthorizationMiddleware(), h.ListPermissions)

		// مدیریت کاربران
		auth.POST("/users", middleware.RequirePermission("users.create"), middleware.AuthorizationMiddleware(), h.CreateUser)
		auth.GET("/users", middleware.RequirePermission("users.read"), middleware.AuthorizationMiddleware(), h.ListUsers)
		auth.GET("/users/doctors", middleware.RequirePermission("reception.read", "users.read"), middleware.AuthorizationMiddleware(), h.ListDoctors)
		auth.GET("/users/assistants", middleware.RequirePermission("reception.read", "users.read"), middleware.AuthorizationMiddleware(), h.ListAssistants)
		auth.GET("/users/:id", h.GetUser)
		auth.PUT("/users/:id", h.UpdateUser)

		// تنظیمات امنیتی
		auth.GET("/security/settings", middleware.RequirePermission("security.settings"), middleware.AuthorizationMiddleware(), h.ListSecuritySettings)
		auth.PUT("/security/settings", middleware.RequirePermission("security.settings"), middleware.AuthorizationMiddleware(), h.UpdateSecuritySetting)

		// نشست‌ها
		auth.GET("/sessions", h.ListSessions)
		auth.GET("/profile", h.GetUserProfile)
		auth.DELETE("/sessions/:id", h.DeleteSession)
	}
}
