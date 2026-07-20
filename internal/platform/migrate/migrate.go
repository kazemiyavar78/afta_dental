package migrate

import (
	"fmt"
	"log"

	"gorm.io/gorm"
)

// Migrate جداول را یکی‌یکی AutoMigrate می‌کند؛ خطای یک جدول بقیه را متوقف نمی‌کند.
func Migrate(db *gorm.DB, tables []interface{}) error {
	var firstErr error
	for _, table := range tables {
		if err := db.AutoMigrate(table); err != nil {
			log.Printf("هشدار migrate: %v", err)
			if firstErr == nil {
				firstErr = fmt.Errorf("migrate failed: %w", err)
			}
		}
	}
	return firstErr
}

// DropTable جداول داده‌شده را حذف می‌کند.
func DropTable(db *gorm.DB, tables []interface{}) error {
	for _, table := range tables {
		if err := db.Migrator().DropTable(table); err != nil {
			return err
		}
	}
	return nil
}
