//go:build integration

package audit_test

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/tpdenta/afta-reception/internal/platform/security/audit"
	"github.com/tpdenta/afta-reception/internal/platform/testutil"
	"gorm.io/gorm"
)

// TestIntegration_D11_ConcurrentLogAndVerifyChain دو goroutine هر کدام ۵۰ رویداد را روی پایگاه واقعی می‌نویسد.
// سپس زنجیره را از صفر بررسی می‌کند و انتظار زنجیره سالم دارد.
// داده داخل تراکنش می‌ماند و rollback می‌شود. خود جدول AuditChainStates بیرون از تراکنش ساخته می‌شود
// تا بررسی SQL بعدی آن را ببیند؛ AutoMigrate داخل تراکنش با rollback جدول را پاک می‌کرد.
// پکیج تست بیرونی است تا import به testutil دور نسازد.
func TestIntegration_D11_ConcurrentLogAndVerifyChain(t *testing.T) {
	database := testutil.OpenTestDB(t)
	if err := database.AutoMigrate(&audit.SecurityEvent{}, &audit.AuditChainState{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	testutil.WithTestTx(t, func(tx *gorm.DB) {
		if err := tx.Exec("DELETE FROM SecurityEvents").Error; err != nil {
			t.Fatalf("پاک کردن موقت رویدادها: %v", err)
		}
		if err := tx.Exec("DELETE FROM AuditChainStates").Error; err != nil {
			t.Fatalf("پاک کردن موقت وضعیت زنجیره: %v", err)
		}
		m := audit.NewManager(tx, "integration-audit-key")
		const goroutines = 2
		const perGoroutine = 50
		errCh := make(chan error, goroutines*perGoroutine)
		var wg sync.WaitGroup
		for g := 0; g < goroutines; g++ {
			wg.Add(1)
			go func(g int) {
				defer wg.Done()
				for i := 0; i < perGoroutine; i++ {
					if err := m.Record(context.Background(), audit.Event{
						Type: audit.EventLoginSuccess, Result: audit.ResultSuccess,
						Actor: audit.Actor{IP: "127.0.0.1"}, Description: fmt.Sprintf("g%d-%d", g, i),
					}); err != nil {
						errCh <- err
						return
					}
				}
			}(g)
		}
		wg.Wait()
		close(errCh)
		for err := range errCh {
			if err != nil {
				t.Fatal(err)
			}
		}
		var n int64
		if err := tx.Model(&audit.SecurityEvent{}).Count(&n).Error; err != nil {
			t.Fatalf("شمارش: %v", err)
		}
		if n != goroutines*perGoroutine {
			t.Fatalf("تعداد رویداد %d است", n)
		}
		lastID, hash, broken, err := m.VerifyChain(audit.AuditChainState{})
		if err != nil {
			t.Fatalf("بررسی زنجیره: %v", err)
		}
		if broken != 0 || hash == "" || lastID == 0 {
			t.Fatalf("زنجیره سالم نیست: id=%d broken=%d", lastID, broken)
		}
	})
}
