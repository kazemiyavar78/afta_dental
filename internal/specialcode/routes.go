package specialcode

import (
	"github.com/gin-gonic/gin"
	"github.com/tpdenta/afta-reception/internal/platform/middleware"
)

// RegisterRoutes مسیرهای API کد خاص را ثبت می‌کند.
func RegisterRoutes(r *gin.RouterGroup, h *Handler) {
	r.GET("/special-codes", middleware.RequirePermission("special_code.read"), middleware.AuthorizationMiddleware(), h.List)
	r.POST("/special-codes", middleware.RequirePermission("special_code.create"), middleware.AuthorizationMiddleware(), h.Create)
	r.GET("/special-codes/by-code/:code", middleware.RequirePermission("special_code.read"), middleware.AuthorizationMiddleware(), h.GetByCode)
	r.GET("/special-codes/:id", middleware.RequirePermission("special_code.read"), middleware.AuthorizationMiddleware(), h.GetByID)
	r.PUT("/special-codes/:id", middleware.RequirePermission("special_code.update"), middleware.AuthorizationMiddleware(), h.Update)
	r.DELETE("/special-codes/:id", middleware.RequirePermission("special_code.delete"), middleware.AuthorizationMiddleware(), h.Delete)
}
