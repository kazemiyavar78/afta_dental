package user

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/tpdenta/afta-reception/internal/platform/apperror"
	"github.com/tpdenta/afta-reception/internal/platform/listquery"
	"github.com/tpdenta/afta-reception/internal/platform/middleware"
	"github.com/tpdenta/afta-reception/internal/platform/security/audit"
	"github.com/tpdenta/afta-reception/internal/platform/security/session"
)

// sessionDirectory خواندن صفحه‌بندی‌شده نشست و حذف آن است.
// *session.Service همین دو متد را دارد تا هندلر به مخزن گره نخورد.
type sessionDirectory interface {
	Search(ctx context.Context, q listquery.Query, actorID int, canListAll bool) (listquery.Page[session.ListedSession], error)
	DeleteSession(ctx context.Context, id uuid.UUID, actor audit.Actor, canManageOthers bool) error
}

// ListSessions نشست‌ها را با فیلتر و مرتب‌سازی سمت سرور برمی‌گرداند.
// کاربر عادی فقط نشست خودش را می‌بیند. مدیر یا users.listSessions ستون user_id را هم دارد.
// حذف نشست در همین مسیر نیست. FTA_TAH / R1-05.
func (h *Handler) ListSessions(c *gin.Context) {
	if h == nil || h.sessionSvc == nil {
		middleware.WriteError(c, apperror.ErrInternal)
		return
	}
	actorID, ok := currentUserID(c)
	if !ok {
		middleware.WriteError(c, apperror.ErrUnauthorized)
		return
	}
	canListAll := middleware.IsAdmin(c) || HasPermission(middleware.GetPermissions(c), "users.listSessions")
	q, err := listquery.Parse(c.Request.URL.Query(), session.ListSpec(canListAll))
	if err != nil {
		middleware.WriteError(c, err)
		return
	}
	page, err := h.sessionSvc.Search(c.Request.Context(), q, actorID, canListAll)
	if err != nil {
		middleware.WriteError(c, err)
		return
	}
	c.JSON(http.StatusOK, sessionPage(page))
}

// currentUserID شناسه کاربر نشست را از context می‌خواند.
// c درخواست است. ok نادرست یعنی شناسه عددی در context نیست.
func currentUserID(c *gin.Context) (int, bool) {
	raw, ok := c.Get(middleware.ContextKeyUserID)
	if !ok {
		return 0, false
	}
	id, ok := raw.(int)
	return id, ok
}

// sessionPage ردیف‌های سرویس را به پاسخ API و قالب زمان قبلی تبدیل می‌کند.
// page نتیجه Search است. خروجی همان صفحه‌بندی با DTO نشست است.
func sessionPage(page listquery.Page[session.ListedSession]) listquery.Page[SessionResponse] {
	items := make([]SessionResponse, 0, len(page.Items))
	for _, item := range page.Items {
		items = append(items, SessionResponse{
			ID:           item.ID,
			Username:     item.Username,
			Ip:           item.IP,
			Browser:      item.Browser,
			CreationTime: item.CreationTime.UTC().Format("2006-01-02T15:04:05Z"),
			UserID:       item.UserID,
		})
	}
	return listquery.Page[SessionResponse]{
		Items:    items,
		Total:    page.Total,
		Page:     page.Page,
		PageSize: page.PageSize,
	}
}
