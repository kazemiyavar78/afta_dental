package patient

import (
	"context"
	"testing"

	"github.com/tpdenta/afta-reception/internal/platform/security/audit"
	"github.com/tpdenta/afta-reception/internal/platform/security/encryption"
	"github.com/tpdenta/afta-reception/internal/platform/testutil"
	"gorm.io/gorm"
)

type memPatientRepo struct {
	p *Patient
}

func (r *memPatientRepo) Create(p *Patient) error {
	p.ID = 9
	cp := *p
	r.p = &cp
	return nil
}

func (r *memPatientRepo) Update(p *Patient) error {
	cp := *p
	r.p = &cp
	return nil
}

func (r *memPatientRepo) Delete(p *Patient) error { return nil }

func (r *memPatientRepo) FindByID(id uint) (*Patient, error) {
	if r.p == nil || r.p.ID != id {
		return nil, gorm.ErrRecordNotFound
	}
	cp := *r.p
	return &cp, nil
}

func (r *memPatientRepo) FindAll() ([]Patient, error)            { return nil, nil }
func (r *memPatientRepo) Search(SearchFilter) ([]Patient, error) { return nil, nil }
func (r *memPatientRepo) FindByNationalCode(string) (*Patient, error) {
	return nil, gorm.ErrRecordNotFound
}
func (r *memPatientRepo) FindByFileNumber(string) (*Patient, error) {
	return nil, gorm.ErrRecordNotFound
}
func (r *memPatientRepo) FindByFirstNameAndLastName(string, string) (*Patient, error) {
	return nil, gorm.ErrRecordNotFound
}
func (r *memPatientRepo) FindLatestFileNumber() (string, error) { return "", nil }

func testPatientService(t *testing.T, rec *testutil.FakeAuditRecorder) *Service {
	t.Helper()
	enc, err := encryption.NewEncryptor("12345678901234567890123456789012")
	if err != nil {
		t.Fatal(err)
	}
	return &Service{
		repo:       &memPatientRepo{},
		audit:      rec,
		encryptSvc: encryption.NewPatientEncryptionService(enc),
	}
}

// TestD13_PatientCreatedRecordsActorAndTarget ایجاد بیمار را با فاعل و نام نمایشی ثبت می‌کند.
func TestD13_PatientCreatedRecordsActorAndTarget(t *testing.T) {
	rec := &testutil.FakeAuditRecorder{}
	svc := testPatientService(t, rec)
	id := 4
	actor := audit.Actor{UserID: &id, Username: "admin", IP: "10.0.0.4"}
	phone := "09121234567"
	_, err := svc.Create(context.Background(), CreateRequest{
		FirstName: "علی", LastName: "رضایی", FileNumber: "F1", BirthDate: "2000-01-02",
		IsForeignNational: true, MobilePhoneNumber: &phone,
	}, actor)
	if err != nil {
		t.Fatal(err)
	}
	if len(rec.Events) != 1 {
		t.Fatalf("یک رویداد ایجاد انتظار می‌رود، دریافت %d", len(rec.Events))
	}
	ev := rec.Events[0].Event
	if ev.Type != audit.EventEntityCreated || ev.Target.Type != audit.TargetPatient {
		t.Fatalf("نوع رویداد نادرست است: %+v", ev)
	}
	if ev.Actor.UserID == nil || *ev.Actor.UserID != 4 || ev.Actor.Username != "admin" {
		t.Fatalf("فاعل نادرست است: %+v", ev.Actor)
	}
	if ev.Target.Name != "علی ر" || ev.Target.ID != "9" {
		t.Fatalf("هدف باید نام نمایشی باشد: %+v", ev.Target)
	}
	if ev.Target.Name == "ATBA-F1" {
		t.Fatal("نام هدف نباید کد شناسایی باشد")
	}
}

// TestD13_PatientUpdatedMasksNationalCodeAndPhone به‌روزرسانی فقط اختلاف را با کد ملی و تلفن ماسک‌شده می‌نویسد.
func TestD13_PatientUpdatedMasksNationalCodeAndPhone(t *testing.T) {
	rec := &testutil.FakeAuditRecorder{}
	svc := testPatientService(t, rec)
	id := 4
	actor := audit.Actor{UserID: &id, Username: "admin", IP: "10.0.0.4"}
	oldPhone := "09121111111"
	created, err := svc.Create(context.Background(), CreateRequest{
		FirstName: "علی", LastName: "رضایی", FileNumber: "F1", BirthDate: "2000-01-02",
		IsForeignNational: true, MobilePhoneNumber: &oldPhone,
	}, actor)
	if err != nil {
		t.Fatal(err)
	}
	newPhone := "09122222222"
	_, err = svc.Update(context.Background(), created.ID, UpdateRequest{
		FirstName: "علی", LastName: "رضایی", FileNumber: "F2", BirthDate: "2000-01-02",
		IsForeignNational: true, MobilePhoneNumber: &newPhone,
	}, actor)
	if err != nil {
		t.Fatal(err)
	}
	if len(rec.Events) != 2 {
		t.Fatalf("ایجاد و به‌روزرسانی انتظار می‌رود، دریافت %d", len(rec.Events))
	}
	details, ok := rec.Events[1].Event.Details.(map[string]audit.Change)
	if !ok {
		t.Fatalf("جزئیات Diff نیست: %#v", rec.Events[1].Event.Details)
	}
	phone, ok := details["mobile_phone"]
	if !ok || phone.Before == oldPhone || phone.After == newPhone {
		t.Fatalf("تلفن باید ماسک شود: %+v", phone)
	}
	code, ok := details["national_code"]
	if !ok || code.Before == "ATBA-F1" || code.After == "ATBA-F2" {
		t.Fatalf("کد شناسایی باید ماسک شود: %+v", code)
	}
}
