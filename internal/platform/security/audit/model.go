package audit

import (
	"context"
	"time"

	"gorm.io/gorm"
)

// SecurityEvent مدل جدول رویدادهای امنیتی با Hash-Chain.
// UserID شناسه فاعل است. ActorUsername نام همان کاربر در لحظه رویداد است.
// ردیف نسخه ۱ این نام را ندارد؛ فهرست لاگ برای نمایش و فیلتر، آن را با LEFT JOIN به Users پر می‌کند.
// ستون‌های جدید تهی‌پذیرند تا ردیف قدیمی بماند. HashVersion پیش‌فرض ۱ است.
// FAU_GEN.2: فاعل، هدف، نتیجه، دلیل، IP و User-Agent در خود ردیف هستند.
type SecurityEvent struct {
	ID            int64     `gorm:"column:ID;primaryKey;autoIncrement" json:"id"`
	UserID        *int      `gorm:"column:UserID" json:"user_id,omitempty"`
	ActorUsername string    `gorm:"column:ActorUsername;type:nvarchar(100)" json:"actor_username,omitempty"`
	TargetType    string    `gorm:"column:TargetType;type:nvarchar(50)" json:"target_type,omitempty"`
	TargetID      string    `gorm:"column:TargetID;type:nvarchar(64)" json:"target_id,omitempty"`
	TargetName    string    `gorm:"column:TargetName;type:nvarchar(200)" json:"target_name,omitempty"`
	Ip            string    `gorm:"column:Ip;size:45;not null" json:"ip"`
	UserAgent     string    `gorm:"column:UserAgent;type:nvarchar(512)" json:"user_agent,omitempty"`
	EventType     string    `gorm:"column:EventType;size:100;not null" json:"event_type"`
	Result        string    `gorm:"column:Result;type:nvarchar(20)" json:"result,omitempty"`
	Reason        string    `gorm:"column:Reason;type:nvarchar(50)" json:"reason,omitempty"`
	Description   string    `gorm:"column:Description;type:nvarchar(max);not null" json:"description"`
	Details       string    `gorm:"column:Details;type:nvarchar(max)" json:"details,omitempty"`
	CreatedAt     time.Time `gorm:"column:CreatedAt;not null" json:"created_at"`
	PrevHash      string    `gorm:"column:PrevHash;size:128;not null" json:"-"`
	RowHash       string    `gorm:"column:RowHash;size:128;not null" json:"-"`
	HashVersion   uint8     `gorm:"column:HashVersion;type:tinyint;not null;default:1" json:"-"`
}

// TableName نام جدول SecurityEvents را برمی‌گرداند.
func (SecurityEvent) TableName() string {
	return "SecurityEvents"
}

// Repository دسترسی به رویدادها و ردیف وضعیت زنجیره است.
// apply در DeleteVerifiedPrefix وضعیت جدید را از روی ردیف‌های در حال حذف می‌سازد.
// اگر nil برگرداند حذفی انجام نمی‌شود. apply نباید رویداد تازه بسازد.
type Repository interface {
	AppendChained(build func(prevHash string) (*SecurityEvent, error)) error
	AppendChainedCtx(ctx context.Context, build func(prevHash string) (*SecurityEvent, error)) error
	AppendInTx(tx *gorm.DB, build func(prevHash string) (*SecurityEvent, error)) error
	FindAfterID(afterID int64, limit int) ([]SecurityEvent, error)
	Count() (int64, error)
	GetTableSizeMB() (float64, error)
	LoadChainState() (AuditChainState, error)
	SaveChainState(state *AuditChainState) error
	DeleteVerifiedPrefix(maxID int64, limit int, apply func([]SecurityEvent) (*AuditChainState, error)) ([]SecurityEvent, error)
}
