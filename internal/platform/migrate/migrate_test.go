package migrate

import (
	"errors"
	"strings"
	"testing"
)

// TestCombineMigrateErrors_EmptyIsNil ورودی خالی را موفقیت می‌داند.
// failures هیچ شکستی ندارد. خروجی باید nil باشد تا راه‌اندازی ادامه یابد.
func TestCombineMigrateErrors_EmptyIsNil(t *testing.T) {
	if err := combineMigrateErrors(nil); err != nil {
		t.Fatalf("بدون شکست نباید خطا برگردد: %v", err)
	}
	if err := combineMigrateErrors([]tableFailure{}); err != nil {
		t.Fatalf("فهرست خالی نباید خطا برگردد: %v", err)
	}
}

// TestCombineMigrateErrors_NamesEveryTable نام هر جدول شکست‌خورده را در خطا می‌گذارد.
// failures دو جدول با متن خطای جدا هستند. خروجی باید هر دو نام و هر دو متن را داشته باشد.
func TestCombineMigrateErrors_NamesEveryTable(t *testing.T) {
	first := errors.New("ستون ساخته نشد")
	second := errors.New("ایندکس تکراری")
	err := combineMigrateErrors([]tableFailure{
		{Table: "LogCleanupSummaries", Err: first},
		{Table: "UploadCounters", Err: second},
	})
	if err == nil {
		t.Fatal("شکست جدول باید خطا برگرداند")
	}
	msg := err.Error()
	for _, want := range []string{"LogCleanupSummaries", "UploadCounters", "ستون ساخته نشد", "ایندکس تکراری", "AutoMigrate ناموفق بود"} {
		if !strings.Contains(msg, want) {
			t.Fatalf("متن خطا %q را ندارد: %s", want, msg)
		}
	}
	if !errors.Is(err, first) || !errors.Is(err, second) {
		t.Fatalf("خطای اصلی باید قابل بازکردن باشد: %v", err)
	}
}

type namedTable struct{}

func (namedTable) TableName() string { return "UploadCounters" }

type blankTableName struct{}

func (blankTableName) TableName() string { return "  " }

// TestModelTableName_UsesTableNameThenTypeName نام جدول را بدون اتصال پایگاه برمی‌گرداند.
// مدل با TableName همان نام را می‌دهد. نام خالی و مدل بدون متد به نام نوع Go برمی‌گردند.
func TestModelTableName_UsesTableNameThenTypeName(t *testing.T) {
	if got := modelTableName(&namedTable{}); got != "UploadCounters" {
		t.Fatalf("TableName: got %q", got)
	}
	if got := modelTableName(blankTableName{}); got != "blankTableName" {
		t.Fatalf("نام خالی باید به نوع برگردد: got %q", got)
	}
	if got := modelTableName(struct{ ID int }{}); got != "struct { ID int }" {
		t.Fatalf("نوع بی‌نام: got %q", got)
	}
	if got := modelTableName(nil); got != "unknown" {
		t.Fatalf("nil: got %q", got)
	}
}
