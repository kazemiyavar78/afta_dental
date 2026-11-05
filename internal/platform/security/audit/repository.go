package audit

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
)

// errNilAuditTx یعنی RecordTx بدون تراکنش باز صدا زده شده است.
var errNilAuditTx = errors.New("تراکنش ممیزی خالی است")

// errNilChainState یعنی ذخیره‌سازی وضعیت بدون ردیف صدا زده شده است.
var errNilChainState = errors.New("وضعیت زنجیره خالی است")

type gormRepository struct {
	db *gorm.DB
}

// NewRepository مخزن GORM رویدادهای امنیتی را می‌سازد.
// db اتصال باز است. خروجی همان Repository است که Manager و API لاگ استفاده می‌کنند.
func NewRepository(db *gorm.DB) Repository {
	return &gormRepository{db: db}
}

// AppendChained آخرین هش را زیر قفل ردیف می‌خواند و رویداد را در همان تراکنش درج می‌کند.
// build با PrevHash قطعی صدا زده می‌شود و ردیف آماده درج را برمی‌گرداند.
// اگر ردیفی نمانده باشد و لنگر پاک‌سازی وجود داشته باشد، همان لنگر هش قبلی است.
func (r *gormRepository) AppendChained(build func(prevHash string) (*SecurityEvent, error)) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		return insertChained(tx, build)
	})
}

// AppendChainedCtx مانند AppendChained است ولی context درخواست را به تراکنش می‌دهد.
// ctx لغو یا مهلت درخواست است. build ردیف را از PrevHash قطعی می‌سازد.
// اگر ctx تمام شده باشد درج شروع نمی‌شود.
func (r *gormRepository) AppendChainedCtx(ctx context.Context, build func(prevHash string) (*SecurityEvent, error)) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return insertChained(tx, build)
	})
}

// AppendInTx رویداد را در تراکنش داده‌شده درج می‌کند و تراکنش تازه باز نمی‌کند.
// tx تراکنش عملیات تغییردهنده است. build ردیف آماده را از PrevHash می‌سازد.
// tx تهی خطا می‌دهد تا ثبت بیرون از تراکنش، خاموش و بدون زنجیره انجام نشود.
// FAU_STG.4: شکست درج به صدازننده برمی‌گردد تا کل عملیات برگردد.
func (r *gormRepository) AppendInTx(tx *gorm.DB, build func(prevHash string) (*SecurityEvent, error)) error {
	if tx == nil {
		return errNilAuditTx
	}
	return insertChained(tx, build)
}

// insertChained آخرین هش را قفل می‌کند و ردیف را در همان اتصال درج می‌کند.
// tx تراکنش باز است. build نباید خودش رویداد دیگری ثبت کند.
func insertChained(tx *gorm.DB, build func(prevHash string) (*SecurityEvent, error)) error {
	prevHash, err := lockedPrevHash(tx)
	if err != nil {
		return err
	}
	event, err := build(prevHash)
	if err != nil {
		return err
	}
	return tx.Create(event).Error
}

// lockedPrevHash ردیف وضعیت و سپس آخرین رویداد را با UPDLOCK, HOLDLOCK قفل می‌کند.
// tx تراکنش باز است. خروجی هش ردیف آخر، یا لنگر اگر جدولی خالی باشد، یا رشته خالی در شروع زنجیره است.
// FAU_STG.1: ترتیب قفل (اول وضعیت، بعد رویداد) با پاک‌سازی یکی است تا دو نمونه به بن‌بست نرسند.
func lockedPrevHash(tx *gorm.DB) (string, error) {
	var anchor struct {
		AnchorHash string `gorm:"column:AnchorHash"`
	}
	err := tx.Raw(`
		SELECT TOP (1) AnchorHash
		FROM AuditChainStates WITH (UPDLOCK, HOLDLOCK)
		WHERE ID = ?
	`, chainStateRowID).Scan(&anchor).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return "", err
	}

	var tail struct {
		ID      int64  `gorm:"column:ID"`
		RowHash string `gorm:"column:RowHash"`
	}
	err = tx.Raw(`
		SELECT TOP (1) ID, RowHash
		FROM SecurityEvents WITH (UPDLOCK, HOLDLOCK)
		ORDER BY ID DESC
	`).Scan(&tail).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return "", err
	}
	if tail.ID != 0 {
		return tail.RowHash, nil
	}
	return anchor.AnchorHash, nil
}

// FindAfterID رویدادهای با ID بزرگ‌تر از afterID را به ترتیب و حداکثر limit ردیف برمی‌گرداند.
// afterID آخرین شناسه دیده‌شده است. limit سقف همان دسته است. خطا خطای خواندن پایگاه است.
func (r *gormRepository) FindAfterID(afterID int64, limit int) ([]SecurityEvent, error) {
	var events []SecurityEvent
	err := r.db.Where("ID > ?", afterID).Order("ID ASC").Limit(limit).Find(&events).Error
	return events, err
}

// Count تعداد ردیف‌های SecurityEvents را برمی‌گرداند.
func (r *gormRepository) Count() (int64, error) {
	var count int64
	err := r.db.Model(&SecurityEvent{}).Count(&count).Error
	return count, err
}

// GetTableSizeMB حجم رزروشده جدول SecurityEvents را به مگابایت برمی‌گرداند.
// خروجی از sys.dm_db_partition_stats است. اگر آمار در دسترس نباشد خطا برمی‌گردد.
func (r *gormRepository) GetTableSizeMB() (float64, error) {
	var result struct {
		ReservedKB float64 `gorm:"column:reserved_kb"`
	}
	err := r.db.Raw(`
		SELECT SUM(reserved_page_count) * 8.0 / 1024 AS reserved_kb
		FROM sys.dm_db_partition_stats
		WHERE object_id = OBJECT_ID('SecurityEvents')
	`).Scan(&result).Error
	if err != nil {
		return 0, err
	}
	return result.ReservedKB / 1024.0, nil
}

// FindEvents رویدادها را با فیلتر اختیاری و ترتیب شناسه نزولی برمی‌گرداند.
// filter کاربر، IP، نام کاربری و بازه تاریخ است. خروجی همان ردیف‌های جدول است.
func (r *gormRepository) FindEvents(filter EventFilter) ([]SecurityEvent, error) {
	var events []SecurityEvent
	q := r.db.Model(&SecurityEvent{})
	if filter.UserID != nil {
		q = q.Where("SecurityEvents.UserID = ?", *filter.UserID)
	}
	if filter.IP != "" {
		q = q.Where("SecurityEvents.Ip = ?", filter.IP)
	}
	if filter.Username != "" {
		q = q.Joins("INNER JOIN Users ON Users.ID = SecurityEvents.UserID").
			Where("Users.Username = ?", filter.Username)
	}
	if filter.FromDate != nil && !filter.FromDate.IsZero() {
		q = q.Where("SecurityEvents.CreatedAt >= ?", *filter.FromDate)
	}
	if filter.ToDate != nil && !filter.ToDate.IsZero() {
		to := endOfDay(*filter.ToDate)
		q = q.Where("SecurityEvents.CreatedAt <= ?", to)
	}
	err := q.Order("SecurityEvents.ID DESC").Find(&events).Error
	return events, err
}

// endOfDay پایان روز را برای فیلتر toDate برمی‌گرداند.
// t تاریخ ورودی است. اگر ساعت صفر باشد تا آخرین لحظه همان روز جلو می‌رود.
func endOfDay(t time.Time) time.Time {
	if t.Hour() == 0 && t.Minute() == 0 && t.Second() == 0 && t.Nanosecond() == 0 {
		return t.Add(24*time.Hour - time.Nanosecond)
	}
	return t
}

// LoadChainState ردیف ثابت وضعیت را می‌خواند.
// اگر ردیفی نباشد مقدار صفر با ID ثابت برمی‌گردد تا بررسی از ابتدای زنجیره شروع شود.
func (r *gormRepository) LoadChainState() (AuditChainState, error) {
	var state AuditChainState
	err := r.db.First(&state, chainStateRowID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return AuditChainState{ID: chainStateRowID}, nil
	}
	if err != nil {
		return AuditChainState{}, err
	}
	state.normalize()
	return state, nil
}

// SaveChainState ردیف وضعیت را درج یا به‌روز می‌کند.
// state فیلدهای جدید است. ID همیشه ردیف ثابت است. خطا یعنی ذخیره انجام نشده.
func (r *gormRepository) SaveChainState(state *AuditChainState) error {
	if state == nil {
		return errNilChainState
	}
	return r.db.Transaction(func(tx *gorm.DB) error {
		return saveChainStateTx(tx, state)
	})
}

// DeleteVerifiedPrefix ردیف‌های ID<=maxID را تا limit حذف می‌کند و وضعیت را در همان تراکنش ذخیره می‌کند.
// maxID آخرین ردیف تأییدشده است. limit سقف حذف است. apply ردیف‌های انتخاب‌شده را می‌گیرد و وضعیت جدید را برمی‌گرداند.
// اگر apply مقدار nil بدهد حذفی انجام نمی‌شود. خروجی ردیف‌های حذف‌شده به ترتیب ID است.
// FAU_STG.1: حذف و لنگر با هم commit می‌شوند تا درج همزمان هش قبلی را از دست ندهد.
func (r *gormRepository) DeleteVerifiedPrefix(maxID int64, limit int, apply func([]SecurityEvent) (*AuditChainState, error)) ([]SecurityEvent, error) {
	var deleted []SecurityEvent
	err := r.db.Transaction(func(tx *gorm.DB) error {
		if _, err := lockedPrevHash(tx); err != nil {
			return err
		}
		q := tx.Where("ID <= ?", maxID).Order("ID ASC")
		if limit > 0 {
			q = q.Limit(limit)
		}
		if err := q.Find(&deleted).Error; err != nil {
			return err
		}
		if len(deleted) == 0 {
			return nil
		}
		updated, err := apply(deleted)
		if err != nil || updated == nil {
			return err
		}
		ids := make([]int64, len(deleted))
		for i, row := range deleted {
			ids[i] = row.ID
		}
		if err := tx.Where("ID IN ?", ids).Delete(&SecurityEvent{}).Error; err != nil {
			return err
		}
		return saveChainStateTx(tx, updated)
	})
	if err != nil {
		return nil, err
	}
	return deleted, nil
}

// saveChainStateTx ردیف ثابت را در تراکنش داده‌شده درج یا جایگزین می‌کند.
// tx تراکنش باز است و state باید HMAC نهایی را داشته باشد.
func saveChainStateTx(tx *gorm.DB, state *AuditChainState) error {
	state.ID = chainStateRowID
	state.normalize()
	var existing int64
	if err := tx.Model(&AuditChainState{}).Where("ID = ?", chainStateRowID).Count(&existing).Error; err != nil {
		return err
	}
	if existing == 0 {
		return tx.Create(state).Error
	}
	return tx.Save(state).Error
}
