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
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.WriteError(c, err)
		return
	}
	uid, _ := c.Get(middleware.ContextKeyUserID)
	resp, err := h.service.Create(req, uid.(int), c.ClientIP())
	if err != nil {
		middleware.WriteError(c, err)
		return
	}
	c.JSON(http.StatusCreated, resp)
}

// GetByID خدمت را با شناسه برمی‌گرداند.
func (h *Handler) GetByID(c *gin.Context) {
	var uri struct {
		ID int `uri:"id" binding:"required"`
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
		ID int `uri:"id" binding:"required"`
	}
	if err := c.ShouldBindUri(&uri); err != nil {
		middleware.WriteError(c, err)
		return
	}
	var req UpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.WriteError(c, err)
		return
	}
	uid, _ := c.Get(middleware.ContextKeyUserID)
	resp, err := h.service.Update(uri.ID, req, uid.(int), c.ClientIP())
	if err != nil {
		middleware.WriteError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// Delete خدمت را حذف می‌کند.
func (h *Handler) Delete(c *gin.Context) {
	var uri struct {
		ID int `uri:"id" binding:"required"`
	}
	if err := c.ShouldBindUri(&uri); err != nil {
		middleware.WriteError(c, err)
		return
	}
	uid, _ := c.Get(middleware.ContextKeyUserID)
	if err := h.service.Delete(uri.ID, uid.(int), c.ClientIP()); err != nil {
		middleware.WriteError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
