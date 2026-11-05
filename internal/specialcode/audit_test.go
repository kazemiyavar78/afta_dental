package specialcode

import (
	"context"
	"testing"

	"github.com/tpdenta/afta-reception/internal/platform/security/audit"
	"github.com/tpdenta/afta-reception/internal/platform/security/integrity"
	"github.com/tpdenta/afta-reception/internal/platform/testutil"
	"gorm.io/gorm"
)

type memSpecialRepo struct{}

func (memSpecialRepo) Create(item *SpecialCode) error {
	item.ID = 5
	return nil
}
func (memSpecialRepo) Update(*SpecialCode) error           { return nil }
func (memSpecialRepo) Delete(*SpecialCode) error           { return nil }
func (memSpecialRepo) FindByID(uint) (*SpecialCode, error) { return nil, gorm.ErrRecordNotFound }
func (memSpecialRepo) FindByCode(string) (*SpecialCode, error) {
	return nil, gorm.ErrRecordNotFound
}
func (memSpecialRepo) FindAll() ([]SpecialCode, error) { return nil, nil }

// TestD13_SpecialCodeCreatedRecordsActorAndTarget ایجاد کد خاص را با فاعل و خود کد به‌عنوان نام هدف ثبت می‌کند.
func TestD13_SpecialCodeCreatedRecordsActorAndTarget(t *testing.T) {
	rec := &testutil.FakeAuditRecorder{}
	svc := &Service{
		repo:   memSpecialRepo{},
		audit:  rec,
		signer: integrity.NewSigner("test-integrity-key-32bytes-long!!"),
	}
	id := 7
	_, err := svc.Create(context.Background(), CreateRequest{Code: "SC-9", Name: "جانباز"}, audit.Actor{
		UserID: &id, Username: "ali", IP: "10.0.0.7",
	})
	if err != nil {
		t.Fatal(err)
	}
	ev := rec.Events[0].Event
	if ev.Type != audit.EventEntityCreated || ev.Target.Type != audit.TargetSpecialCode || ev.Target.Name != "SC-9" || ev.Target.ID != "5" {
		t.Fatalf("هدف کد خاص نادرست است: %+v", ev)
	}
	if ev.Actor.Username != "ali" || ev.Actor.UserID == nil || *ev.Actor.UserID != 7 {
		t.Fatalf("فاعل نادرست است: %+v", ev.Actor)
	}
}
