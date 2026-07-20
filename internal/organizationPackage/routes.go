package organizationpackage

import (
	"github.com/gin-gonic/gin"
	"github.com/tpdenta/afta-reception/internal/platform/middleware"
)

// RegisterRoutes مسیرهای API بسته‌های تعرفه را ثبت می‌کند.
func RegisterRoutes(r *gin.RouterGroup, h *Handler) {
	r.GET("/organization-packages", middleware.RequirePermission("organization_packages.read"), middleware.AuthorizationMiddleware(), h.List)
	r.POST("/organization-packages", middleware.RequirePermission("organization_packages.create"), middleware.AuthorizationMiddleware(), h.Create)
	r.GET("/organization-packages/:id", middleware.RequirePermission("organization_packages.read"), middleware.AuthorizationMiddleware(), h.GetByID)
	r.PUT("/organization-packages/:id", middleware.RequirePermission("organization_packages.update"), middleware.AuthorizationMiddleware(), h.Update)
	r.DELETE("/organization-packages/:id", middleware.RequirePermission("organization_packages.delete"), middleware.AuthorizationMiddleware(), h.Delete)
}
