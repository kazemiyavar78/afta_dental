package regulation

import (
	"context"
	"testing"

	"github.com/tpdenta/afta-reception/internal/platform/security/audit"
	"github.com/tpdenta/afta-reception/internal/platform/testutil"
	"gorm.io/gorm"
)

type memRegulationRepo struct{}

func (memRegulationRepo) Create(item *Regulation) error {
	item.ID = 11
	return nil
}
func (memRegulationRepo) Update(*Regulation) error           { return nil }
func (memRegulationRepo) Delete(*Regulation) error           { return nil }
func (memRegulationRepo) ReplaceServices(uint, []uint) error { return nil }
func (memRegulationRepo) FindByID(uint) (*Regulation, error) { return nil, gorm.ErrRecordNotFound }
func (memRegulationRepo) FindAll() ([]Regulation, error)     { return nil, nil }
func (memRegulationRepo) FindActive() ([]Regulation, error)  { return nil, nil }

// TestD13_RegulationCreatedRecordsActorAndTarget ایجاد ضابطه را با فاعل و توضیح به‌عنوان نام هدف ثبت می‌کند.
func TestD13_RegulationCreatedRecordsActorAndTarget(t *testing.T) {
	rec := &testutil.FakeAuditRecorder{}
	svc := &Service{repo: memRegulationRepo{}, audit: rec}
	id := 2
	_, err := svc.Create(context.Background(), CreateRequest{
		ServiceIDs: []uint{1}, DurationDays: 30, Description: "عکس بعد از درمان",
	}, audit.Actor{UserID: &id, Username: "neda", IP: "10.0.0.2"})
	if err != nil {
		t.Fatal(err)
	}
	ev := rec.Events[0].Event
	if ev.Type != audit.EventEntityCreated || ev.Target.Type != audit.TargetRegulation || ev.Target.Name != "عکس بعد از درمان" || ev.Target.ID != "11" {
		t.Fatalf("هدف ضابطه نادرست است: %+v", ev)
	}
	if ev.Actor.UserID == nil || *ev.Actor.UserID != 2 {
		t.Fatalf("فاعل نادرست است: %+v", ev.Actor)
	}
}
