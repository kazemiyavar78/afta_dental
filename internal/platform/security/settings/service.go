package settings

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/tpdenta/afta-reception/internal/platform/apperror"
	"github.com/tpdenta/afta-reception/internal/platform/security/audit"
	"github.com/tpdenta/afta-reception/internal/platform/security/integrity"
	"gorm.io/gorm"
)

// Service سرویس مدیریت تنظیمات امنیتی.
type Service struct {
	repo    Repository
	signer  *integrity.Signer
	audit   *audit.Manager
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
		_ = s.audit.LogEvent(nil, "", audit.EventDataTampering, fmt.Sprintf("دستکاری تنظیم امنیتی: %s", name))
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

// UpdateSetting مقدار تنظیم را به‌روزرسانی و رویداد امنیتی ثبت می‌کند.
func (s *Service) UpdateSettingByID(id int, value string, actorUserID int, ip string) error {
	setting, err := s.repo.FindByID(id)
	if err != nil {
		return apperror.New("DB_ERROR", "خطا در خواندن تنظیمات.", err.Error(), 500)
	}
	fmt.Println(setting.SettingName, value)
	hash := s.signer.Sign(setting.SettingName, value)
	setting = &SecuritySetting{
		SettingName:     setting.SettingName,
		SettingValue:    value,
		IntegrityHash:   hash,
		UpdatedAt:       time.Now().UTC(),
		UpdatedByUserID: &actorUserID,
	}

	if err := s.repo.Upsert(setting); err != nil {
		log.Println("خطا در ذخیره تنظیمات:", err)
		return apperror.New("DB_ERROR", "خطا در ذخیره تنظیمات.", err.Error(), 500)
	}

	uid := actorUserID
	_ = s.audit.LogEvent(&uid, ip, audit.EventSecuritySettingChange,
		fmt.Sprintf("تغییر تنظیم %s به %s", setting.SettingName, value))

	return nil
}

// UpdateSetting مقدار تنظیم را به‌روزرسانی و رویداد امنیتی ثبت می‌کند.
func (s *Service) UpdateSetting(name, value string, actorUserID int, ip string) error {
	hash := s.signer.Sign(name, value)
	setting := &SecuritySetting{
		SettingName:     name,
		SettingValue:    value,
		IntegrityHash:   hash,
		UpdatedAt:       time.Now().UTC(),
		UpdatedByUserID: &actorUserID,
	}

	if err := s.repo.Upsert(setting); err != nil {
		return apperror.New("DB_ERROR", "خطا در ذخیره تنظیمات.", err.Error(), 500)
	}

	uid := actorUserID
	_ = s.audit.LogEvent(&uid, ip, audit.EventSecuritySettingChange,
		fmt.Sprintf("تغییر تنظیم %s به %s", name, value))

	return nil
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

// VerifyAllIntegrity یکپارچگی تمام تنظیمات را بررسی می‌کند.
func (s *Service) VerifyAllIntegrity(ip string) (bool, error) {
	list, err := s.repo.FindAll(false)
	if err != nil {
		return false, err
	}
	for _, setting := range list {
		if !s.signer.Verify(setting.IntegrityHash, setting.SettingName, setting.SettingValue) {
			_ = s.audit.LogEvent(nil, ip, audit.EventDataTampering,
				fmt.Sprintf("دستکاری تنظیم امنیتی: %s", setting.SettingName))
			return false, nil
		}
	}
	return true, nil
}
