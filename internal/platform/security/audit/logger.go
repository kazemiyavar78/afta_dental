package audit

import (
	"context"

	"gorm.io/gorm"
)

// Recorder ثبت رویداد نسخه ۲ است.
// Record و RecordTx همان متدهای Manager هستند. FakeAuditRecorder برای تست همین قرارداد را دارد.
// سرویس‌های مهاجرت‌شده به این اینترفیس وابسته‌اند تا تست بدون پایگاه، فاعل و جزئیات را ببیند.
type Recorder interface {
	Record(ctx context.Context, e Event) error
	RecordTx(tx *gorm.DB, e Event) error
}

// اطمینان می‌دهد Manager قرارداد ثبت نسخه ۲ را برآورده می‌کند.
var _ Recorder = (*Manager)(nil)
