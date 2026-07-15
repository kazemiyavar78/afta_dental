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
		auth.GET("/roles/assignable", h.ListAssignableRoles)

		// مدیریت نقش‌ها
		auth.GET("/roles", middleware.RequireRole("Admin"), middleware.AuthorizationMiddleware(), h.ListRoles)
		auth.POST("/roles", middleware.RequireRole("Admin"), middleware.AuthorizationMiddleware(), h.CreateRole)
		auth.GET("/roles/:id", middleware.RequireRole("Admin"), middleware.AuthorizationMiddleware(), h.GetRole)
		auth.PUT("/roles/:id", middleware.RequireRole("Admin"), middleware.AuthorizationMiddleware(), h.UpdateRole)
		auth.DELETE("/roles/:id", middleware.RequireRole("Admin"), middleware.AuthorizationMiddleware(), h.DeleteRole)
		auth.GET("/permissions", middleware.RequireRole("Admin"), middleware.AuthorizationMiddleware(), h.ListPermissions)

		// مدیریت کاربران
		auth.POST("/users", middleware.RequireRole("Admin"), middleware.AuthorizationMiddleware(), h.CreateUser)
		auth.GET("/users", middleware.RequireRole("Admin"), middleware.AuthorizationMiddleware(), h.ListUsers)
		auth.GET("/users/:id", h.GetUser)
		auth.PUT("/users/:id", h.UpdateUser)

		// تنظیمات امنیتی
		auth.GET("/security/settings", middleware.RequireRole("Admin"), middleware.AuthorizationMiddleware(), h.ListSecuritySettings)
		auth.PUT("/security/settings", middleware.RequireRole("Admin"), middleware.AuthorizationMiddleware(), h.UpdateSecuritySetting)

		// نشست‌ها
		auth.GET("/sessions", h.ListSessions)
		auth.GET("/profile", h.GetUserProfile)
		auth.DELETE("/sessions/:id", h.DeleteSession)
	}
}
