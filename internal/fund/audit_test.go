package fund

import (
	"context"
	"testing"

	"github.com/tpdenta/afta-reception/internal/platform/security/audit"
	"github.com/tpdenta/afta-reception/internal/platform/testutil"
	"gorm.io/gorm"
)

type memFundRepo struct{}

func (memFundRepo) Create(f *Fund) error {
	f.ID = 3
	return nil
}
func (memFundRepo) FindByID(int) (*Fund, error) { return nil, gorm.ErrRecordNotFound }
func (memFundRepo) FindAll() ([]Fund, error)    { return nil, nil }

// TestD13_FundCreatedRecordsActorAndTarget ایجاد صندوق را با فاعل و نام صندوق ثبت می‌کند.
func TestD13_FundCreatedRecordsActorAndTarget(t *testing.T) {
	rec := &testutil.FakeAuditRecorder{}
	svc := &Service{repo: memFundRepo{}, audit: rec}
	id := 6
	_, err := svc.Create(context.Background(), CreateRequest{Name: "صندوق نقد"}, audit.Actor{
		UserID: &id, Username: "sara", IP: "10.0.0.6",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(rec.Events) != 1 {
		t.Fatalf("یک رویداد انتظار می‌رود، دریافت %d", len(rec.Events))
	}
	ev := rec.Events[0].Event
	if ev.Type != audit.EventEntityCreated || ev.Target.Type != audit.TargetFund || ev.Target.ID != "3" || ev.Target.Name != "صندوق نقد" {
		t.Fatalf("هدف صندوق نادرست است: %+v", ev.Target)
	}
	if ev.Actor.UserID == nil || *ev.Actor.UserID != 6 || ev.Actor.Username != "sara" {
		t.Fatalf("فاعل نادرست است: %+v", ev.Actor)
	}
}
