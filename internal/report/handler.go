package report

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tpdenta/afta-reception/internal/platform/middleware"
)

// Handler کنترلر HTTP گزارش‌ها.
type Handler struct {
	service *Service
}

// NewHandler نمونه Handler گزارش می‌سازد.
func NewHandler(s *Service) *Handler {
	return &Handler{service: s}
}

// RevenueByOrganization گزارش درآمد به تفکیک سازمان را برمی‌گرداند.
func (h *Handler) RevenueByOrganization(c *gin.Context) {
	var q RevenueByOrganizationQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		middleware.WriteError(c, err)
		return
	}

	resp, err := h.service.RevenueByOrganization(q)
	if err != nil {
		middleware.WriteError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// DoctorPerformanceByOrganization گزارش کارکرد پزشکان به تفکیک سازمان را برمی‌گرداند.
func (h *Handler) DoctorPerformanceByOrganization(c *gin.Context) {
	var q DoctorPerformanceByOrganizationQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		middleware.WriteError(c, err)
		return
	}

	resp, err := h.service.DoctorPerformanceByOrganization(q)
	if err != nil {
		middleware.WriteError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// DoctorPerformanceByService گزارش کارکرد پزشکان به تفکیک خدمت را برمی‌گرداند.
func (h *Handler) DoctorPerformanceByService(c *gin.Context) {
	var q DoctorPerformanceByServiceQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		middleware.WriteError(c, err)
		return
	}

	resp, err := h.service.DoctorPerformanceByService(q)
	if err != nil {
		middleware.WriteError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// DoctorPerformancePatientList لیست بیماران پذیرش‌شده توسط پزشکان را برمی‌گرداند.
func (h *Handler) DoctorPerformancePatientList(c *gin.Context) {
	var q DoctorPerformancePatientListQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		middleware.WriteError(c, err)
		return
	}

	resp, err := h.service.DoctorPerformancePatientList(q)
	if err != nil {
		middleware.WriteError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// RevenueByDoctors گزارش درآمد به تفکیک پزشکان را برمی‌گرداند.
func (h *Handler) RevenueByDoctors(c *gin.Context) {
	var q RevenueByDoctorsQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		middleware.WriteError(c, err)
		return
	}

	resp, err := h.service.RevenueByDoctors(q)
	if err != nil {
		middleware.WriteError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// UserPerformanceByReception گزارش کارکرد کاربران بر اساس پذیرش را برمی‌گرداند.
func (h *Handler) UserPerformanceByReception(c *gin.Context) {
	var q UserPerformanceByReceptionQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		middleware.WriteError(c, err)
		return
	}

	resp, err := h.service.UserPerformanceByReception(q)
	if err != nil {
		middleware.WriteError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// UserPerformanceByTransaction گزارش کارکرد کاربران بر اساس تراکنش مالی را برمی‌گرداند.
func (h *Handler) UserPerformanceByTransaction(c *gin.Context) {
	var q UserPerformanceByTransactionQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		middleware.WriteError(c, err)
		return
	}

	resp, err := h.service.UserPerformanceByTransaction(q)
	if err != nil {
		middleware.WriteError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}
