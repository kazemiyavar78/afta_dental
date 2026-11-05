// این ماژول نمونه است؛ باقی Entity ها و Endpoint ها طبق همین الگو در فازهای بعدی تکمیل می‌شوند.
package fund

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tpdenta/afta-reception/internal/platform/middleware"
)

type Handler struct{ service *Service }

func NewHandler(s *Service) *Handler { return &Handler{service: s} }

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

func (h *Handler) List(c *gin.Context) {
	list, err := h.service.List()
	if err != nil {
		middleware.WriteError(c, err)
		return
	}
	c.JSON(http.StatusOK, list)
}
