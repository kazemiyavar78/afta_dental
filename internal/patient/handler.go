package patient

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tpdenta/afta-reception/internal/platform/middleware"
)

// Handler کنترلر HTTP ماژول بیمار.
type Handler struct{ service *Service }

// NewHandler نمونه Handler بیمار می‌سازد.
func NewHandler(s *Service) *Handler { return &Handler{service: s} }

// Create بیمار جدید ایجاد می‌کند.
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

// GetByID بیمار را با شناسه برمی‌گرداند.
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

// List لیست بیماران یا نتیجه جستجو بر اساس query را برمی‌گرداند.
func (h *Handler) List(c *gin.Context) {
	var req SearchRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		middleware.WriteError(c, err)
		return
	}

	hasFilter := req.FirstName != "" || req.LastName != "" || req.NationalCode != "" ||
		req.BirthDate != "" || req.Address != "" || req.HomePhoneNumber != "" ||
		req.MobilePhoneNumber != "" || req.FileNumber != "" || req.Sex != nil

	var (
		list []Response
		err  error
	)
	if hasFilter {
		list, err = h.service.Search(req)
	} else {
		list, err = h.service.List()
	}
	if err != nil {
		middleware.WriteError(c, err)
		return
	}
	c.JSON(http.StatusOK, list)
}

// Update بیمار را بروزرسانی می‌کند.
func (h *Handler) Update(c *gin.Context) {
	var uri struct {
		ID uint `uri:"id" binding:"required"`
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

// Delete بیمار را حذف می‌کند.
func (h *Handler) Delete(c *gin.Context) {
	var uri struct {
		ID uint `uri:"id" binding:"required"`
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
