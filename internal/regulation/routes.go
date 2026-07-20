package regulation

import (
	"github.com/gin-gonic/gin"
	"github.com/tpdenta/afta-reception/internal/platform/middleware"
)

// RegisterRoutes مسیرهای API ضوابط را ثبت می‌کند.
func RegisterRoutes(r *gin.RouterGroup, h *Handler) {
	r.GET("/regulations", middleware.RequirePermission("regulation.read"), middleware.AuthorizationMiddleware(), h.List)
	r.POST("/regulations", middleware.RequirePermission("regulation.create"), middleware.AuthorizationMiddleware(), h.Create)
	r.GET("/regulations/:id", middleware.RequirePermission("regulation.read"), middleware.AuthorizationMiddleware(), h.GetByID)
	r.PUT("/regulations/:id", middleware.RequirePermission("regulation.update"), middleware.AuthorizationMiddleware(), h.Update)
	r.DELETE("/regulations/:id", middleware.RequirePermission("regulation.delete"), middleware.AuthorizationMiddleware(), h.Delete)
}
