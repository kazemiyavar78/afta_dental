package logs

import (
	"github.com/gin-gonic/gin"
	"github.com/tpdenta/afta-reception/internal/platform/middleware"
)

// RegisterRoutes مسیرهای خواندن لاگ را ثبت می‌کند.
// r گروه /api است و h هندلر لاگ است. خروجی ندارد.
// meta و integrity قبل از مسیرهای ثابت دیگر ثبت می‌شوند و هر دو مجوز logs.read دارند.
func RegisterRoutes(r *gin.RouterGroup, h *Handler) {
	r.GET("/logs/meta", middleware.RequirePermission("logs.read"), middleware.AuthorizationMiddleware(), h.Meta)
	r.GET("/logs/integrity", middleware.RequirePermission("logs.read"), middleware.AuthorizationMiddleware(), h.Integrity)
	r.GET("/logs", middleware.RequirePermission("logs.read"), middleware.AuthorizationMiddleware(), h.List)
	r.GET("/logs/user", h.GetMyLogs)
}
