package organization

import (
	"fmt"
	"log"

	"gorm.io/gorm"
)

// EnsureCenterPackageColumn قبل از AutoMigrate، ستون CenterPackageID را برای MSSQL آماده می‌کند:
// مقادیر NULL را پر می‌کند، ایندکس/FK وابسته را موقتاً حذف می‌کند و سپس NOT NULL می‌گذارد.
func EnsureCenterPackageColumn(db *gorm.DB) error {
	if !db.Migrator().HasTable(&Organization{}) {
		return nil
	}
	if !db.Migrator().HasColumn(&Organization{}, "CenterPackageID") {
		return nil
	}

	var isNullable string
	err := db.Raw(`
		SELECT IS_NULLABLE
		FROM INFORMATION_SCHEMA.COLUMNS
		WHERE TABLE_NAME = 'Organizations' AND COLUMN_NAME = 'CenterPackageID'
	`).Scan(&isNullable).Error
	if err != nil {
		return fmt.Errorf("بررسی nullability ستون CenterPackageID: %w", err)
	}
	if isNullable != "YES" {
		return nil
	}

	if err := db.Exec(`
		UPDATE Organizations
		SET CenterPackageID = PackageID
		WHERE CenterPackageID IS NULL AND PackageID IS NOT NULL
	`).Error; err != nil {
		return fmt.Errorf("پر کردن CenterPackageID خالی: %w", err)
	}

	var remainingNulls int64
	if err := db.Raw(`SELECT COUNT(1) FROM Organizations WHERE CenterPackageID IS NULL`).Scan(&remainingNulls).Error; err != nil {
		return fmt.Errorf("شمارش CenterPackageID خالی: %w", err)
	}
	if remainingNulls > 0 {
		return fmt.Errorf("هنوز %d ردیف با CenterPackageID خالی وجود دارد؛ ابتدا PackageID را تکمیل کنید", remainingNulls)
	}

	if err := dropCenterPackageDependencies(db); err != nil {
		return err
	}

	if err := db.Exec(`ALTER TABLE Organizations ALTER COLUMN CenterPackageID bigint NOT NULL`).Error; err != nil {
		return fmt.Errorf("تغییر CenterPackageID به NOT NULL: %w", err)
	}

	log.Printf("مهاجرت: ستون Organizations.CenterPackageID به NOT NULL تبدیل شد")
	return nil
}

// dropCenterPackageDependencies ایندکس‌ها و FKهای وابسته به CenterPackageID را حذف می‌کند تا ALTER COLUMN در MSSQL ممکن شود.
func dropCenterPackageDependencies(db *gorm.DB) error {
	var fkNames []string
	if err := db.Raw(`
		SELECT fk.name
		FROM sys.foreign_keys fk
		INNER JOIN sys.foreign_key_columns fkc ON fk.object_id = fkc.constraint_object_id
		INNER JOIN sys.columns c ON fkc.parent_object_id = c.object_id AND fkc.parent_column_id = c.column_id
		WHERE OBJECT_NAME(fk.parent_object_id) = 'Organizations'
		  AND c.name = 'CenterPackageID'
	`).Scan(&fkNames).Error; err != nil {
		return fmt.Errorf("یافتن FKهای CenterPackageID: %w", err)
	}
	for _, name := range fkNames {
		if err := db.Exec(fmt.Sprintf(`ALTER TABLE Organizations DROP CONSTRAINT [%s]`, name)).Error; err != nil {
			return fmt.Errorf("حذف FK %s: %w", name, err)
		}
	}

	var indexNames []string
	if err := db.Raw(`
		SELECT DISTINCT i.name
		FROM sys.indexes i
		INNER JOIN sys.index_columns ic ON i.object_id = ic.object_id AND i.index_id = ic.index_id
		INNER JOIN sys.columns c ON ic.object_id = c.object_id AND ic.column_id = c.column_id
		WHERE OBJECT_NAME(i.object_id) = 'Organizations'
		  AND c.name = 'CenterPackageID'
		  AND i.is_primary_key = 0
		  AND i.name IS NOT NULL
	`).Scan(&indexNames).Error; err != nil {
		return fmt.Errorf("یافتن ایندکس‌های CenterPackageID: %w", err)
	}
	for _, name := range indexNames {
		if err := db.Exec(fmt.Sprintf(`DROP INDEX [%s] ON Organizations`, name)).Error; err != nil {
			return fmt.Errorf("حذف ایندکس %s: %w", name, err)
		}
	}

	return nil
}
