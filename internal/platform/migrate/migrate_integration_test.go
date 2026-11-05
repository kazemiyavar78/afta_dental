//go:build integration

package migrate

import (
	"testing"

	"github.com/tpdenta/afta-reception/internal/fund"
	"github.com/tpdenta/afta-reception/internal/organization"
	organizationpackage "github.com/tpdenta/afta-reception/internal/organizationPackage"
	"github.com/tpdenta/afta-reception/internal/patient"
	"github.com/tpdenta/afta-reception/internal/platform/security/audit"
	"github.com/tpdenta/afta-reception/internal/platform/security/loginguard"
	"github.com/tpdenta/afta-reception/internal/platform/security/session"
	"github.com/tpdenta/afta-reception/internal/platform/security/settings"
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
		&user.Role{},
		&user.Permission{},
		&user.RolePermission{},
		&loginguard.LoginAttempt{},
	}
}
