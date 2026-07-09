// این ماژول نمونه است؛ باقی Entity ها و Endpoint ها طبق همین الگو در فازهای بعدی تکمیل می‌شوند.
package reception

import (
	"net/http"

	"github.com/gin-gonic/gin"
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
	var req CreateReceptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.WriteError(c, err)
		return
	}

	actorID, _ := c.Get(middleware.ContextKeyUserID)
	uid, _ := actorID.(int)

	resp, err := h.service.CreateReception(req, uid, c.ClientIP())
	if err != nil {
		middleware.WriteError(c, err)
		return
	}

	c.JSON(http.StatusCreated, resp)
}

// GetByID دریافت پذیرش با شناسه.
func (h *Handler) GetByID(c *gin.Context) {
	var uri struct {
		ID int `uri:"id" binding:"required"`
	}
	if err := c.ShouldBindUri(&uri); err != nil {
		middleware.WriteError(c, err)
		return
	}

	resp, err := h.service.GetReception(uri.ID)
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
