package logretention

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

// TestWorker_FirstCycleRunsBeforeTicker دور اول را پیش از اولین تیک اجرا می‌کند.
// FAU_STG.2: وضعیت زنجیره نباید تا پایان فاصله ticker خالی بماند.
func TestWorker_FirstCycleRunsBeforeTicker(t *testing.T) {
	var n atomic.Int32
	ctx, cancel := context.WithCancel(context.Background())
	ticks := make(chan time.Time)
	done := make(chan struct{})
	go func() {
		defer close(done)
		runCycles(ctx, ticks, func() { n.Add(1) })
	}()

	waitForCount(t, &n, 1)
	if got := n.Load(); got != 1 {
		t.Fatalf("پیش از تیک باید فقط دور اول اجرا شده باشد: %d", got)
	}
	ticks <- time.Now()
	waitForCount(t, &n, 2)
	cancel()
	<-done
	if got := n.Load(); got != 2 {
		t.Fatalf("بعد از یک تیک باید دو دور اجرا شده باشد: %d", got)
	}
}

// waitForCount تا رسیدن شمارنده یا پایان مهلت صبر می‌کند.
// t تست است. n شمارنده دورها و want تعداد مورد انتظار است.
func waitForCount(t *testing.T, n *atomic.Int32, want int32) {
	t.Helper()
	deadline := time.Now().Add(time.Second)
	for n.Load() < want && time.Now().Before(deadline) {
		time.Sleep(5 * time.Millisecond)
	}
	if n.Load() < want {
		t.Fatalf("شمارنده به %d نرسید: %d", want, n.Load())
	}
}
