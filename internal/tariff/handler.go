package tariff

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tpdenta/afta-reception/internal/platform/middleware"
)

// Handler کنترلر HTTP ماژول تعرفه.
type Handler struct{ service *Service }

// NewHandler نمونه Handler تعرفه می‌سازد.
func NewHandler(s *Service) *Handler { return &Handler{service: s} }

// CalculateTariffForOrganization تست محاسبه تعرفه برای سازمان را انجام می‌دهد (صندوق منفی مجاز است).
func (h *Handler) CalculateTariffForOrganization(c *gin.Context) {
	uid, _ := c.Get(middleware.ContextKeyUserID)
	var req CalculateTariffForOrganizationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.WriteError(c, err)
		return
	}
	resp, err := h.service.CalculateTariffForOrganization(req, uid.(int), c.ClientIP())
	if err != nil {
		middleware.WriteError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// SaveTariffsForOrganization تعرفه‌های محاسبه‌شده را در یک تراکنش ذخیره یا بروزرسانی می‌کند.
func (h *Handler) SaveTariffsForOrganization(c *gin.Context) {
	uid, _ := c.Get(middleware.ContextKeyUserID)
	var req CalculateTariffForOrganizationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.WriteError(c, err)
		return
	}
	resp, err := h.service.SaveTariffsForOrganization(req, uid.(int), c.ClientIP())
	if err != nil {
		middleware.WriteError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// ListByOrganization لیست تعرفه‌های ذخیره‌شده یک سازمان را برمی‌گرداند.
func (h *Handler) ListByOrganization(c *gin.Context) {
	var uri struct {
		OrganizationID uint `uri:"organizationId" binding:"required"`
	}
	if err := c.ShouldBindUri(&uri); err != nil {
		middleware.WriteError(c, err)
		return
	}
	resp, err := h.service.ListByOrganizationID(uri.OrganizationID)
	if err != nil {
		middleware.WriteError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// RecalculateTariff یک تعرفه خاص را با سه مبلغ مرکز بازمحاسبه و ذخیره می‌کند.
func (h *Handler) RecalculateTariff(c *gin.Context) {
	uid, _ := c.Get(middleware.ContextKeyUserID)
	var uri struct {
		ID uint `uri:"id" binding:"required"`
	}
	if err := c.ShouldBindUri(&uri); err != nil {
		middleware.WriteError(c, err)
		return
	}
	var req RecalculateTariffRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.WriteError(c, err)
		return
	}
	resp, err := h.service.RecalculateTariff(uri.ID, req, uid.(int), c.ClientIP())
	if err != nil {
		middleware.WriteError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}
