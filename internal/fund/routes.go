// این ماژول نمونه است؛ باقی Entity ها و Endpoint ها طبق همین الگو در فازهای بعدی تکمیل می‌شوند.
package fund

import (
	"github.com/gin-gonic/gin"
	"github.com/tpdenta/afta-reception/internal/platform/middleware"
)

func RegisterRoutes(r *gin.RouterGroup, h *Handler) {
	r.GET("/fund", middleware.RequirePermission("fund.read"), middleware.AuthorizationMiddleware(), h.List)
	r.POST("/fund", middleware.RequirePermission("fund.create"), middleware.AuthorizationMiddleware(), h.Create)
	r.GET("/fund/:id", middleware.RequirePermission("fund.read"), middleware.AuthorizationMiddleware(), h.GetByID)
}
