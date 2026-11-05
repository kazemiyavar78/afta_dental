package settings

import (
	"context"
	"strings"
	"testing"

	"github.com/tpdenta/afta-reception/internal/platform/security/audit"
	"github.com/tpdenta/afta-reception/internal/platform/security/integrity"
	"gorm.io/gorm"
)

// memAudit ثبت ممیزی را در حافظه نگه می‌دارد.
// این بسته نمی‌تواند testutil را وارد کند چون testutil به session و session به settings وابسته است.
type memAudit struct {
	events []audit.Event
	err    error
}

// Record رویداد را ذخیره می‌کند. ctx نادیده گرفته می‌شود مگر لغو شده باشد. e رویداد است.
func (m *memAudit) Record(ctx context.Context, e audit.Event) error {
	if ctx != nil {
		if err := ctx.Err(); err != nil {
			return err
		}
	}
	if m.err != nil {
		return m.err
	}
	m.events = append(m.events, e)
	return nil
}

// RecordTx همان Record است. tx در این تست باز نمی‌شود.
func (m *memAudit) RecordTx(tx *gorm.DB, e audit.Event) error {
	return m.Record(context.Background(), e)
}

// TestSigner_SignSettings مسیر خوشحال امضای تنظیمات را بررسی می‌کند.
func TestSigner_SignSettings(t *testing.T) {
	s := integrity.NewSigner("test-integrity-key-32bytes-long!!")
	hash := s.Sign(MaximumNumberOfFailedLogin, "5")
	if !s.Verify(hash, MaximumNumberOfFailedLogin, "5") {
		t.Fatal("امضای تنظیمات باید معتبر باشد")
	}
}

// TestR1_24_LastVerifiedEventIdIsNotAUserSetting کلید checkpoint را از پیش‌فرض و برچسب UI حذف‌شده می‌داند.
func TestR1_24_LastVerifiedEventIdIsNotAUserSetting(t *testing.T) {
	if _, ok := DefaultSettings[retiredLastVerifiedEventSetting]; ok {
		t.Fatal("کلید checkpoint نباید در پیش‌فرض تنظیمات باشد")
	}
	if _, ok := SettingsDicFN[retiredLastVerifiedEventSetting]; ok {
		t.Fatal("کلید checkpoint نباید در صفحه تنظیمات باشد")
	}
}

type memSettingsRepo struct {
	rows   map[string]*SecuritySetting
	listed []SecuritySetting
}

func (m *memSettingsRepo) FindByName(name string) (*SecuritySetting, error) {
	row, ok := m.rows[name]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	return row, nil
}

func (m *memSettingsRepo) FindAll(bool) ([]SecuritySetting, error) {
	if m.listed != nil {
		return append([]SecuritySetting(nil), m.listed...), nil
	}
	return nil, nil
}

func (m *memSettingsRepo) Upsert(*SecuritySetting) error { return nil }

func (m *memSettingsRepo) FindByID(id int) (*SecuritySetting, error) {
	for _, row := range m.rows {
		if row.ID == id {
			cp := *row
			return &cp, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

func (m *memSettingsRepo) DeleteByName(name string) error {
	delete(m.rows, name)
	return nil
}

// TestR1_24_RemoveRetiredCheckpointAuditsOnce ردیف قدیمی را حذف می‌کند، یک SETTING_REMOVED می‌نویسد و بار دوم ساکت است.
func TestR1_24_RemoveRetiredCheckpointAuditsOnce(t *testing.T) {
	repo := &memSettingsRepo{rows: map[string]*SecuritySetting{
		retiredLastVerifiedEventSetting: {SettingName: retiredLastVerifiedEventSetting, SettingValue: "6"},
	}}
	aud := &memAudit{}
	svc := &Service{repo: repo, audit: aud}
	if err := svc.RemoveRetiredChainCheckpoint(); err != nil {
		t.Fatalf("حذف اول: %v", err)
	}
	if _, ok := repo.rows[retiredLastVerifiedEventSetting]; ok {
		t.Fatal("ردیف checkpoint باید حذف شود")
	}
	if err := svc.RemoveRetiredChainCheckpoint(); err != nil {
		t.Fatalf("حذف دوم: %v", err)
	}
	if len(aud.events) != 1 {
		t.Fatalf("باید یک رویداد حذف ثبت شود، %d ثبت شد", len(aud.events))
	}
	got := aud.events[0]
	if got.Type != audit.EventSettingRemoved {
		t.Fatalf("نوع رویداد نادرست است: %s", got.Type)
	}
	if strings.Contains(got.Description, "SettingValue") || got.Actor.Username != "system" {
		t.Fatalf("رویداد حذف نادرست است: %+v", got)
	}
}

// TestR1_17_SettingChangedRecordsBeforeAndAfter مقدار قبلی را قبل از نوشتن می‌خواند و برچسب فارسی را هدف می‌گذارد.
func TestR1_17_SettingChangedRecordsBeforeAndAfter(t *testing.T) {
	repo := &memSettingsRepo{rows: map[string]*SecuritySetting{
		MaximumNumberOfFailedLogin: {
			ID:           4,
			SettingName:  MaximumNumberOfFailedLogin,
			SettingValue: "5",
		},
	}}
	aud := &memAudit{}
	svc := &Service{
		repo:   repo,
		signer: integrity.NewSigner("test-integrity-key-32bytes-long!!"),
		audit:  aud,
	}
	actorID := 3
	actor := audit.Actor{UserID: &actorID, Username: "sara", IP: "198.51.100.8"}
	if err := svc.UpdateSettingByID(context.Background(), 4, "7", actor); err != nil {
		t.Fatalf("به‌روزرسانی: %v", err)
	}
	if len(aud.events) != 1 {
		t.Fatalf("باید یک SETTING_CHANGED باشد، %d ثبت شد", len(aud.events))
	}
	got := aud.events[0]
	if got.Type != audit.EventSettingChanged || got.Actor.Username != "sara" {
		t.Fatalf("فاعل یا نوع نادرست است: %+v", got)
	}
	if got.Target.Name != SettingsDicFN[MaximumNumberOfFailedLogin] || got.Target.Type != audit.TargetSetting {
		t.Fatalf("هدف باید برچسب فارسی تنظیم باشد: %+v", got.Target)
	}
	details, ok := got.Details.(map[string]string)
	if !ok || details["key"] != MaximumNumberOfFailedLogin || details["before"] != "5" || details["after"] != "7" {
		t.Fatalf("جزئیات قبل و بعد نادرست است: %#v", got.Details)
	}
}

// TestD23_InvalidSettings_ReturnsEveryBrokenRow هر دو تنظیم خراب را برمی‌گرداند و رویداد نمی‌سازد.
// تنظیم سالم وسط فهرست نباید بررسی را قطع کند.
func TestD23_InvalidSettings_ReturnsEveryBrokenRow(t *testing.T) {
	signer := integrity.NewSigner("test-integrity-key-32bytes-long!!")
	ok := SecuritySetting{SettingName: "Password_MinLength", SettingValue: "8"}
	ok.IntegrityHash = signer.Sign(ok.SettingName, ok.SettingValue)
	repo := &memSettingsRepo{listed: []SecuritySetting{
		{SettingName: "A", SettingValue: "1", IntegrityHash: "bad"},
		ok,
		{SettingName: "B", SettingValue: "2", IntegrityHash: "bad"},
	}}
	aud := &memAudit{}
	got, err := (&Service{repo: repo, signer: signer, audit: aud}).InvalidSettings()
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0].EntityID != "A" || got[1].EntityID != "B" || got[0].EntityType != audit.TargetSetting {
		t.Fatalf("هر دو تنظیم خراب باید برگردند: %+v", got)
	}
	if len(aud.events) != 0 {
		t.Fatal("سرویس نباید رویداد بسازد")
	}
}
