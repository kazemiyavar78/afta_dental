package loginguard

import (
	"sync"
	"time"
)

// BlockedLoginLog زمان آخرین LOGIN_BLOCKED هر IP را در حافظه نگه می‌دارد.
// هدف این است که در مدت قفل، برای هر IP بیش از یک رویداد در دقیقه نوشته نشود.
// R1-18: شمارنده پایگاه نیست؛ با ری‌استارت فرایند از نو شروع می‌شود.
type BlockedLoginLog struct {
	mu   sync.Mutex
	last map[string]time.Time
	now  func() time.Time
}

// NewBlockedLoginLog شمارنده خالی می‌سازد.
// خروجی آماده Allow است. ساعت پیش‌فرض time.Now است و تست می‌تواند now را عوض کند.
func NewBlockedLoginLog() *BlockedLoginLog {
	return &BlockedLoginLog{
		last: map[string]time.Time{},
		now:  time.Now,
	}
}

// Allow می‌گوید این IP همین حالا مجاز است یک LOGIN_BLOCKED تازه داشته باشد.
// ip نشانی درخواست است. خروجی true یعنی باید رویداد ثبت شود و پنجره یک‌دقیقه‌ای از همین لحظه شروع شده است.
// خروجی false یعنی در شصت ثانیه گذشته برای همین IP رویداد ثبت شده و این تلاش نباید لاگ شود.
func (b *BlockedLoginLog) Allow(ip string) bool {
	if b == nil {
		return true
	}
	now := time.Now()
	if b.now != nil {
		now = b.now()
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.last == nil {
		b.last = map[string]time.Time{}
	}
	if prev, ok := b.last[ip]; ok && now.Sub(prev) < time.Minute {
		return false
	}
	b.last[ip] = now
	return true
}
