package patient

import (
	"github.com/gin-gonic/gin"
	"github.com/tpdenta/afta-reception/internal/platform/middleware"
)

// RegisterRoutes مسیرهای API بیمار را ثبت می‌کند.
func RegisterRoutes(r *gin.RouterGroup, h *Handler) {
	r.GET("/patients", middleware.RequireRole("Admin", "Reception"), middleware.AuthorizationMiddleware(), h.List)
	r.POST("/patients", middleware.RequireRole("Admin", "Reception"), middleware.AuthorizationMiddleware(), h.Create)
	r.GET("/patients/:id", middleware.RequireRole("Admin", "Reception"), middleware.AuthorizationMiddleware(), h.GetByID)
	r.PUT("/patients/:id", middleware.RequireRole("Admin", "Reception"), middleware.AuthorizationMiddleware(), h.Update)
	r.DELETE("/patients/:id", middleware.RequireRole("Admin"), middleware.AuthorizationMiddleware(), h.Delete)
}
