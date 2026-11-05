//go:build integration

package migrate

import (
	"fmt"
	"strings"
	"testing"

	"github.com/tpdenta/afta-reception/internal/fund"
	"github.com/tpdenta/afta-reception/internal/organization"
	organizationpackage "github.com/tpdenta/afta-reception/internal/organizationPackage"
	"github.com/tpdenta/afta-reception/internal/patient"
	"github.com/tpdenta/afta-reception/internal/platform/security/audit"
	"github.com/tpdenta/afta-reception/internal/platform/security/loginguard"
	"github.com/tpdenta/afta-reception/internal/platform/security/logretention"
	"github.com/tpdenta/afta-reception/internal/platform/security/session"
	"github.com/tpdenta/afta-reception/internal/platform/security/settings"
	"github.com/tpdenta/afta-reception/internal/platform/security/uploadguard"
	"github.com/tpdenta/afta-reception/internal/platform/testutil"
	"github.com/tpdenta/afta-reception/internal/reception"
	"github.com/tpdenta/afta-reception/internal/regulation"
	"github.com/tpdenta/afta-reception/internal/services"
	"github.com/tpdenta/afta-reception/internal/specialcode"
	"github.com/tpdenta/afta-reception/internal/tariff"
	"github.com/tpdenta/afta-reception/internal/user"
	"github.com/tpdenta/afta-reception/internal/wallet"
	"gorm.io/gorm"
)

// TestIntegration_AutoMigrate_AllModelsCreate همان مدل‌های cmd/server/main.go را migrate می‌کند.
// اگر حتی یک جدول خطا بدهد Migrate همان خطا را برمی‌گرداند و این تست شکست می‌خورد.
// تراکنش در پایان rollback می‌شود تا اثرش در پایگاه نماند.
func TestIntegration_AutoMigrate_AllModelsCreate(t *testing.T) {
	testutil.WithTestTx(t, func(tx *gorm.DB) {
		if err := Migrate(tx, productionModels()); err != nil {
			t.Fatalf("حداقل یک جدول migrate نشد: %v", err)
		}
	})
}

// productionModels فهرست AutoMigrate در cmd/server/main.go است.
// با تغییر آن فهرست، این تابع هم باید همان مدل‌ها را داشته باشد.
func productionModels() []interface{} {
	return []interface{}{
		&user.User{},
		&organizationpackage.OrganizationPackage{},
		&organizationpackage.OrganizationPackageRelation{},
		&organization.Organization{},
		&patient.Patient{},
		&services.ServiceItem{},
		&services.ExcludedService{},
		&fund.Fund{},
		&tariff.Tariff{},
		&reception.Reception{},
		&reception.ReceptionService{},
		&reception.ReceptionPhoto{},
		&specialcode.SpecialCode{},
		&regulation.Regulation{},
		&regulation.RegulationService{},
		&wallet.CashFund{},
		&wallet.PatientWallet{},
		&wallet.BankAccount{},
		&wallet.CreditAccount{},
		&wallet.Transaction{},
		&settings.SecuritySetting{},
		&session.Session{},
		&audit.SecurityEvent{},
		&audit.AuditChainState{},
		&user.Role{},
		&user.Permission{},
		&user.RolePermission{},
		&loginguard.LoginAttempt{},
		&logretention.LogCleanupSummary{},
		&uploadguard.UploadCounter{},
	}
}

// schemaTableRow یک نام جدول از INFORMATION_SCHEMA.TABLES است.
type schemaTableRow struct {
	TableName string `gorm:"column:TABLE_NAME"`
}

// TestIntegration_LogCleanupAndUploadCounterTablesExist بعد از Migrate وجود دو جدول جاافتاده را می‌سنجد.
// tx همان تراکنشی است که Migrate روی آن اجرا شده و در پایان rollback می‌شود.
// اگر LogCleanupSummaries یا UploadCounters در INFORMATION_SCHEMA نباشد تست شکست می‌خورد.
func TestIntegration_LogCleanupAndUploadCounterTablesExist(t *testing.T) {
	testutil.WithTestTx(t, func(tx *gorm.DB) {
		if err := Migrate(tx, productionModels()); err != nil {
			t.Fatalf("migrate: %v", err)
		}
		required := []string{"LogCleanupSummaries", "UploadCounters"}
		missing := missingSchemaTables(t, tx, required)
		if len(missing) > 0 {
			t.Fatalf("بعد از Migrate این جدول‌ها در INFORMATION_SCHEMA نیستند: %s", strings.Join(missing, ", "))
		}
	})
}

// TestIntegration_SecurityEventNewColumnsExist ستون‌های فاعل، هدف و نسخه هش را بعد از Migrate می‌سنجد.
// tx تراکنش تست است و در پایان rollback می‌شود. نبودن هر ستون تست را شکست می‌دهد.
func TestIntegration_SecurityEventNewColumnsExist(t *testing.T) {
	testutil.WithTestTx(t, func(tx *gorm.DB) {
		if err := Migrate(tx, []interface{}{&audit.SecurityEvent{}}); err != nil {
			t.Fatalf("migrate: %v", err)
		}
		required := []string{
			"ActorUsername", "TargetType", "TargetID", "TargetName",
			"Result", "Reason", "UserAgent", "Details", "HashVersion",
		}
		missing := missingNullableAuditColumns(t, tx, required)
		if len(missing) > 0 {
			t.Fatalf("ستون‌های SecurityEvents پس از Migrate نیستند یا تهی‌پذیری‌شان نادرست است: %s", strings.Join(missing, ", "))
		}
	})
}

// schemaColumnRow یک ستون از INFORMATION_SCHEMA.COLUMNS است.
type schemaColumnRow struct {
	ColumnName    string `gorm:"column:COLUMN_NAME"`
	IsNullable    string `gorm:"column:IS_NULLABLE"`
	ColumnDefault string `gorm:"column:COLUMN_DEFAULT"`
}

// missingNullableAuditColumns ستون‌هایی را برمی‌گرداند که نیستند یا قاعده تهی‌پذیری را ندارند.
// db تراکنش تست است و names ستون‌های تازه‌اند. HashVersion باید NOT NULL و پیش‌فرض ۱ باشد.
// بقیه باید تهی‌پذیر باشند تا ردیف قدیمی هنگام افزودن ستون نشکند.
func missingNullableAuditColumns(t *testing.T, db *gorm.DB, names []string) []string {
	t.Helper()
	placeholders := make([]string, len(names))
	args := make([]interface{}, 0, len(names)+1)
	args = append(args, "SecurityEvents")
	for i, name := range names {
		placeholders[i] = "?"
		args = append(args, name)
	}
	query := fmt.Sprintf(`
		SELECT COLUMN_NAME, IS_NULLABLE, COLUMN_DEFAULT
		FROM INFORMATION_SCHEMA.COLUMNS
		WHERE TABLE_NAME = ? AND COLUMN_NAME IN (%s)
	`, strings.Join(placeholders, ", "))

	var found []schemaColumnRow
	if err := db.Raw(query, args...).Scan(&found).Error; err != nil {
		t.Fatalf("خواندن ستون‌ها: %v", err)
	}
	byName := make(map[string]schemaColumnRow, len(found))
	for _, row := range found {
		byName[row.ColumnName] = row
	}
	var missing []string
	for _, name := range names {
		row, ok := byName[name]
		if !ok {
			missing = append(missing, name)
			continue
		}
		if name == "HashVersion" {
			if !strings.EqualFold(row.IsNullable, "NO") || !strings.Contains(row.ColumnDefault, "1") {
				missing = append(missing, name+" (default)")
			}
			continue
		}
		if !strings.EqualFold(row.IsNullable, "YES") {
			missing = append(missing, name+" (null)")
		}
	}
	return missing
}

// missingSchemaTables نام‌هایی را برمی‌گرداند که در INFORMATION_SCHEMA.TABLES نیستند.
// db اتصال یا تراکنش تست است و names نام جدول‌های مورد انتظار هستند.
// خروجی زیرمجموعهٔ names است؛ فهرست خالی یعنی همه موجودند. خطای خواندن تست را متوقف می‌کند.
func missingSchemaTables(t *testing.T, db *gorm.DB, names []string) []string {
	t.Helper()
	placeholders := make([]string, len(names))
	args := make([]interface{}, len(names))
	for i, name := range names {
		placeholders[i] = "?"
		args[i] = name
	}
	query := fmt.Sprintf(`
		SELECT TABLE_NAME
		FROM INFORMATION_SCHEMA.TABLES
		WHERE TABLE_TYPE = 'BASE TABLE' AND TABLE_NAME IN (%s)
	`, strings.Join(placeholders, ", "))

	var found []schemaTableRow
	if err := db.Raw(query, args...).Scan(&found).Error; err != nil {
		t.Fatalf("خواندن INFORMATION_SCHEMA: %v", err)
	}
	have := make(map[string]struct{}, len(found))
	for _, row := range found {
		have[row.TableName] = struct{}{}
	}
	var missing []string
	for _, name := range names {
		if _, ok := have[name]; !ok {
			missing = append(missing, name)
		}
	}
	return missing
}
