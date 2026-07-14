package services

import (
	"github.com/gin-gonic/gin"
	"github.com/tpdenta/afta-reception/internal/platform/middleware"
)

// RegisterRoutes مسیرهای API خدمات را ثبت می‌کند.
func RegisterRoutes(r *gin.RouterGroup, h *Handler) {
	r.GET("/services", middleware.RequireRole("Admin", "Reception"), middleware.AuthorizationMiddleware(), h.List)
	r.POST("/services", middleware.RequireRole("Admin"), middleware.AuthorizationMiddleware(), h.Create)
	r.GET("/services/:id", middleware.RequireRole("Admin", "Reception"), middleware.AuthorizationMiddleware(), h.GetByID)
	r.PUT("/services/:id", middleware.RequireRole("Admin"), middleware.AuthorizationMiddleware(), h.Update)
	r.DELETE("/services/:id", middleware.RequireRole("Admin"), middleware.AuthorizationMiddleware(), h.Delete)
}
