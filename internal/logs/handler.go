package logs

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tpdenta/afta-reception/internal/platform/middleware"
	"github.com/tpdenta/afta-reception/internal/platform/security/audit"
)

type Handler struct{ service audit.Repository }

func NewHandler(s audit.Repository) *Handler { return &Handler{service: s} }

// List همه لاگ‌ها را با فیلتر ip، username، from_date و to_date برای ادمین برمی‌گرداند.
func (h *Handler) List(c *gin.Context) {
	var req LogFilterQuery
	if !middleware.BindQuery(c, &req) {
		return
	}

	list, err := h.service.FindEvents(audit.EventFilter{
		IP:       req.IP,
		Username: req.Username,
		FromDate: req.FromDate,
		ToDate:   req.ToDate,
	})
	if err != nil {
		middleware.WriteError(c, err)
		return
	}
	c.JSON(http.StatusOK, list)
}

// GetMyLogs لاگ‌های کاربر جاری را برمی‌گرداند.
func (h *Handler) GetMyLogs(c *gin.Context) {
	uid, _ := c.Get(middleware.ContextKeyUserID)
	userID := uid.(int)

	list, err := h.service.FindEvents(audit.EventFilter{UserID: &userID})
	if err != nil {
		middleware.WriteError(c, err)
		return
	}
	c.JSON(http.StatusOK, list)
}
