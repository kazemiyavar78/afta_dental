package services

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tpdenta/afta-reception/internal/platform/middleware"
)

// Handler کنترلر HTTP ماژول خدمات.
type Handler struct{ service *Service }

// NewHandler نمونه Handler خدمات می‌سازد.
func NewHandler(s *Service) *Handler { return &Handler{service: s} }

// Create خدمت جدید ایجاد می‌کند.
func (h *Handler) Create(c *gin.Context) {
	var req CreateRequest
	if !middleware.BindJSON(c, &req) {
		return
	}
	resp, err := h.service.Create(c.Request.Context(), req, middleware.ActorFromContext(c))
	if err != nil {
		middleware.WriteError(c, err)
		return
	}
	c.JSON(http.StatusCreated, resp)
}

// GetByID خدمت را با شناسه برمی‌گرداند.
func (h *Handler) GetByID(c *gin.Context) {
	var uri struct {
		ID uint `uri:"id" binding:"required"`
	}
	if err := c.ShouldBindUri(&uri); err != nil {
		middleware.WriteError(c, err)
		return
	}
	resp, err := h.service.Get(uri.ID)
	if err != nil {
		middleware.WriteError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// List لیست خدمات را برمی‌گرداند.
func (h *Handler) List(c *gin.Context) {
	list, err := h.service.List()
	if err != nil {
		middleware.WriteError(c, err)
		return
	}
	c.JSON(http.StatusOK, list)
}

// Update خدمت را بروزرسانی می‌کند.
func (h *Handler) Update(c *gin.Context) {
	var uri struct {
		ID uint `uri:"id" binding:"required"`
	}
	if err := c.ShouldBindUri(&uri); err != nil {
		middleware.WriteError(c, err)
		return
	}
	var req UpdateRequest
	if !middleware.BindJSON(c, &req) {
		return
	}
	resp, err := h.service.Update(c.Request.Context(), uri.ID, req, middleware.ActorFromContext(c))
	if err != nil {
		middleware.WriteError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// Delete خدمت را حذف می‌کند.
func (h *Handler) Delete(c *gin.Context) {
	var uri struct {
		ID uint `uri:"id" binding:"required"`
	}
	if err := c.ShouldBindUri(&uri); err != nil {
		middleware.WriteError(c, err)
		return
	}
	if err := h.service.Delete(c.Request.Context(), uri.ID, middleware.ActorFromContext(c)); err != nil {
		middleware.WriteError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// ListExcludedServices خدمات خارج‌شده یک سازمان را برمی‌گرداند.
func (h *Handler) ListExcludedServices(c *gin.Context) {
	var uri struct {
		OrganizationID uint `uri:"organizationId" binding:"required"`
	}
	if err := c.ShouldBindUri(&uri); err != nil {
		middleware.WriteError(c, err)
		return
	}
	list, err := h.service.ListExcludedServicesByOrganization(uri.OrganizationID)
	if err != nil {
		middleware.WriteError(c, err)
		return
	}
	c.JSON(http.StatusOK, list)
}

// AddExcludedServices خدمات را به لیست خارج‌شده سازمان اضافه می‌کند.
func (h *Handler) AddExcludedServices(c *gin.Context) {
	var req ExcludedServicesRequest
	if !middleware.BindJSON(c, &req) {
		return
	}
	resp, err := h.service.AddExcludedServices(c.Request.Context(), req, middleware.ActorFromContext(c))
	if err != nil {
		middleware.WriteError(c, err)
		return
	}
	c.JSON(http.StatusCreated, resp)
}

// RemoveExcludedServices خدمات را از لیست خارج‌شده سازمان حذف می‌کند.
func (h *Handler) RemoveExcludedServices(c *gin.Context) {
	var req ExcludedServicesRequest
	if !middleware.BindJSON(c, &req) {
		return
	}
	if err := h.service.RemoveExcludedServices(c.Request.Context(), req, middleware.ActorFromContext(c)); err != nil {
		middleware.WriteError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
