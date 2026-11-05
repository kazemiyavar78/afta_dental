package audit

import (
	"context"
	"time"

	"gorm.io/gorm"
)

// gormFindingStore یافته‌های یکپارچگی را در IntegrityFindings نگه می‌دارد.
type gormFindingStore struct {
	db *gorm.DB
}

// NewFindingStore مخزن GORM یافته‌ها را می‌سازد.
// db اتصال باز است. خروجی برای worker و خواندن یافته‌های باز صفحه لاگ است.
func NewFindingStore(db *gorm.DB) FindingStore {
	return &gormFindingStore{db: db}
}

// ListOpen یافته‌های باز یک نوع را به ترتیب شناسه برمی‌گرداند.
// entityType همان نوع اسکن است. خروجی ردیف‌هایی است که ResolvedAt تهی دارند.
func (s *gormFindingStore) ListOpen(ctx context.Context, entityType string) ([]IntegrityFinding, error) {
	ctx = findingContext(ctx)
	var rows []IntegrityFinding
	err := s.db.WithContext(ctx).
		Where("EntityType = ? AND ResolvedAt IS NULL", entityType).
		Order("ID ASC").
		Find(&rows).Error
	return rows, err
}

// ListAllOpen همه یافته‌های باز را برای نمایش به مدیر برمی‌گرداند.
// ترتیب بر اساس اولین مشاهده است. خروجی خالی یعنی ردیف باز نیست.
func (s *gormFindingStore) ListAllOpen(ctx context.Context) ([]IntegrityFinding, error) {
	ctx = findingContext(ctx)
	var rows []IntegrityFinding
	err := s.db.WithContext(ctx).
		Where("ResolvedAt IS NULL").
		Order("FirstSeenAt ASC, ID ASC").
		Find(&rows).Error
	return rows, err
}

// Insert یک یافته تازه را ذخیره می‌کند و شناسه را در row می‌نویسد.
// row فیلدهای ردیف نامعتبر است. خطا یعنی درج انجام نشده است.
func (s *gormFindingStore) Insert(ctx context.Context, row *IntegrityFinding) error {
	return s.db.WithContext(findingContext(ctx)).Create(row).Error
}

// Touch فقط LastSeenAt یک یافته باز را جلو می‌برد.
// id شناسه یافته و lastSeen زمان این دور است. نام و اولین مشاهده عوض نمی‌شوند.
func (s *gormFindingStore) Touch(ctx context.Context, id int, lastSeen time.Time) error {
	return s.db.WithContext(findingContext(ctx)).
		Model(&IntegrityFinding{}).
		Where("ID = ?", id).
		Update("LastSeenAt", lastSeen).Error
}

// Resolve زمان بسته شدن یک یافته را می‌نویسد.
// id شناسه یافته و at زمان سالم شدن ردیف است.
func (s *gormFindingStore) Resolve(ctx context.Context, id int, at time.Time) error {
	return s.db.WithContext(findingContext(ctx)).
		Model(&IntegrityFinding{}).
		Where("ID = ?", id).
		Update("ResolvedAt", at).Error
}

// Delete یافته‌ای را که رویدادش ثبت نشده پاک می‌کند.
// id شناسه همان درج ناموفق است. دور بعد می‌تواند هشدار را دوباره بسازد.
func (s *gormFindingStore) Delete(ctx context.Context, id int) error {
	return s.db.WithContext(findingContext(ctx)).Delete(&IntegrityFinding{}, id).Error
}

// Reopen زمان بسته شدن را تهی می‌کند تا رفعِ ثبت‌نشده در دور بعد دوباره تلاش شود.
// id شناسه یافته است. مقدار NULL با map نوشته می‌شود تا GORM ستون را جا نیندازد.
func (s *gormFindingStore) Reopen(ctx context.Context, id int) error {
	return s.db.WithContext(findingContext(ctx)).
		Model(&IntegrityFinding{}).
		Where("ID = ?", id).
		Updates(map[string]any{"ResolvedAt": nil}).Error
}

// findingContext زمینه تهی را به زمینه پس‌زمینه تبدیل می‌کند.
// ctx می‌تواند nil باشد. خروجی برای WithContext امن است.
func findingContext(ctx context.Context) context.Context {
	if ctx == nil {
		return context.Background()
	}
	return ctx
}
