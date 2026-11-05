package audit

import (
	"context"

	"gorm.io/gorm"
)

// EventLogger قرارداد ثبت قدیمی است و تا مهاجرت فراخوانی‌های باقی‌مانده می‌ماند.
// LogEvent شناسه کاربر (nil اگر فاعل ناشناس باشد)، IP، نوع رویداد و توضیح را می‌گیرد
// و خطای ذخیره‌سازی را برمی‌گرداند. Manager همین متد را دارد و ردیف نسخه ۱ می‌نویسد.
type EventLogger interface {
	LogEvent(userID *int, ip, eventType, description string) error
}

// Recorder ثبت رویداد نسخه ۲ است.
// Record و RecordTx همان متدهای Manager هستند. FakeAuditRecorder برای تست همین قرارداد را دارد.
// سرویس‌های مهاجرت‌شده به این اینترفیس وابسته‌اند تا تست بدون پایگاه، فاعل و جزئیات را ببیند.
type Recorder interface {
	Record(ctx context.Context, e Event) error
	RecordTx(tx *gorm.DB, e Event) error
}

// اطمینان می‌دهد Manager هر دو قرارداد ثبت را برآورده می‌کند.
var (
	_ EventLogger = (*Manager)(nil)
	_ Recorder    = (*Manager)(nil)
)
