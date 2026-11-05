package logs

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/tpdenta/afta-reception/internal/platform/apperror"
	"github.com/tpdenta/afta-reception/internal/platform/listquery"
	"github.com/tpdenta/afta-reception/internal/platform/middleware"
	"github.com/tpdenta/afta-reception/internal/platform/security/audit"
)

// logAPI خواندن فهرست، فراداده و وضعیت زنجیره است. *Service همین متدها را دارد.
type logAPI interface {
	List(ctx context.Context, q listquery.Query, onlyUserID *int) (listquery.Page[EventItem], error)
	Meta(ctx context.Context) (MetaResponse, error)
	Integrity(ctx context.Context) (IntegrityResponse, error)
}

// Handler فهرست لاگ را برمی‌گرداند و خواندن صفحه اول را ثبت می‌کند.
type Handler struct {
	api   logAPI
	audit audit.Recorder
}

// NewHandler هندلر لاگ را می‌سازد.
// api خواندن ردیف‌ها و rec ثبت AUDIT_READ است. خروجی هندلر آماده ثبت مسیر است.
func NewHandler(api logAPI, rec audit.Recorder) *Handler {
	return &Handler{api: api, audit: rec}
}

// List لاگ‌ها را با فیلتر و مرتب‌سازی سمت سرور برمی‌گرداند.
// فقط صفحه اول یک جستجو AUDIT_READ می‌نویسد تا ورق زدن جدول لاگ را پر نکند.
// FAU_SAR.3: ستون ناشناس پیش از خواندن پایگاه رد می‌شود.
func (h *Handler) List(c *gin.Context) {
	q, err := listquery.Parse(c.Request.URL.Query(), logSpec())
	if err != nil {
		middleware.WriteError(c, err)
		return
	}
	page, err := h.api.List(c.Request.Context(), q, nil)
	if err != nil {
		middleware.WriteError(c, err)
		return
	}
	if firstAuditPage(c.Query("page")) {
		if err := h.recordRead(c, q.FilterSummary()); err != nil {
			middleware.WriteError(c, err)
			return
		}
	}
	c.JSON(http.StatusOK, page)
}

// GetMyLogs لاگ‌های کاربر جاری را صفحه‌بندی می‌کند.
// صفحه اول AUDIT_READ با محدوده self می‌نویسد. فیلترها فقط روی ردیف‌های همان کاربر اعمال می‌شوند.
func (h *Handler) GetMyLogs(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		middleware.WriteError(c, apperror.ErrUnauthorized)
		return
	}
	q, err := listquery.Parse(c.Request.URL.Query(), logSpec())
	if err != nil {
		middleware.WriteError(c, err)
		return
	}
	page, err := h.api.List(c.Request.Context(), q, &userID)
	if err != nil {
		middleware.WriteError(c, err)
		return
	}
	if firstAuditPage(c.Query("page")) {
		details := q.FilterSummary()
		details["scope"] = "self"
		if err := h.recordRead(c, details); err != nil {
			middleware.WriteError(c, err)
			return
		}
	}
	c.JSON(http.StatusOK, page)
}

// Meta مقدارهای فیلتر را با برچسب فارسی برمی‌گرداند.
// این مسیر خواندن ردیف لاگ نیست و AUDIT_READ جدا نمی‌نویسد.
func (h *Handler) Meta(c *gin.Context) {
	meta, err := h.api.Meta(c.Request.Context())
	if err != nil {
		middleware.WriteError(c, err)
		return
	}
	c.JSON(http.StatusOK, meta)
}

// Integrity نتیجه آخرین بررسی زنجیره و یافته‌های باز را برمی‌گرداند.
// مجوز مسیر logs.read است. FAU_STG.2: نمایش نتیجه بررسی و تعداد یافته، نه خود هش.
func (h *Handler) Integrity(c *gin.Context) {
	status, err := h.api.Integrity(c.Request.Context())
	if err != nil {
		middleware.WriteError(c, err)
		return
	}
	c.JSON(http.StatusOK, status)
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

// currentUserID شناسه کاربر نشست را از context می‌خواند.
// c درخواست است. ok نادرست یعنی نشست کاربر ندارد و باید ۴۰۱ برگردد.
func currentUserID(c *gin.Context) (int, bool) {
	raw, ok := c.Get(middleware.ContextKeyUserID)
	if !ok {
		return 0, false
	}
	id, ok := raw.(int)
	return id, ok
}
