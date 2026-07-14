// این ماژول نمونه است؛ باقی Entity ها و Endpoint ها طبق همین الگو در فازهای بعدی تکمیل می‌شوند.
package organization

import (
	"github.com/gin-gonic/gin"
	"github.com/tpdenta/afta-reception/internal/platform/middleware"
)

// RegisterRoutes مسیرهای API سازمان را ثبت می‌کند.
func RegisterRoutes(r *gin.RouterGroup, h *Handler) {
	r.GET("/organization", middleware.RequireRole("Admin", "Reception"), middleware.AuthorizationMiddleware(), h.List)
	r.POST("/organization", middleware.RequireRole("Admin"), middleware.AuthorizationMiddleware(), h.Create)
	r.GET("/organization/:id", middleware.RequireRole("Admin", "Reception"), middleware.AuthorizationMiddleware(), h.GetByID)
	r.PUT("/organization/:id", middleware.RequireRole("Admin"), middleware.AuthorizationMiddleware(), h.Update)
	r.DELETE("/organization/:id", middleware.RequireRole("Admin"), middleware.AuthorizationMiddleware(), h.Delete)
}
