package organizationpackage

import (
	"context"
	"testing"

	"github.com/tpdenta/afta-reception/internal/platform/security/audit"
	"github.com/tpdenta/afta-reception/internal/platform/security/encryption"
	"github.com/tpdenta/afta-reception/internal/platform/testutil"
	"gorm.io/gorm"
)

type memPackageRepo struct{}

func (memPackageRepo) Create(o *OrganizationPackage) error {
	o.ID = 4
	return nil
}
func (memPackageRepo) Update(*OrganizationPackage) error { return nil }
func (memPackageRepo) Delete(*OrganizationPackage) error { return nil }
func (memPackageRepo) FindByID(int) (*OrganizationPackage, error) {
	return nil, gorm.ErrRecordNotFound
}
func (memPackageRepo) FindAll() ([]OrganizationPackage, error) { return nil, nil }

// TestD13_OrganizationPackageCreatedRecordsActorAndTarget ایجاد بسته را با فاعل و نام بسته ثبت می‌کند.
func TestD13_OrganizationPackageCreatedRecordsActorAndTarget(t *testing.T) {
	enc, err := encryption.NewEncryptor("12345678901234567890123456789012")
	if err != nil {
		t.Fatal(err)
	}
	rec := &testutil.FakeAuditRecorder{}
	svc := &Service{
		repo:       memPackageRepo{},
		audit:      rec,
		encryptSvc: encryption.NewOrganizationPackageEncryptionService(enc),
	}
	id := 3
	_, err = svc.Create(context.Background(), CreateRequest{PackageName: "بسته پایه"}, audit.Actor{
		UserID: &id, Username: "mina", IP: "10.0.0.3",
	})
	if err != nil {
		t.Fatal(err)
	}
	ev := rec.Events[0].Event
	if ev.Type != audit.EventEntityCreated || ev.Target.Type != audit.TargetOrganizationPackage || ev.Target.Name != "بسته پایه" || ev.Target.ID != "4" {
		t.Fatalf("هدف بسته نادرست است: %+v", ev)
	}
	if ev.Actor.UserID == nil || *ev.Actor.UserID != 3 {
		t.Fatalf("فاعل نادرست است: %+v", ev.Actor)
	}
}
