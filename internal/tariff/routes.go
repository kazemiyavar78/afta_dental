// این ماژول نمونه است؛ باقی Entity ها و Endpoint ها طبق همین الگو در فازهای بعدی تکمیل می‌شوند.
package tariff

import (
	"github.com/gin-gonic/gin"
	"github.com/tpdenta/afta-reception/internal/platform/middleware"
)

func RegisterRoutes(r *gin.RouterGroup, h *Handler) {
	r.GET("/tariff", middleware.RequireRole("Admin", "Reception"), middleware.AuthorizationMiddleware(), h.List)
	r.POST("/tariff", middleware.RequireRole("Admin"), middleware.AuthorizationMiddleware(), h.Create)
	r.GET("/tariff/:id", middleware.RequireRole("Admin", "Reception"), middleware.AuthorizationMiddleware(), h.GetByID)
}
