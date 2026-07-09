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
		auth.GET("/roles", h.ListRoles)
		auth.POST("/change-password", h.ChangePassword)

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
		auth.DELETE("/sessions/:id", h.DeleteSession)
	}
}
