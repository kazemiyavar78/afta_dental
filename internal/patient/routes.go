package patient

import (
	"github.com/gin-gonic/gin"
	"github.com/tpdenta/afta-reception/internal/platform/middleware"
)

// RegisterRoutes مسیرهای API بیمار را ثبت می‌کند.
func RegisterRoutes(r *gin.RouterGroup, h *Handler) {
	r.GET("/patients", middleware.RequirePermission("patient.read"), middleware.AuthorizationMiddleware(), h.List)
	r.POST("/patients", middleware.RequirePermission("patient.create"), middleware.AuthorizationMiddleware(), h.Create)
	r.GET("/patients/:id", middleware.RequirePermission("patient.read"), middleware.AuthorizationMiddleware(), h.GetByID)
	r.PUT("/patients/:id", middleware.RequirePermission("patient.update"), middleware.AuthorizationMiddleware(), h.Update)
	r.DELETE("/patients/:id", middleware.RequirePermission("patient.delete"), middleware.AuthorizationMiddleware(), h.Delete)
}
