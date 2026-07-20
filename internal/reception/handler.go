package reception

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/tpdenta/afta-reception/internal/platform/apperror"
	"github.com/tpdenta/afta-reception/internal/platform/middleware"
)

// Handler کنترلر HTTP ماژول پذیرش.
type Handler struct {
	service *Service
}

// NewHandler نمونه Handler می‌سازد.
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// Create ایجاد پذیرش جدید.
func (h *Handler) Create(c *gin.Context) {
	var req UpsertReceptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.WriteError(c, err)
		return
	}
	uid := actorID(c)
	resp, err := h.service.CreateReception(req, uid, c.ClientIP())
	if err != nil {
		middleware.WriteError(c, err)
		return
	}
	c.JSON(http.StatusCreated, resp)
}

// Update ویرایش پذیرش.
func (h *Handler) Update(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var req UpsertReceptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.WriteError(c, err)
		return
	}
	resp, err := h.service.UpdateReception(id, req, actorID(c), c.ClientIP())
	if err != nil {
		middleware.WriteError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// GetByID دریافت پذیرش با شناسه.
func (h *Handler) GetByID(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	resp, err := h.service.GetReception(id)
	if err != nil {
		middleware.WriteError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// List لیست پذیرش‌ها.
func (h *Handler) List(c *gin.Context) {
	list, err := h.service.ListReceptions()
	if err != nil {
		middleware.WriteError(c, err)
		return
	}
	c.JSON(http.StatusOK, list)
}

// Navigate ناوبری بین پذیرش‌ها.
func (h *Handler) Navigate(c *gin.Context) {
	var q NavQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		middleware.WriteError(c, err)
		return
	}
	resp, err := h.service.Navigate(q.Cursor, q.Dir)
	if err != nil {
		middleware.WriteError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// Calculate محاسبه خدمات بدون ذخیره.
func (h *Handler) Calculate(c *gin.Context) {
	var req CalculateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.WriteError(c, err)
		return
	}
	resp, err := h.service.CalculateServicesWithActor(req, actorID(c), c.ClientIP())
	if err != nil {
		middleware.WriteError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// ListByPatient لیست پذیرش‌های یک پرونده.
func (h *Handler) ListByPatient(c *gin.Context) {
	id, ok := parseUintParam(c, "patientId")
	if !ok {
		return
	}
	list, err := h.service.ListByPatient(id)
	if err != nil {
		middleware.WriteError(c, err)
		return
	}
	c.JSON(http.StatusOK, list)
}

// PatientServiceHistory تاریخچه خدمات پرونده.
func (h *Handler) PatientServiceHistory(c *gin.Context) {
	id, ok := parseUintParam(c, "patientId")
	if !ok {
		return
	}
	list, err := h.service.PatientServiceHistory(id)
	if err != nil {
		middleware.WriteError(c, err)
		return
	}
	c.JSON(http.StatusOK, list)
}

// EndReception پایان پذیرش را ثبت می‌کند.
func (h *Handler) EndReception(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	resp, err := h.service.EndReception(id, actorID(c), c.ClientIP())
	if err != nil {
		middleware.WriteError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// UploadPhoto آپلود عکس دندان برای ضوابط.
func (h *Handler) UploadPhoto(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	file, err := c.FormFile("file")
	if err != nil {
		middleware.WriteError(c, apperror.New("BAD_REQUEST", "فایل آپلود نشده است.", err.Error(), 400))
		return
	}
	f, err := file.Open()
	if err != nil {
		middleware.WriteError(c, err)
		return
	}
	defer f.Close()
	resp, err := h.service.UploadReceptionPhoto(id, file.Filename, f, actorID(c), c.ClientIP())
	if err != nil {
		middleware.WriteError(c, err)
		return
	}
	c.JSON(http.StatusCreated, resp)
}

// GetPhoto دانلود/نمایش عکس پذیرش.
func (h *Handler) GetPhoto(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	photoID, ok := parseUintParam(c, "photoId")
	if !ok {
		return
	}
	path, name, err := h.service.GetReceptionPhotoPath(id, photoID)
	if err != nil {
		middleware.WriteError(c, err)
		return
	}
	c.FileAttachment(path, name)
}

// Delete حذف نرم پذیرش.
func (h *Handler) Delete(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	if err := h.service.SoftDelete(id, actorID(c), c.ClientIP()); err != nil {
		middleware.WriteError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// Restore بازیابی پذیرش حذف‌شده.
func (h *Handler) Restore(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	resp, err := h.service.Restore(id, actorID(c), c.ClientIP())
	if err != nil {
		middleware.WriteError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// parseID شناسه مسیر را می‌خواند.
func parseID(c *gin.Context) (uint, bool) {
	return parseUintParam(c, "id")
}

// parseUintParam پارامتر عددی مسیر را می‌خواند.
func parseUintParam(c *gin.Context, name string) (uint, bool) {
	raw := c.Param(name)
	id64, err := strconv.ParseUint(raw, 10, 64)
	if err != nil {
		middleware.WriteError(c, apperror.New("BAD_REQUEST", "شناسه نامعتبر است.", err.Error(), 400))
		return 0, false
	}
	return uint(id64), true
}

// actorID شناسه کاربر فعلی را از کانتکست می‌خواند.
func actorID(c *gin.Context) int {
	actorIDVal, _ := c.Get(middleware.ContextKeyUserID)
	uid, _ := actorIDVal.(int)
	return uid
}
