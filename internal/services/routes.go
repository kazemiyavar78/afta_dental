package services

import (
	"github.com/gin-gonic/gin"
	"github.com/tpdenta/afta-reception/internal/platform/middleware"
)

// RegisterRoutes مسیرهای API خدمات را ثبت می‌کند.
func RegisterRoutes(r *gin.RouterGroup, h *Handler) {
	r.GET("/services", middleware.RequirePermission("services.read"), middleware.AuthorizationMiddleware(), h.List)
	r.POST("/services", middleware.RequirePermission("services.create"), middleware.AuthorizationMiddleware(), h.Create)
	r.GET("/services/:id", middleware.RequirePermission("services.read"), middleware.AuthorizationMiddleware(), h.GetByID)
	r.PUT("/services/:id", middleware.RequirePermission("services.update"), middleware.AuthorizationMiddleware(), h.Update)
	r.DELETE("/services/:id", middleware.RequirePermission("services.delete"), middleware.AuthorizationMiddleware(), h.Delete)
}
