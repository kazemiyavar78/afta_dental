// این ماژول نمونه است؛ باقی Entity ها و Endpoint ها طبق همین الگو در فازهای بعدی تکمیل می‌شوند.
package reception

import (
	"github.com/gin-gonic/gin"
	"github.com/tpdenta/afta-reception/internal/platform/middleware"
)

// RegisterRoutes مسیرهای ماژول پذیرش را ثبت می‌کند.
func RegisterRoutes(r *gin.RouterGroup, h *Handler) {
	r.GET("/reception",
		middleware.RequireRole("Admin", "Reception", "Doctor"),
		middleware.AuthorizationMiddleware(),
		h.List)
	r.POST("/reception",
		middleware.RequireRole("Admin", "Reception", "Doctor"),
		middleware.AuthorizationMiddleware(),
		h.Create)
	r.GET("/reception/:id",
		middleware.RequireRole("Admin", "Reception", "Doctor"),
		middleware.AuthorizationMiddleware(),
		h.GetByID)
}
