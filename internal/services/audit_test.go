package services

import (
	"context"
	"testing"

	"github.com/tpdenta/afta-reception/internal/platform/security/audit"
	"github.com/tpdenta/afta-reception/internal/platform/security/encryption"
	"github.com/tpdenta/afta-reception/internal/platform/testutil"
	"gorm.io/gorm"
)

type memServiceRepo struct{}

func (memServiceRepo) Create(item *ServiceItem) error {
	item.ID = 15
	return nil
}
func (memServiceRepo) Update(*ServiceItem) error           { return nil }
func (memServiceRepo) Delete(*ServiceItem) error           { return nil }
func (memServiceRepo) FindByID(uint) (*ServiceItem, error) { return nil, gorm.ErrRecordNotFound }
func (memServiceRepo) FindByServiceCode(string) (*ServiceItem, error) {
	return nil, gorm.ErrRecordNotFound
}
func (memServiceRepo) FindAll() ([]ServiceItem, error) { return nil, nil }
func (memServiceRepo) FindByExcludeServices([]uint) ([]ServiceItem, error) {
	return nil, nil
}
func (memServiceRepo) CreateExcludedService(*ExcludedService) error  { return nil }
func (memServiceRepo) RestoreExcludedService(*ExcludedService) error { return nil }
func (memServiceRepo) RemoveExcludedService(uint, uint) error        { return nil }
func (memServiceRepo) FindByOrganizationID(uint) ([]ExcludedService, error) {
	return nil, nil
}
func (memServiceRepo) FindByOrganizationIDAndServiceID(uint, uint) (*ExcludedService, error) {
	return nil, gorm.ErrRecordNotFound
}
func (memServiceRepo) FindExcludedIncludingDeleted(uint, uint) (*ExcludedService, error) {
	return nil, gorm.ErrRecordNotFound
}

// TestD13_ServiceCreatedRecordsActorAndTarget ایجاد خدمت را با فاعل و کد خدمت به‌عنوان نام هدف ثبت می‌کند.
func TestD13_ServiceCreatedRecordsActorAndTarget(t *testing.T) {
	enc, err := encryption.NewEncryptor("12345678901234567890123456789012")
	if err != nil {
		t.Fatal(err)
	}
	rec := &testutil.FakeAuditRecorder{}
	svc := &Service{
		repo:       memServiceRepo{},
		audit:      rec,
		encryptSvc: encryption.NewServiceEncryptionService(enc),
	}
	id := 9
	_, err = svc.Create(context.Background(), CreateRequest{
		ServiceCode: "D0120", Name: "معاینه",
	}, audit.Actor{UserID: &id, Username: "dr", IP: "10.0.0.9"})
	if err != nil {
		t.Fatal(err)
	}
	ev := rec.Events[0].Event
	if ev.Type != audit.EventEntityCreated || ev.Target.Type != audit.TargetService || ev.Target.Name != "D0120" || ev.Target.ID != "15" {
		t.Fatalf("هدف خدمت باید کد باشد نه نام: %+v", ev.Target)
	}
	if ev.Actor.UserID == nil || *ev.Actor.UserID != 9 {
		t.Fatalf("فاعل نادرست است: %+v", ev.Actor)
	}
}
