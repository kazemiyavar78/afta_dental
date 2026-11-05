package listquery

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ApplyFilters فقط شرط‌های فیلتر را به db اضافه می‌کند.
// db زنجیره GORM است که Model و Join را از قبل دارد. خروجی برای Count بدون صفحه‌بندی است.
// مقدارها پارامتر هستند و نام ستون از تعریف سرور آمده است.
func (q Query) ApplyFilters(db *gorm.DB) *gorm.DB {
	if db == nil {
		return nil
	}
	for _, filter := range q.filters {
		db = db.Where(clause.Expr{SQL: filter.sql, Vars: filter.args})
	}
	return db
}

// Apply فیلتر، مرتب‌سازی و صفحه‌بندی را روی db می‌گذارد.
// db همان زنجیره آماده Model است. خروجی برای Find یا Scan است.
// اندازه صفحه حتی اگر بعد از Parse دستکاری شده باشد از سقف ۱۰۰ بیشتر نمی‌شود.
func (q Query) Apply(db *gorm.DB) *gorm.DB {
	db = q.ApplyFilters(db)
	if db == nil {
		return nil
	}
	if q.sortExpr != "" {
		db = db.Order(clause.OrderByColumn{
			Column: clause.Column{Name: q.sortExpr, Raw: true},
			Desc:   q.sortDesc,
		})
	}
	page, size := clampPage(q.Page, q.PageSize)
	return db.Offset((page - 1) * size).Limit(size)
}

// clampPage شماره و اندازه صفحه را به بازه مجاز برمی‌گرداند.
// page و size مقادیر Query هستند. خروجی حداقل صفحه ۱ و اندازه حداکثر ۱۰۰ است.
func clampPage(page, size int) (int, int) {
	if page < 1 {
		page = DefaultPage
	}
	if size < 1 {
		size = DefaultPageSize
	}
	if size > MaxPageSize {
		size = MaxPageSize
	}
	return page, size
}

// Page پاسخ استاندارد یک فهرست صفحه‌بندی‌شده است.
// Items هرگز تهیِ JSON نیست؛ فهرست خالی آرایه خالی است.
type Page[T any] struct {
	Items    []T   `json:"items"`
	Total    int64 `json:"total"`
	Page     int   `json:"page"`
	PageSize int   `json:"page_size"`
}

// NewPage پاسخ استاندارد را از ردیف‌ها و Query می‌سازد.
// items نتیجه همان صفحه است و total تعداد کل بدون صفحه‌بندی است.
// اگر items تهی باشد آرایه خالی برمی‌گردد تا JSON مقدار null ندهد.
func NewPage[T any](items []T, total int64, q Query) Page[T] {
	if items == nil {
		items = []T{}
	}
	page, size := clampPage(q.Page, q.PageSize)
	return Page[T]{Items: items, Total: total, Page: page, PageSize: size}
}
