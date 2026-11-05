package logs

import "time"

// EventItem یک ردیف فهرست لاگ است. هش زنجیره و جزئیات خام در پاسخ نیستند.
// برچسب فارسی نوع و دلیل برای نمایش است و جدا از کد ذخیره‌شده فیلتر می‌شود.
type EventItem struct {
	ID             int64     `json:"id"`
	CreatedAt      time.Time `json:"created_at"`
	ActorUsername  string    `json:"actor_username"`
	TargetName     string    `json:"target_name"`
	TargetType     string    `json:"target_type"`
	EventType      string    `json:"event_type"`
	EventTypeLabel string    `json:"event_type_label"`
	Result         string    `json:"result"`
	Reason         string    `json:"reason"`
	ReasonLabel    string    `json:"reason_label"`
	IP             string    `json:"ip"`
	UserAgent      string    `json:"user_agent"`
	Description    string    `json:"description"`
}

// Option یک مقدار فیلتر و برچسب فارسی آن برای فهرست کشویی است.
type Option struct {
	Value string `json:"value"`
	Label string `json:"label"`
}

// MetaResponse فهرست مقدارهای شناخته‌شده برای فیلتر لاگ است.
type MetaResponse struct {
	EventTypes  []Option `json:"event_types"`
	Reasons     []Option `json:"reasons"`
	TargetTypes []Option `json:"target_types"`
}

// IntegrityResponse وضعیت آخرین بررسی زنجیره است.
// هش و لنگر در پاسخ نیستند. FAU_STG.2: نتیجه بررسی به دارنده مجوز خواندن لاگ نشان داده می‌شود.
type IntegrityResponse struct {
	LastCheckAt *time.Time `json:"last_check_at"`
	LastResult  string     `json:"last_result"`
	BrokenAtID  int64      `json:"broken_at_id"`
}

// eventRow نتیجه اسکن کوئری است و هنوز برچسب فارسی ندارد.
type eventRow struct {
	ID            int64     `gorm:"column:id"`
	CreatedAt     time.Time `gorm:"column:created_at"`
	ActorUsername string    `gorm:"column:actor_username"`
	TargetName    string    `gorm:"column:target_name"`
	TargetType    string    `gorm:"column:target_type"`
	EventType     string    `gorm:"column:event_type"`
	Result        string    `gorm:"column:result"`
	Reason        string    `gorm:"column:reason"`
	IP            string    `gorm:"column:ip"`
	UserAgent     string    `gorm:"column:user_agent"`
	Description   string    `gorm:"column:description"`
}
