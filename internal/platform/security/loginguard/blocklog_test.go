package loginguard

import (
	"testing"
	"time"
)

// TestR1_18_LoginBlockedOncePerMinute برای هر IP در هر دقیقه فقط یک اجازه ثبت می‌دهد.
func TestR1_18_LoginBlockedOncePerMinute(t *testing.T) {
	start := time.Date(2026, 11, 5, 12, 0, 0, 0, time.UTC)
	now := start
	log := NewBlockedLoginLog()
	log.now = func() time.Time { return now }

	if !log.Allow("203.0.113.9") {
		t.Fatal("اولین تلاش IP باید ثبت شود")
	}
	if log.Allow("203.0.113.9") {
		t.Fatal("تلاش دوم همان IP در همان دقیقه نباید ثبت شود")
	}
	if !log.Allow("203.0.113.10") {
		t.Fatal("IP دیگر باید ثبت شود")
	}
	now = start.Add(time.Minute)
	if !log.Allow("203.0.113.9") {
		t.Fatal("بعد از یک دقیقه همان IP دوباره باید ثبت شود")
	}
}
