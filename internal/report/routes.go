package report

import (
	"github.com/gin-gonic/gin"
	"github.com/tpdenta/afta-reception/internal/platform/middleware"
)

// RegisterRoutes مسیرهای گزارش را ثبت می‌کند.
func RegisterRoutes(r *gin.RouterGroup, h *Handler) {
	r.GET("/reports/revenue-by-organization",
		middleware.RequirePermission("reports.read"),
		middleware.AuthorizationMiddleware(),
		h.RevenueByOrganization,
	)
	r.GET("/reports/doctor-performance/by-organization",
		middleware.RequirePermission("reports.read"),
		middleware.AuthorizationMiddleware(),
		h.DoctorPerformanceByOrganization,
	)
	r.GET("/reports/doctor-performance/by-service",
		middleware.RequirePermission("reports.read"),
		middleware.AuthorizationMiddleware(),
		h.DoctorPerformanceByService,
	)
	r.GET("/reports/doctor-performance/patient-list",
		middleware.RequirePermission("reports.read"),
		middleware.AuthorizationMiddleware(),
		h.DoctorPerformancePatientList,
	)
	r.GET("/reports/doctor-performance/revenue-by-doctors",
		middleware.RequirePermission("reports.read"),
		middleware.AuthorizationMiddleware(),
		h.RevenueByDoctors,
	)
	r.GET("/reports/user-performance/by-reception",
		middleware.RequirePermission("reports.read"),
		middleware.AuthorizationMiddleware(),
		h.UserPerformanceByReception,
	)
	r.GET("/reports/user-performance/by-transaction",
		middleware.RequirePermission("reports.read"),
		middleware.AuthorizationMiddleware(),
		h.UserPerformanceByTransaction,
	)
}
