package settings

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/tpdenta/afta-reception/internal/platform/apperror"
	"github.com/tpdenta/afta-reception/internal/platform/security/audit"
	"github.com/tpdenta/afta-reception/internal/platform/security/integrity"
	"gorm.io/gorm"
)

// retiredLastVerifiedEventSetting کلید قدیمی checkpoint در SecuritySettings است.
// R1-24 / FMT_MTD: این نشانگر تنظیم کاربر نیست و تغییر دستی‌اش بررسی لاگ را دور می‌زد.
const retiredLastVerifiedEventSetting = "LastVerifiedEventId"

// Service سرویس مدیریت تنظیمات امنیتی.
type Service struct {
	repo   Repository
	signer *integrity.Signer
	audit  audit.Recorder
}

// NewService نمونه Service می‌سازد.
func NewService(db *gorm.DB, signer *integrity.Signer, auditMgr *audit.Manager) *Service {
	return &Service{
		repo:   NewRepository(db),
		signer: signer,
		audit:  auditMgr,
	}
}

// GetSettingValue مقدار یک تنظیم را برمی‌گرداند؛ در صورت نبود از پیش‌فرض استفاده می‌کند.
func (s *Service) GetSettingValue(name string) (string, error) {
	setting, err := s.repo.FindByName(name)
	if err == gorm.ErrRecordNotFound {
		if def, ok := DefaultSettings[name]; ok {
			return def, nil
		}
		return "", apperror.New("SETTING_NOT_FOUND", "تنظیم مورد نظر یافت نشد.", fmt.Sprintf("setting %s not found", name), 404)
	}
	if err != nil {
		return "", apperror.New("DB_ERROR", "خطا در خواندن تنظیمات.", err.Error(), 500)
	}

	if !s.signer.Verify(setting.IntegrityHash, setting.SettingName, setting.SettingValue) {
		if err := s.recordIntegrity(name); err != nil {
			return "", err
		}
		return "", apperror.ErrIntegrity
	}

	return setting.SettingValue, nil
}

// GetSettingInt مقدار عددی تنظیم را برمی‌گرداند.
func (s *Service) GetSettingInt(name string) (int, error) {
	val, err := s.GetSettingValue(name)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(val)
}

// UpdateSettingByID مقدار یک تنظیم را با شناسه ردیف به‌روز می‌کند و مقدار قبل و بعد را ثبت می‌کند.
// ctx لغو درخواست است. id شناسه ردیف، value مقدار جدید و actor کاربر عامل است.
// در خطا، خطای برنامه را برمی‌گرداند. مقدار تنظیم در stdout نوشته نمی‌شود (D-09).
// FMT_MTD / R1-17: رویداد SETTING_CHANGED هر دو مقدار را دارد و نام هدف برچسب فارسی کلید است.
func (s *Service) UpdateSettingByID(ctx context.Context, id int, value string, actor audit.Actor) error {
	setting, err := s.repo.FindByID(id)
	if err != nil {
		return apperror.New("DB_ERROR", "خطا در خواندن تنظیمات.", err.Error(), 500)
	}
	before := setting.SettingValue
	key := setting.SettingName
	next := &SecuritySetting{
		SettingName:     key,
		SettingValue:    value,
		IntegrityHash:   s.signer.Sign(key, value),
		UpdatedAt:       time.Now().UTC(),
		UpdatedByUserID: actorUserID(actor),
	}
	if err := s.repo.Upsert(next); err != nil {
		log.Println("خطا در ذخیره تنظیمات:", key, err)
		return apperror.New("DB_ERROR", "خطا در ذخیره تنظیمات.", err.Error(), 500)
	}
	return s.recordSettingChange(ctx, actor, key, before, value)
}

// UpdateSetting مقدار تنظیم را به‌روزرسانی و رویداد قبل/بعد ثبت می‌کند.
// ctx لغو درخواست است. name کلید، value مقدار جدید و actor فاعل است.
// مقدار قبلی قبل از نوشتن خوانده می‌شود. خطا یعنی ذخیره یا ممیزی انجام نشده است.
func (s *Service) UpdateSetting(ctx context.Context, name, value string, actor audit.Actor) error {
	before := ""
	existing, err := s.repo.FindByName(name)
	if err != nil && err != gorm.ErrRecordNotFound {
		return apperror.New("DB_ERROR", "خطا در خواندن تنظیمات.", err.Error(), 500)
	}
	if existing != nil {
		before = existing.SettingValue
	}
	setting := &SecuritySetting{
		SettingName:     name,
		SettingValue:    value,
		IntegrityHash:   s.signer.Sign(name, value),
		UpdatedAt:       time.Now().UTC(),
		UpdatedByUserID: actorUserID(actor),
	}
	if err := s.repo.Upsert(setting); err != nil {
		return apperror.New("DB_ERROR", "خطا در ذخیره تنظیمات.", err.Error(), 500)
	}
	return s.recordSettingChange(ctx, actor, name, before, value)
}

// SeedDefaults تنظیمات پیش‌فرض را در صورت نبود درج می‌کند.
func (s *Service) SeedDefaults() error {
	for name, value := range DefaultSettings {
		_, err := s.repo.FindByName(name)
		if err == gorm.ErrRecordNotFound {
			hash := s.signer.Sign(name, value)
			if err := s.repo.Upsert(&SecuritySetting{
				SettingName:   name,
				SettingValue:  value,
				IntegrityHash: hash,
				UpdatedAt:     time.Now().UTC(),
			}); err != nil {
				return err
			}
		}
	}
	return nil
}

// GetAllSettings تمام تنظیمات را برمی‌گرداند.
func (s *Service) GetAllSettings(IsFn bool) ([]SecuritySetting, error) {
	return s.repo.FindAll(IsFn)
}

// InvalidSettings تنظیماتی را برمی‌گرداند که HMAC آن‌ها با نام و مقدار نمی‌خواند.
// خروجی همه ردیف‌های نامعتبر است. حلقه تا آخر ادامه دارد و اینجا رویدادی ثبت نمی‌شود.
// هشدار دوره‌ای در audit.ApplyIntegrityFindings یک بار برای هر ردیف ساخته می‌شود.
func (s *Service) InvalidSettings() ([]audit.IntegritySubject, error) {
	list, err := s.repo.FindAll(false)
	if err != nil {
		return nil, err
	}
	bad := make([]audit.IntegritySubject, 0)
	for _, setting := range list {
		if s.signer.Verify(setting.IntegrityHash, setting.SettingName, setting.SettingValue) {
			continue
		}
		bad = append(bad, audit.IntegritySubject{
			EntityType: audit.TargetSetting,
			EntityID:   setting.SettingName,
			EntityName: settingLabel(setting.SettingName),
		})
	}
	return bad, nil
}

// RemoveRetiredChainCheckpoint اگر ردیف LastVerifiedEventId مانده باشد آن را حذف می‌کند و SETTING_REMOVED می‌نویسد.
// خروجی nil یعنی ردیفی نبود، یا حذف و رویداد ممیزی هر دو انجام شد.
// مقدار قدیمی به AuditChainStates کپی نمی‌شود؛ بررسی اول زنجیره از صفر است.
// R1-24 / FMT_MTD: این کلید نباید در صفحه تنظیمات بماند.
func (s *Service) RemoveRetiredChainCheckpoint() error {
	_, err := s.repo.FindByName(retiredLastVerifiedEventSetting)
	if err == gorm.ErrRecordNotFound {
		return nil
	}
	if err != nil {
		return err
	}
	if err := s.repo.DeleteByName(retiredLastVerifiedEventSetting); err != nil {
		return err
	}
	return s.record(context.Background(), audit.Event{
		Type:        audit.EventSettingRemoved,
		Result:      audit.ResultSuccess,
		Actor:       audit.SystemActor(),
		Target:      audit.Target{Type: audit.TargetSetting, ID: retiredLastVerifiedEventSetting, Name: retiredLastVerifiedEventSetting},
		Description: "حذف تنظیم بازنشسته LastVerifiedEventId؛ وضعیت زنجیره از جدول AuditChainStates و از ابتدا بررسی می‌شود",
	})
}

// recordSettingChange رویداد SETTING_CHANGED را با کلید و مقدار قبل و بعد می‌نویسد.
// ctx زمینه درخواست است. actor فاعل، key کلید تنظیم و before/after مقدارها هستند.
// نام هدف برچسب فارسی است. خطا یعنی ردیف ممیزی نوشته نشده است.
func (s *Service) recordSettingChange(ctx context.Context, actor audit.Actor, key, before, after string) error {
	return s.record(ctx, audit.Event{
		Type:   audit.EventSettingChanged,
		Result: audit.ResultSuccess,
		Actor:  actor,
		Target: audit.Target{Type: audit.TargetSetting, ID: key, Name: settingLabel(key)},
		Details: map[string]string{
			"key":    key,
			"before": before,
			"after":  after,
		},
		Description: "تغییر تنظیم «" + settingLabel(key) + "»",
	})
}

// recordIntegrity دستکاری یک تنظیم را ثبت می‌کند.
// name کلید تنظیم است. خروجی خطای ممیزی یا nil است و سپس خواندن باید ErrIntegrity بدهد.
func (s *Service) recordIntegrity(name string) error {
	return s.record(context.Background(), audit.Event{
		Type:        audit.EventDataTampering,
		Result:      audit.ResultFailure,
		Reason:      audit.ReasonIntegrityViolation,
		Actor:       audit.SystemActor(),
		Target:      audit.Target{Type: audit.TargetSetting, ID: name, Name: settingLabel(name)},
		Description: "دستکاری تنظیم امنیتی: " + name,
	})
}

// record رویداد را می‌نویسد و شکست را به AUDIT_UNAVAILABLE تبدیل می‌کند.
// ctx زمینه است. e رویداد است. nil یعنی ثبت شد.
func (s *Service) record(ctx context.Context, e audit.Event) error {
	if s.audit == nil {
		return audit.Unavailable(apperror.ErrInternal)
	}
	return audit.Unavailable(s.audit.Record(ctx, e))
}

// settingLabel برچسب فارسی کلید را برمی‌گرداند و اگر نباشد خود کلید را.
// name کلید تنظیم است. خروجی برای TargetName است.
func settingLabel(name string) string {
	if label, ok := SettingsDicFN[name]; ok {
		return label
	}
	return name
}

// actorUserID شناسه فاعل را برای ستون به‌روزرسانی برمی‌گرداند.
// actor فاعل ممیزی است. فاعل سیستمی یا بدون شناسه nil می‌دهد تا صفر ذخیره نشود.
func actorUserID(actor audit.Actor) *int {
	if actor.UserID == nil || *actor.UserID <= 0 {
		return nil
	}
	id := *actor.UserID
	return &id
}
