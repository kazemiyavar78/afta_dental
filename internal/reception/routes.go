package reception

import (
	"github.com/gin-gonic/gin"
	"github.com/tpdenta/afta-reception/internal/platform/middleware"
)

// RegisterRoutes مسیرهای ماژول پذیرش را ثبت می‌کند.
func RegisterRoutes(r *gin.RouterGroup, h *Handler) {
	roles := []string{"Admin", "Reception", "Doctor"}

	r.GET("/reception",
		middleware.RequireRole(roles...),
		middleware.AuthorizationMiddleware(),
		h.List)
	r.GET("/reception/nav",
		middleware.RequireRole(roles...),
		middleware.AuthorizationMiddleware(),
		h.Navigate)
	r.POST("/reception/calculate",
		middleware.RequireRole(roles...),
		middleware.AuthorizationMiddleware(),
		h.Calculate)
	r.POST("/reception",
		middleware.RequireRole(roles...),
		middleware.AuthorizationMiddleware(),
		h.Create)
	r.GET("/reception/:id",
		middleware.RequireRole(roles...),
		middleware.AuthorizationMiddleware(),
		h.GetByID)
	r.PUT("/reception/:id",
		middleware.RequireRole(roles...),
		middleware.AuthorizationMiddleware(),
		h.Update)
	r.DELETE("/reception/:id",
		middleware.RequireRole(roles...),
		middleware.AuthorizationMiddleware(),
		h.Delete)
	r.POST("/reception/:id/restore",
		middleware.RequireRole("Admin"),
		middleware.AuthorizationMiddleware(),
		h.Restore)
}
