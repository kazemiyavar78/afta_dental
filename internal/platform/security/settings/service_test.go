package settings

import (
	"strings"
	"testing"

	"github.com/tpdenta/afta-reception/internal/platform/security/audit"
	"github.com/tpdenta/afta-reception/internal/platform/security/integrity"
	"gorm.io/gorm"
)

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
	rows map[string]*SecuritySetting
}

func (m *memSettingsRepo) FindByName(name string) (*SecuritySetting, error) {
	row, ok := m.rows[name]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	return row, nil
}

func (m *memSettingsRepo) FindAll(bool) ([]SecuritySetting, error) { return nil, nil }

func (m *memSettingsRepo) Upsert(*SecuritySetting) error { return nil }

func (m *memSettingsRepo) FindByID(int) (*SecuritySetting, error) {
	return nil, gorm.ErrRecordNotFound
}

func (m *memSettingsRepo) DeleteByName(name string) error {
	delete(m.rows, name)
	return nil
}

type recordingAuditor struct {
	events []string
}

func (r *recordingAuditor) LogEvent(userID *int, ip, eventType, description string) error {
	r.events = append(r.events, eventType+"|"+description)
	return nil
}

// TestR1_24_RemoveRetiredCheckpointAuditsOnce ردیف قدیمی را حذف می‌کند، یک SETTING_REMOVED می‌نویسد و بار دوم ساکت است.
func TestR1_24_RemoveRetiredCheckpointAuditsOnce(t *testing.T) {
	repo := &memSettingsRepo{rows: map[string]*SecuritySetting{
		retiredLastVerifiedEventSetting: {SettingName: retiredLastVerifiedEventSetting, SettingValue: "6"},
	}}
	aud := &recordingAuditor{}
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
	if !strings.HasPrefix(aud.events[0], audit.EventSettingRemoved+"|") {
		t.Fatalf("نوع رویداد نادرست است: %s", aud.events[0])
	}
	if strings.Contains(aud.events[0], "SettingValue") {
		t.Fatal("مقدار تنظیم نباید در توضیح رویداد باشد")
	}
}
