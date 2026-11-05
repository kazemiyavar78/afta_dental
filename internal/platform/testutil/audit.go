package testutil

import (
	"context"

	"github.com/tpdenta/afta-reception/internal/platform/security/audit"
	"gorm.io/gorm"
)

// RecordedAuditEvent یک فراخوانی ثبت ممیزی است که تست بعداً آن را می‌خواند.
// Event وقتی پر است که ثبت از Record آمده باشد. فیلدهای تخت برای LogEvent قدیمی مانده‌اند.
type RecordedAuditEvent struct {
	UserID      *int
	IP          string
	EventType   string
	Description string
	Event       audit.Event
}

// FakeAuditRecorder رویدادهای ممیزی را در حافظه نگه می‌دارد و به پایگاه وصل نمی‌شود.
// Err اگر تنظیم شود Record و RecordTx همان خطا را برمی‌گردانند تا شکست ثبت تست شود.
type FakeAuditRecorder struct {
	Events []RecordedAuditEvent
	Err    error
}

// LogEvent رویداد نسخه ۱ را به Events اضافه می‌کند.
// userID، ip، eventType و description همان آرگومان‌های audit.EventLogger هستند.
// شناسه کاربر کپی می‌شود تا تغییر بعدی اشاره‌گرِ صداکننده، سابقه را عوض نکند.
// خروجی همیشه nil است چون این پیاده‌سازی خطای پایگاه ندارد.
func (f *FakeAuditRecorder) LogEvent(userID *int, ip, eventType, description string) error {
	storedID := copyID(userID)
	f.Events = append(f.Events, RecordedAuditEvent{
		UserID:      storedID,
		IP:          ip,
		EventType:   eventType,
		Description: description,
		Event: audit.Event{
			Type:        eventType,
			Description: description,
			Actor:       audit.Actor{UserID: storedID, IP: ip},
		},
	})
	return nil
}

// Record رویداد نسخه ۲ را نگه می‌دارد.
// ctx لغو درخواست است و اگر تمام شده باشد خطا برمی‌گردد. e همان ورودی Manager.Record است.
// اگر Err تنظیم شده باشد رویداد ذخیره نمی‌شود. شناسه فاعل کپی می‌شود.
func (f *FakeAuditRecorder) Record(ctx context.Context, e audit.Event) error {
	if ctx != nil {
		if err := ctx.Err(); err != nil {
			return err
		}
	}
	if f.Err != nil {
		return f.Err
	}
	e.Actor.UserID = copyID(e.Actor.UserID)
	f.Events = append(f.Events, RecordedAuditEvent{
		UserID:      e.Actor.UserID,
		IP:          e.Actor.IP,
		EventType:   e.Type,
		Description: e.Description,
		Event:       e,
	})
	return nil
}

// RecordTx همان Record است چون تست تراکنش پایگاه باز نمی‌کند.
// tx نادیده گرفته می‌شود. e رویداد داخل عملیات است. خروجی خطای Err یا nil است.
func (f *FakeAuditRecorder) RecordTx(_ *gorm.DB, e audit.Event) error {
	return f.Record(context.Background(), e)
}

// copyID یک کپی از شناسه می‌سازد تا سابقه تست به متغیر صداکننده وصل نماند.
// userID می‌تواند nil باشد. خروجی nil همان فاعل بدون شناسه است.
func copyID(userID *int) *int {
	if userID == nil {
		return nil
	}
	id := *userID
	return &id
}

// FakeAuditRecorder باید قرارداد ثبت قدیمی و جدید را داشته باشد.
var (
	_ audit.EventLogger = (*FakeAuditRecorder)(nil)
	_ audit.Recorder    = (*FakeAuditRecorder)(nil)
)
