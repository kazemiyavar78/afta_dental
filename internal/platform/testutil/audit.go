package testutil

import "github.com/tpdenta/afta-reception/internal/platform/security/audit"

// RecordedAuditEvent یک فراخوانی LogEvent است که تست بعداً آن را می‌خواند.
type RecordedAuditEvent struct {
	UserID      *int
	IP          string
	EventType   string
	Description string
}

// FakeAuditRecorder رویدادهای ممیزی را در حافظه نگه می‌دارد و به پایگاه وصل نمی‌شود.
// تست‌ها Events را می‌خوانند تا فاعل، نوع رویداد و توضیح ثبت‌شده را بررسی کنند.
type FakeAuditRecorder struct {
	Events []RecordedAuditEvent
}

// LogEvent رویداد را به Events اضافه می‌کند.
// userID، ip، eventType و description همان آرگومان‌های audit.EventLogger هستند.
// شناسه کاربر کپی می‌شود تا تغییر بعدی اشاره‌گرِ صداکننده، سابقه را عوض نکند.
// خروجی همیشه nil است چون این پیاده‌سازی خطای پایگاه ندارد.
func (f *FakeAuditRecorder) LogEvent(userID *int, ip, eventType, description string) error {
	var storedID *int
	if userID != nil {
		id := *userID
		storedID = &id
	}
	f.Events = append(f.Events, RecordedAuditEvent{
		UserID:      storedID,
		IP:          ip,
		EventType:   eventType,
		Description: description,
	})
	return nil
}

// FakeAuditRecorder باید همان قرارداد LogEvent را داشته باشد که Manager دارد.
var _ audit.EventLogger = (*FakeAuditRecorder)(nil)
