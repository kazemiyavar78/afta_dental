package tariff

import (
	"github.com/gin-gonic/gin"
	"github.com/tpdenta/afta-reception/internal/platform/middleware"
)

// RegisterRoutes مسیرهای API تعرفه را ثبت می‌کند.
func RegisterRoutes(r *gin.RouterGroup, h *Handler) {
	r.POST("/tariff/calculate", middleware.RequireRole("Admin", "Reception"), middleware.AuthorizationMiddleware(), h.CalculateTariffForOrganization)
	r.POST("/tariff/save", middleware.RequireRole("Admin"), middleware.AuthorizationMiddleware(), h.SaveTariffsForOrganization)
	r.GET("/tariff/organization/:organizationId", middleware.RequireRole("Admin", "Reception"), middleware.AuthorizationMiddleware(), h.ListByOrganization)
	r.PUT("/tariff/:id/recalculate", middleware.RequireRole("Admin"), middleware.AuthorizationMiddleware(), h.RecalculateTariff)
}
