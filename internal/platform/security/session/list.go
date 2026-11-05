package session

import (
	"context"
	"time"

	"github.com/tpdenta/afta-reception/internal/platform/apperror"
	"github.com/tpdenta/afta-reception/internal/platform/listquery"
	"gorm.io/gorm"
)

// sessionUserJoin نام صاحب نشست را بدون حذف نشست بی‌صاحب می‌آورد.
// اگر ردیف کاربر نباشد نام خالی می‌ماند و خود نشست در فهرست هست.
const sessionUserJoin = "LEFT JOIN Users ON Users.ID = Sessions.PersonnelAccountID"

// sessionSelectSQL ستون‌های پاسخ فهرست نشست است.
const sessionSelectSQL = "" +
	"Sessions.Id AS id, " +
	"Users.Username AS username, " +
	"Sessions.Ip AS ip, " +
	"Sessions.Browser AS browser, " +
	"Sessions.CreationTime AS creation_time, " +
	"Sessions.PersonnelAccountID AS user_id"

// ListedSession یک ردیف فهرست نشست پس از join نام کاربر است.
type ListedSession struct {
	ID           string
	Username     string
	IP           string
	Browser      string
	CreationTime time.Time
	UserID       int
}

type sessionScan struct {
	ID           MSSQLUUID `gorm:"column:id"`
	Username     string    `gorm:"column:username"`
	IP           string    `gorm:"column:ip"`
	Browser      string    `gorm:"column:browser"`
	CreationTime time.Time `gorm:"column:creation_time"`
	UserID       int       `gorm:"column:user_id"`
}

// ListSpec ستون‌های GET /api/sessions را برمی‌گرداند.
// includeUserID برای مدیر یا دارنده users.listSessions ستون user_id را هم مجاز می‌کند.
// بقیه ستون‌ها برای هر کاربر واردشده قابل فیلتر و مرتب‌سازی‌اند. FTA_TAH / R1-05.
func ListSpec(includeUserID bool) listquery.Spec {
	columns := map[string]listquery.Column{
		"username":      {Expr: "Users.Username", Type: listquery.KindText, Sortable: true, Filterable: true},
		"ip":            {Expr: "Sessions.Ip", Type: listquery.KindText, Sortable: true, Filterable: true, Trim: true},
		"browser":       {Expr: "Sessions.Browser", Type: listquery.KindText, Sortable: true, Filterable: true},
		"creation_time": {Expr: "Sessions.CreationTime", Type: listquery.KindDate, Sortable: true, Filterable: true},
	}
	if includeUserID {
		columns["user_id"] = listquery.Column{
			Expr: "Sessions.PersonnelAccountID", Type: listquery.KindNumber, Sortable: true, Filterable: true,
		}
	}
	return listquery.Spec{Columns: columns, DefaultSort: "creation_time", DefaultDesc: true}
}

// Search نشست‌ها را با فیلتر سمت سرور برمی‌گرداند.
// ctx لغو درخواست است. q از ListSpec ساخته شده. actorID درخواست‌کننده است.
// canListAll یعنی فهرست همه کاربران؛ وگرنه شرط ثابت فقط نشست‌های خودش را می‌گذارد.
// خروجی صفحه استاندارد یا خطای خواندن است. حذف نشست از این مسیر انجام نمی‌شود.
func (s *Service) Search(ctx context.Context, q listquery.Query, actorID int, canListAll bool) (listquery.Page[ListedSession], error) {
	if s == nil || s.db == nil {
		return listquery.Page[ListedSession]{}, apperror.ErrInternal
	}
	rows, total, err := scanSessions(ctx, s.db, q, actorID, canListAll)
	if err != nil {
		return listquery.Page[ListedSession]{}, apperror.New("DB_ERROR", "خطا در خواندن نشست‌ها.", err.Error(), 500)
	}
	items := make([]ListedSession, 0, len(rows))
	for _, row := range rows {
		items = append(items, ListedSession{
			ID:           row.ID.String(),
			Username:     row.Username,
			IP:           row.IP,
			Browser:      row.Browser,
			CreationTime: row.CreationTime,
			UserID:       row.UserID,
		})
	}
	return listquery.NewPage(items, total, q), nil
}

// scanSessions شمارش و صفحه نشست را از پایگاه می‌خواند.
// ctx لغو درخواست، db اتصال، و q فیلتر اعتبارسنجی‌شده است.
// actorID و canListAll محدوده مالکیت را جدا از فیلتر کاربر اعمال می‌کنند.
func scanSessions(ctx context.Context, db *gorm.DB, q listquery.Query, actorID int, canListAll bool) ([]sessionScan, int64, error) {
	var total int64
	if err := q.ApplyFilters(sessionBase(ctx, db, actorID, canListAll)).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []sessionScan
	err := q.Apply(sessionBase(ctx, db, actorID, canListAll)).Select(sessionSelectSQL).Scan(&rows).Error
	return rows, total, err
}

// sessionBase کوئری نشست را با نام کاربر می‌سازد.
// اگر canListAll نادرست باشد فقط PersonnelAccountID همان actor در شرط می‌ماند.
// این شرط از ورودی فیلتر نمی‌آید تا کاربر عادی با sort یا filter فهرست دیگران را نبیند.
func sessionBase(ctx context.Context, db *gorm.DB, actorID int, canListAll bool) *gorm.DB {
	if ctx == nil {
		ctx = context.Background()
	}
	q := db.WithContext(ctx).Model(&Session{}).Joins(sessionUserJoin)
	if !canListAll {
		q = q.Where("Sessions.PersonnelAccountID = ?", actorID)
	}
	return q
}
