package loginguard

import (
	"testing"
	"time"
)

// TestParseTimeOfDay تبدیل زمان را بررسی می‌کند.
func TestParseTimeOfDay(t *testing.T) {
	minutes, err := parseTimeOfDay("08:30")
	if err != nil {
		t.Fatalf("خطا: %v", err)
	}
	if minutes != 8*60+30 {
		t.Fatalf("انتظار %d، دریافت %d", 8*60+30, minutes)
	}
}

// TestLoginAttemptModel ساختار مدل را بررسی می‌کند.
func TestLoginAttemptModel(t *testing.T) {
	la := LoginAttempt{
		Ip:            "192.168.1.1",
		FailedCount:   3,
		LastAttemptAt: time.Now(),
	}
	if la.FailedCount != 3 {
		t.Fatal("شمارنده باید ۳ باشد")
	}
}
