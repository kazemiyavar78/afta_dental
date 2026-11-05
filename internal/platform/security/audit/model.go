package audit

import "time"

// SecurityEvent مدل جدول رویدادهای امنیتی با Hash-Chain.
type SecurityEvent struct {
	ID          int64     `gorm:"column:ID;primaryKey;autoIncrement" json:"id"`
	UserID      *int      `gorm:"column:UserID" json:"user_id,omitempty"`
	Ip          string    `gorm:"column:Ip;size:45;not null" json:"ip"`
	EventType   string    `gorm:"column:EventType;size:100;not null" json:"event_type"`
	Description string    `gorm:"column:Description;type:nvarchar(max);not null" json:"description"`
	CreatedAt   time.Time `gorm:"column:CreatedAt;not null" json:"created_at"`
	PrevHash    string    `gorm:"column:PrevHash;size:128;not null" json:"-"`
	RowHash     string    `gorm:"column:RowHash;size:128;not null" json:"-"`
}

// TableName نام جدول SecurityEvents را برمی‌گرداند.
func (SecurityEvent) TableName() string {
	return "SecurityEvents"
}

// EventFilter فیلترهای جستجوی رویدادهای امنیتی است.
type EventFilter struct {
	UserID   *int
	IP       string
	Username string
	FromDate *time.Time
	ToDate   *time.Time
}

// Repository دسترسی به رویدادها و ردیف وضعیت زنجیره است.
// apply در DeleteVerifiedPrefix وضعیت جدید را از روی ردیف‌های در حال حذف می‌سازد.
// اگر nil برگرداند حذفی انجام نمی‌شود. apply نباید LogEvent صدا بزند.
type Repository interface {
	AppendChained(build func(prevHash string) (*SecurityEvent, error)) error
	FindAfterID(afterID int64, limit int) ([]SecurityEvent, error)
	Count() (int64, error)
	GetTableSizeMB() (float64, error)
	FindEvents(filter EventFilter) ([]SecurityEvent, error)
	LoadChainState() (AuditChainState, error)
	SaveChainState(state *AuditChainState) error
	DeleteVerifiedPrefix(maxID int64, limit int, apply func([]SecurityEvent) (*AuditChainState, error)) ([]SecurityEvent, error)
}
