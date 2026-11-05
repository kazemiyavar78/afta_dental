package migrate

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"gorm.io/gorm"
)

// tableFailure خطای AutoMigrate یک مدل است.
// Table نام جدول است و Err خطای GORM همان مدل.
type tableFailure struct {
	Table string
	Err   error
}

// modelTableName نام جدول مدل را برای متن خطا برمی‌گرداند.
// model همان مقداری است که به AutoMigrate داده می‌شود.
// اگر TableName تعریف شده و خالی نباشد همان را برمی‌گرداند؛ وگرنه نام نوع Go.
func modelTableName(model interface{}) string {
	if namer, ok := model.(interface{ TableName() string }); ok {
		if name := strings.TrimSpace(namer.TableName()); name != "" {
			return name
		}
	}
	t := reflect.TypeOf(model)
	if t == nil {
		return "unknown"
	}
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	if t.Name() != "" {
		return t.Name()
	}
	return t.String()
}

// combineMigrateErrors خطاهای ساخت جدول را در یک خطا جمع می‌کند.
// failures فهرست جدول‌هایی است که AutoMigrate آن‌ها شکست خورده. ورودی خالی یعنی موفق.
// اگر شکست باشد یک خطا برمی‌گرداند که نام هر جدول را دارد تا فراخواننده سرور را متوقف کند.
// FPT_FLS.1: شکست حفظ وضعیت امن — سرور نباید با schema ناقص بالا بیاید.
func combineMigrateErrors(failures []tableFailure) error {
	if len(failures) == 0 {
		return nil
	}
	joined := make([]error, 0, len(failures))
	for _, failure := range failures {
		joined = append(joined, fmt.Errorf("جدول %s: %w", failure.Table, failure.Err))
	}
	return fmt.Errorf("AutoMigrate ناموفق بود: %w", errors.Join(joined...))
}

// Migrate هر مدل را با GORM AutoMigrate می‌سازد یا به‌روز می‌کند.
// db اتصال باز است. tables مدل‌هایی هستند که جدولشان باید وجود داشته باشد.
// اگر حتی یک مدل خطا بدهد، خطای جمع‌شده با نام جدول برمی‌گردد و هیچ جدولی بی‌صدا رد نمی‌شود.
// FPT_FLS.1: خطا برمی‌گردد تا main فرایند را متوقف کند؛ ادامه با schema ناقص ممنوع است.
func Migrate(db *gorm.DB, tables []interface{}) error {
	var failures []tableFailure
	for _, table := range tables {
		if err := db.AutoMigrate(table); err != nil {
			failures = append(failures, tableFailure{
				Table: modelTableName(table),
				Err:   err,
			})
		}
	}
	return combineMigrateErrors(failures)
}

// DropTable جداول داده‌شده را حذف می‌کند.
// db اتصال باز است و tables مدل‌هایی هستند که باید حذف شوند.
// خطای اولین جدول را برمی‌گرداند و بقیه را حذف نمی‌کند.
func DropTable(db *gorm.DB, tables []interface{}) error {
	for _, table := range tables {
		if err := db.Migrator().DropTable(table); err != nil {
			return err
		}
	}
	return nil
}
