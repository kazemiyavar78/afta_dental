package logs

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/tpdenta/afta-reception/internal/platform/apperror"
	"github.com/tpdenta/afta-reception/internal/platform/middleware"
	"github.com/tpdenta/afta-reception/internal/platform/security/audit"
)

// eventFinder خواندن رویدادهای ممیزی است. audit.Repository همین متد را دارد.
type eventFinder interface {
	FindEvents(filter audit.EventFilter) ([]audit.SecurityEvent, error)
}

// Handler فهرست لاگ امنیتی را برمی‌گرداند و خواندن صفحه اول را ثبت می‌کند.
type Handler struct {
	service eventFinder
	audit   audit.Recorder
}

// NewHandler هندلر لاگ را می‌سازد.
// events خواندن ردیف‌ها و rec ثبت AUDIT_READ است. خروجی هندلر آماده ثبت مسیر است.
func NewHandler(events eventFinder, rec audit.Recorder) *Handler {
	return &Handler{service: events, audit: rec}
}

// List همه لاگ‌ها را با فیلتر ip، username، from_date و to_date برای ادمین برمی‌گرداند.
// فقط صفحه اول یک جستجو AUDIT_READ می‌نویسد تا صفحه‌بندی جدول لاگ را پر نکند.
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
	if firstAuditPage(c.Query("page")) {
		if err := h.recordRead(c, filterDetails(req)); err != nil {
			middleware.WriteError(c, err)
			return
		}
	}
	c.JSON(http.StatusOK, list)
}

// GetMyLogs لاگ‌های کاربر جاری را برمی‌گرداند.
// صفحه اول همان جستجو AUDIT_READ با محدوده self می‌نویسد و داده حساس در جزئیات نیست.
func (h *Handler) GetMyLogs(c *gin.Context) {
	uid, _ := c.Get(middleware.ContextKeyUserID)
	userID := uid.(int)

	list, err := h.service.FindEvents(audit.EventFilter{UserID: &userID})
	if err != nil {
		middleware.WriteError(c, err)
		return
	}
	if firstAuditPage(c.Query("page")) {
		if err := h.recordRead(c, map[string]string{"scope": "self"}); err != nil {
			middleware.WriteError(c, err)
			return
		}
	}
	c.JSON(http.StatusOK, list)
}

// recordRead خواندن ممیزی را با فیلترهای غیرحساس ثبت می‌کند.
// c درخواست است. details فیلترها یا محدوده است. خروجی خطای ممیزی یا nil است.
// FAU_GEN.1: خواندن لاگ خودش رویداد ممیزی است.
func (h *Handler) recordRead(c *gin.Context, details map[string]string) error {
	if h.audit == nil {
		return audit.Unavailable(apperror.ErrInternal)
	}
	ctx := context.Background()
	if c.Request != nil {
		ctx = c.Request.Context()
	}
	return audit.Unavailable(h.audit.Record(ctx, audit.Event{
		Type:        audit.EventAuditRead,
		Result:      audit.ResultSuccess,
		Actor:       middleware.ActorFromContext(c),
		Details:     details,
		Description: "خواندن ممیزی",
	}))
}

// firstAuditPage می‌گوید این درخواست صفحه اول جستجو است.
// page پارامتر query است. خالی یا ۱ یعنی صفحه اول. مقدار دیگر صفحه‌بندی است و رویداد نمی‌خواهد.
func firstAuditPage(page string) bool {
	page = strings.TrimSpace(page)
	return page == "" || page == "1"
}

// filterDetails فیلترهای جستجوی لاگ را برای جزئیات ممیزی جمع می‌کند.
// req پارامترهای query است. خروجی فقط همان فیلدهاست و بدنه ردیف‌ها را ندارد.
func filterDetails(req LogFilterQuery) map[string]string {
	details := map[string]string{}
	if req.IP != "" {
		details["ip"] = req.IP
	}
	if req.Username != "" {
		details["username"] = req.Username
	}
	if req.FromDate != nil {
		details["from_date"] = req.FromDate.Format("2006-01-02")
	}
	if req.ToDate != nil {
		details["to_date"] = req.ToDate.Format("2006-01-02")
	}
	return details
}
