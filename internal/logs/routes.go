package logs


import (
	"github.com/gin-gonic/gin"
	"github.com/tpdenta/afta-reception/internal/platform/middleware"
)

func RegisterRoutes(r *gin.RouterGroup, h *Handler) {
	r.GET("/logs", middleware.RequirePermission("logs.read"), middleware.AuthorizationMiddleware(), h.List)
	r.GET("/logs/user", h.GetMyLogs)
}
