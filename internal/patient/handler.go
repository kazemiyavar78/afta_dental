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

// GetLastFileNumber آخرین شماره پرونده و شماره پیشنهادی بعدی را برمی‌گرداند.
func (h *Handler) GetLastFileNumber(c *gin.Context) {
	resp, err := h.service.GetLastFileNumber()
	if err != nil {
		middleware.WriteError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
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
	if !middleware.BindQuery(c, &req) {
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

// Delete بیمار را حذف می‌کند.
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
