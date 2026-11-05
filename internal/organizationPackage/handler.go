package organizationpackage

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tpdenta/afta-reception/internal/platform/middleware"
)

// Handler کنترلر HTTP ماژول بسته‌های تعرفه.
type Handler struct{ service *Service }

// NewHandler نمونه Handler بسته‌های تعرفه می‌سازد.
func NewHandler(s *Service) *Handler { return &Handler{service: s} }

// Create بسته تعرفه جدید ایجاد می‌کند.
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

// GetByID بسته تعرفه را با شناسه برمی‌گرداند.
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

// List لیست بسته‌های تعرفه را برمی‌گرداند.
func (h *Handler) List(c *gin.Context) {
	list, err := h.service.List()
	if err != nil {
		middleware.WriteError(c, err)
		return
	}
	c.JSON(http.StatusOK, list)
}

// Update بسته تعرفه را بروزرسانی می‌کند.
func (h *Handler) Update(c *gin.Context) {
	var uri struct {
		ID int `uri:"id" binding:"required"`
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

// Delete بسته تعرفه را حذف می‌کند.
func (h *Handler) Delete(c *gin.Context) {
	var uri struct {
		ID int `uri:"id" binding:"required"`
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
