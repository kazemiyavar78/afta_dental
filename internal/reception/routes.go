package reception

import (
	"github.com/gin-gonic/gin"
	"github.com/tpdenta/afta-reception/internal/platform/middleware"
)

// RegisterRoutes مسیرهای ماژول پذیرش را ثبت می‌کند.
func RegisterRoutes(r *gin.RouterGroup, h *Handler) {
	r.GET("/reception",
		middleware.RequirePermission("reception.read"),
		middleware.AuthorizationMiddleware(),
		h.List)
	r.GET("/reception/nav",
		middleware.RequirePermission("reception.read"),
		middleware.AuthorizationMiddleware(),
		h.Navigate)
	r.POST("/reception/calculate",
		middleware.RequirePermission("reception.create", "reception.update"),
		middleware.AuthorizationMiddleware(),
		h.Calculate)
	r.GET("/reception/patient/:patientId",
		middleware.RequirePermission("reception.read"),
		middleware.AuthorizationMiddleware(),
		h.ListByPatient)
	r.GET("/reception/patient/:patientId/services-history",
		middleware.RequirePermission("reception.read"),
		middleware.AuthorizationMiddleware(),
		h.PatientServiceHistory)
	r.POST("/reception",
		middleware.RequirePermission("reception.create"),
		middleware.AuthorizationMiddleware(),
		h.Create)
	r.GET("/reception/:id",
		middleware.RequirePermission("reception.read"),
		middleware.AuthorizationMiddleware(),
		h.GetByID)
	r.PUT("/reception/:id",
		middleware.RequirePermission("reception.update"),
		middleware.AuthorizationMiddleware(),
		h.Update)
	r.DELETE("/reception/:id",
		middleware.RequirePermission("reception.delete"),
		middleware.AuthorizationMiddleware(),
		h.Delete)
	r.POST("/reception/:id/restore",
		middleware.RequirePermission("reception.restore"),
		middleware.AuthorizationMiddleware(),
		h.Restore)
	r.POST("/reception/:id/end",
		middleware.RequirePermission("reception.update"),
		middleware.AuthorizationMiddleware(),
		h.EndReception)
	r.POST("/reception/:id/photos",
		middleware.RequirePermission("reception.update"),
		middleware.AuthorizationMiddleware(),
		h.UploadPhoto)
	r.GET("/reception/:id/photos/:photoId",
		middleware.RequirePermission("reception.read"),
		middleware.AuthorizationMiddleware(),
		h.GetPhoto)
}
