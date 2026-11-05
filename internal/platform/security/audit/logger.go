package audit

// EventLogger قرارداد ثبت رویداد ممیزی است.
// LogEvent شناسه کاربر (nil اگر فاعل ناشناس باشد)، IP، نوع رویداد و توضیح را می‌گیرد
// و خطای ذخیره‌سازی را برمی‌گرداند. Manager همین متد را دارد.
// سرویس‌ها هنوز *Manager می‌گیرند؛ وصل‌کردنشان به این interface کار فاز ۱ است.
type EventLogger interface {
	LogEvent(userID *int, ip, eventType, description string) error
}

// اطمینان می‌دهد Manager بدون تغییر سرویس‌ها همین قرارداد را برآورده می‌کند.
var _ EventLogger = (*Manager)(nil)
