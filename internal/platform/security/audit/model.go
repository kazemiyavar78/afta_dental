package audit

import (
	"time"

	"gorm.io/gorm"
)

// SecurityEvent مدل جدول رویدادهای امنیتی با Hash-Chain.
type SecurityEvent struct {
	ID          int64     `gorm:"column:ID;primaryKey;autoIncrement"`
	UserID      *int      `gorm:"column:UserID"`
	Ip          string    `gorm:"column:Ip;size:45;not null"`
	EventType   string    `gorm:"column:EventType;size:100;not null"`
	Description string    `gorm:"column:Description;type:nvarchar(max);not null"`
	CreatedAt   time.Time `gorm:"column:CreatedAt;not null"`
	PrevHash    string    `gorm:"column:PrevHash;size:128;not null"`
	RowHash     string    `gorm:"column:RowHash;size:128;not null"`
}

// TableName نام جدول در دیتابیس.
func (SecurityEvent) TableName() string {
	return "SecurityEvents"
}

// Repository عملیات CRUD خام رویدادهای امنیتی.
type Repository interface {
	GetLastRowHash() (string, error)
	Create(event *SecurityEvent) error
	FindFromID(fromID int64) ([]SecurityEvent, error)
	GetLastEventID() (int64, error)
	Count() (int64, error)
	DeleteOldest(count int) (int64, error)
	GetTableSizeMB() (float64, error)
}

type gormRepository struct {
	db *gorm.DB
}

// NewRepository نمونه Repository می‌سازد.
func NewRepository(db *gorm.DB) Repository {
	return &gormRepository{db: db}
}

func (r *gormRepository) GetLastRowHash() (string, error) {
	var event SecurityEvent
	// Last به‌جای First+Order — جلوگیری از ORDER BY تکراری ID در SQL Server
	err := r.db.Last(&event).Error
	if err == gorm.ErrRecordNotFound {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return event.RowHash, nil
}

func (r *gormRepository) Create(event *SecurityEvent) error {
	return r.db.Create(event).Error
}

func (r *gormRepository) FindFromID(fromID int64) ([]SecurityEvent, error) {
	var events []SecurityEvent
	err := r.db.Where("ID >= ?", fromID).Order("ID ASC").Find(&events).Error
	return events, err
}

func (r *gormRepository) GetLastEventID() (int64, error) {
	var event SecurityEvent
	err := r.db.Last(&event).Error
	if err == gorm.ErrRecordNotFound {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return event.ID, nil
}

func (r *gormRepository) Count() (int64, error) {
	var count int64
	err := r.db.Model(&SecurityEvent{}).Count(&count).Error
	return count, err
}

func (r *gormRepository) DeleteOldest(count int) (int64, error) {
	var ids []int64
	err := r.db.Model(&SecurityEvent{}).Order("ID ASC").Limit(count).Pluck("ID", &ids).Error
	if err != nil || len(ids) == 0 {
		return 0, err
	}
	result := r.db.Where("ID IN ?", ids).Delete(&SecurityEvent{})
	return result.RowsAffected, result.Error
}

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
