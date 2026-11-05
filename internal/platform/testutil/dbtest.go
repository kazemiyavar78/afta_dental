//go:build integration

package testutil

import (
	"os"
	"regexp"
	"strings"
	"sync"
	"testing"

	"github.com/tpdenta/afta-reception/internal/platform/db"
	"gorm.io/gorm"
)

var (
	sharedTestDB   *gorm.DB
	openTestDBOnce sync.Once
	openTestDBErr  error

	sqlServerURLPattern = regexp.MustCompile(`(?i)\b(?:sqlserver|mssql)://\S+`)
	secretKVPattern     = regexp.MustCompile(`(?i)(password|pwd|pass|user id|uid)\s*=\s*\S+`)
)

// OpenTestDB اتصال مشترک تست را از AFTA_TEST_DATABASE_URL برمی‌گرداند.
// اگر متغیر خالی باشد تست با t.Skip رد می‌شود و اتصالی باز نمی‌شود.
// خروجی را نبندید؛ عمر آن به اندازه فرایند تست است.
func OpenTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	raw := strings.TrimSpace(os.Getenv("AFTA_TEST_DATABASE_URL"))
	if raw == "" {
		t.Skip("AFTA_TEST_DATABASE_URL خالی است؛ تست integration اجرا نمی‌شود")
	}

	openTestDBOnce.Do(func() {
		sharedTestDB, openTestDBErr = db.Connect(raw)
	})
	if openTestDBErr != nil {
		t.Fatalf("اتصال به پایگاه تست برقرار نشد: %s", redactTestLog(openTestDBErr.Error()))
	}
	return sharedTestDB
}

// WithTestTx تابع fn را داخل یک تراکنش اجرا می‌کند و در پایان rollback می‌کند.
// t تست جاری است و fn همان کار تست را با *gorm.DB تراکنش انجام می‌دهد.
// داده fn در پایگاه نمی‌ماند، حتی اگر تست شکست بخورد.
func WithTestTx(t *testing.T, fn func(tx *gorm.DB)) {
	t.Helper()
	database := OpenTestDB(t)
	tx := database.Begin()
	if tx.Error != nil {
		t.Fatalf("شروع تراکنش تست ناموفق بود: %s", redactTestLog(tx.Error.Error()))
	}
	t.Cleanup(func() {
		rb := tx.Rollback()
		if rb.Error != nil {
			t.Errorf("rollback تراکنش تست ناموفق بود: %s", redactTestLog(rb.Error.Error()))
		}
	})
	fn(tx)
}

// redactTestLog رمز و نشانی اتصال را از متن خطای درایور حذف می‌کند تا لاگ تست آن را چاپ نکند.
func redactTestLog(msg string) string {
	msg = sqlServerURLPattern.ReplaceAllString(msg, "***")
	msg = secretKVPattern.ReplaceAllString(msg, "$1=***")
	return msg
}
