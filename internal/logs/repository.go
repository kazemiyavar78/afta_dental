package logs

import (
	"context"

	"github.com/tpdenta/afta-reception/internal/platform/apperror"
	"github.com/tpdenta/afta-reception/internal/platform/listquery"
	"github.com/tpdenta/afta-reception/internal/platform/security/audit"
	"gorm.io/gorm"
)

// eventRepo خواندن صفحه‌بندی‌شده SecurityEvents است.
type eventRepo struct {
	db *gorm.DB
}

// base کوئری مشترک شمارش و صفحه را با LEFT JOIN کاربر می‌سازد.
// ctx لغو درخواست است. onlyUserID اگر پر باشد فهرست را به همان کاربر محدود می‌کند.
// خروجی زنجیره GORM بدون فیلتر و بدون صفحه‌بندی است.
func (r eventRepo) base(ctx context.Context, onlyUserID *int) *gorm.DB {
	if ctx == nil {
		ctx = context.Background()
	}
	q := r.db.WithContext(ctx).Model(&audit.SecurityEvent{}).Joins(userJoin)
	if onlyUserID != nil {
		q = q.Where("SecurityEvents.UserID = ?", *onlyUserID)
	}
	return q
}

// list صفحه رویدادها و تعداد کل را برمی‌گرداند.
// ctx لغو درخواست، q فیلتر اعتبارسنجی‌شده و onlyUserID محدوده لاگ خود کاربر است.
// خروجی ردیف‌های همان صفحه و total بدون صفحه‌بندی است. خطا یعنی خواندن انجام نشده.
func (r eventRepo) list(ctx context.Context, q listquery.Query, onlyUserID *int) ([]eventRow, int64, error) {
	if r.db == nil {
		return nil, 0, apperror.ErrInternal
	}
	var total int64
	if err := q.ApplyFilters(r.base(ctx, onlyUserID)).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []eventRow
	err := q.Apply(r.base(ctx, onlyUserID)).Select(eventSelectSQL).Scan(&rows).Error
	if err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}
