package cardreader

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tpdenta/afta-reception/internal/platform/middleware"
)

// Handler هندلر HTTP شبیه‌ساز کارتخوان.
type Handler struct{ service *Service }

// NewHandler هندلر کارتخوان را می‌سازد.
func NewHandler(s *Service) *Handler { return &Handler{service: s} }

// SimulateCharge پاسخ آزمایشی کارتخوان را برمی‌گرداند.
func (h *Handler) SimulateCharge(c *gin.Context) {
	var req ChargeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.WriteError(c, err)
		return
	}
	resp, err := h.service.SimulateCharge(req)
	if err != nil {
		middleware.WriteError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// RegisterRoutes مسیرهای API کارتخوان آزمایشی را ثبت می‌کند.
func RegisterRoutes(r *gin.RouterGroup, h *Handler) {
	r.POST("/card-reader/simulate", middleware.RequirePermission("wallet.card_reader"), middleware.AuthorizationMiddleware(), h.SimulateCharge)
}
