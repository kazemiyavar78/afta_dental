package organization

import (
	"context"
	"testing"

	"github.com/tpdenta/afta-reception/internal/platform/security/audit"
	"github.com/tpdenta/afta-reception/internal/platform/security/encryption"
	"github.com/tpdenta/afta-reception/internal/platform/testutil"
	"gorm.io/gorm"
)

type memOrgRepo struct{}

func (memOrgRepo) Create(o *Organization) error {
	o.ID = 8
	return nil
}
func (memOrgRepo) Update(*Organization) error           { return nil }
func (memOrgRepo) Delete(*Organization) error           { return nil }
func (memOrgRepo) FindByID(uint) (*Organization, error) { return nil, gorm.ErrRecordNotFound }
func (memOrgRepo) FindAll() ([]Organization, error)     { return nil, nil }
func (memOrgRepo) GetActive() ([]Organization, error)   { return nil, nil }
func (memOrgRepo) FindFree() (*Organization, error)     { return nil, gorm.ErrRecordNotFound }

type yesPackage struct{}

func (yesPackage) Exists(int) (bool, error) { return true, nil }

// TestD13_OrganizationCreatedRecordsActorAndTarget ایجاد سازمان را با فاعل و نام سازمان ثبت می‌کند.
func TestD13_OrganizationCreatedRecordsActorAndTarget(t *testing.T) {
	enc, err := encryption.NewEncryptor("12345678901234567890123456789012")
	if err != nil {
		t.Fatal(err)
	}
	rec := &testutil.FakeAuditRecorder{}
	svc := &Service{
		repo:       memOrgRepo{},
		audit:      rec,
		encryptSvc: encryption.NewOrganizationEncryptionService(enc),
		packageSvc: yesPackage{},
	}
	id := 1
	_, err = svc.Create(context.Background(), CreateRequest{
		Name: "تأمین اجتماعی", PackageID: 1, CenterPackageID: 1,
	}, audit.Actor{UserID: &id, Username: "admin", IP: "10.0.0.1"})
	if err != nil {
		t.Fatal(err)
	}
	ev := rec.Events[0].Event
	if ev.Type != audit.EventEntityCreated || ev.Target.Type != audit.TargetOrganization || ev.Target.Name != "تأمین اجتماعی" || ev.Target.ID != "8" {
		t.Fatalf("هدف سازمان نادرست است: %+v", ev)
	}
	if ev.Actor.UserID == nil || *ev.Actor.UserID != 1 || ev.Actor.Username != "admin" {
		t.Fatalf("فاعل نادرست است: %+v", ev.Actor)
	}
}
