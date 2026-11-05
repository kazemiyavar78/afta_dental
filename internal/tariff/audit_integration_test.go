//go:build integration

package tariff

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/tpdenta/afta-reception/internal/organization"
	organizationpackage "github.com/tpdenta/afta-reception/internal/organizationPackage"
	"github.com/tpdenta/afta-reception/internal/platform/security/audit"
	"github.com/tpdenta/afta-reception/internal/platform/security/encryption"
	"github.com/tpdenta/afta-reception/internal/platform/security/integrity"
	"github.com/tpdenta/afta-reception/internal/platform/testutil"
	"github.com/tpdenta/afta-reception/internal/services"
	"gorm.io/gorm"
)

// TestIntegration_D13_TariffSaveRecordsActorAndTarget ذخیره تعرفه یک خدمت را با فاعل و کد خدمت ثبت می‌کند.
// محاسبه تعرفه به سازمان، بسته و فهرست خدمات پایگاه وصل است و بدون پایگاه قابل جداسازی نیست.
func TestIntegration_D13_TariffSaveRecordsActorAndTarget(t *testing.T) {
	database := testutil.OpenTestDB(t)
	if err := database.AutoMigrate(
		&organizationpackage.OrganizationPackage{}, &organization.Organization{},
		&services.ServiceItem{}, &Tariff{}, &audit.SecurityEvent{}, &audit.AuditChainState{},
	); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	testutil.WithTestTx(t, func(tx *gorm.DB) {
		enc, err := encryption.NewEncryptor("12345678901234567890123456789012")
		if err != nil {
			t.Fatal(err)
		}
		mgr := audit.NewManager(tx, "integration-audit-key")
		actorID := 4
		actor := audit.Actor{UserID: &actorID, Username: "reza", IP: "10.2.2.2"}
		pkgSvc := organizationpackage.NewService(tx, mgr, encryption.NewOrganizationPackageEncryptionService(enc))
		pkg, err := pkgSvc.Create(context.Background(), organizationpackage.CreateRequest{PackageName: "بسته تست تعرفه"}, actor)
		if err != nil {
			t.Fatal(err)
		}
		orgSvc := organization.NewService(tx, mgr, encryption.NewOrganizationEncryptionService(enc), pkgSvc)
		org, err := orgSvc.Create(context.Background(), organization.CreateRequest{
			Name: "سازمان تست تعرفه", IsActive: true, PackageID: pkg.ID, CenterPackageID: pkg.ID,
		}, actor)
		if err != nil {
			t.Fatal(err)
		}
		serviceSvc := services.NewService(tx, mgr, encryption.NewServiceEncryptionService(enc), integrity.NewSigner("test-integrity-key-32bytes-long!!"))
		code := fmt.Sprintf("T%d", time.Now().UnixNano())
		item, err := serviceSvc.Create(context.Background(), services.CreateRequest{
			ServiceCode: code, Name: "خدمت تست", ServiceRate: 1000,
		}, actor)
		if err != nil {
			t.Fatal(err)
		}
		all, err := serviceSvc.List()
		if err != nil {
			t.Fatal(err)
		}
		exclude := make([]uint, 0)
		for _, row := range all {
			if row.ID != item.ID {
				exclude = append(exclude, row.ID)
			}
		}
		svc := NewService(tx, mgr, orgSvc, serviceSvc)
		if _, err := svc.SaveTariffsForOrganization(context.Background(), CalculateTariffForOrganizationRequest{
			OrganizationID: org.ID, ExcludeServiceIDs: exclude,
		}, actor); err != nil {
			t.Fatal(err)
		}
		var rows []audit.SecurityEvent
		if err := tx.Where("TargetType = ? AND EventType = ?", audit.TargetTariff, audit.EventEntityCreated).Find(&rows).Error; err != nil {
			t.Fatal(err)
		}
		if len(rows) != 1 {
			t.Fatalf("یک رویداد ایجاد تعرفه انتظار می‌رود، دریافت %d", len(rows))
		}
		row := rows[0]
		if row.UserID == nil || *row.UserID != 4 || row.ActorUsername != "reza" || row.TargetName != code {
			t.Fatalf("فاعل یا کد خدمت هدف نادرست است: %+v", row)
		}
	})
}
