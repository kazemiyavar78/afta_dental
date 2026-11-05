package audit

// EventLogger قرارداد ثبت قدیمی است و تا مهاجرت فراخوانی‌ها می‌ماند.
// LogEvent شناسه کاربر (nil اگر فاعل ناشناس باشد)، IP، نوع رویداد و توضیح را می‌گیرد
// و خطای ذخیره‌سازی را برمی‌گرداند. Manager همین متد را دارد و ردیف نسخه ۱ می‌نویسد.
// ثبت جدید Record روی Manager است. سرویس‌ها هنوز *Manager می‌گیرند.
type EventLogger interface {
	LogEvent(userID *int, ip, eventType, description string) error
}

// اطمینان می‌دهد Manager بدون تغییر سرویس‌ها همین قرارداد را برآورده می‌کند.
var _ EventLogger = (*Manager)(nil)
