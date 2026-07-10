package settings

import (
	"time"

	"gorm.io/gorm"
)

// SecuritySetting مدل جدول تنظیمات امنیتی.
type SecuritySetting struct {
	ID              int       `gorm:"column:ID;primaryKey;autoIncrement"`
	SettingName     string    `gorm:"column:SettingName;uniqueIndex;size:255;not null"`
	SettingValue    string    `gorm:"column:SettingValue;type:nvarchar(max);not null"`
	IntegrityHash   string    `gorm:"column:IntegrityHash;size:128;not null"`
	UpdatedAt       time.Time `gorm:"column:UpdatedAt;not null"`
	UpdatedByUserID *int      `gorm:"column:UpdatedByUserID"`
}

// TableName نام جدول در دیتابیس.
func (SecuritySetting) TableName() string {
	return "SecuritySettings"
}

// Repository عملیات CRUD خام تنظیمات امنیتی.
type Repository interface {
	FindByName(name string) (*SecuritySetting, error)
	FindAll(IsFn bool) ([]SecuritySetting, error)
	Upsert(setting *SecuritySetting) error
	FindByID(id int) (*SecuritySetting, error)
}

type gormRepository struct {
	db *gorm.DB
}

// NewRepository نمونه Repository می‌سازد.
func NewRepository(db *gorm.DB) Repository {
	return &gormRepository{db: db}
}

func (r *gormRepository) FindByName(name string) (*SecuritySetting, error) {
	var s SecuritySetting
	err := r.db.Where("SettingName = ?", name).First(&s).Error
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *gormRepository) FindByID(id int) (*SecuritySetting, error) {
	var s SecuritySetting
	err := r.db.Where("ID = ?", id).First(&s).Error
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *gormRepository) FindAll(IsFn bool) ([]SecuritySetting, error) {
	var list []SecuritySetting
	err := r.db.Find(&list).Error
	if err != nil {
		return list, err
	}

	if IsFn {
		for index, val := range list {
			if key, ok := SettingsDicFN[val.SettingName]; ok {
				list[index].SettingName = key
			} else {
				if index == len(list)-1 {
					list = list[:index]
				} else {
					list = append(list[:index], list[index+1:]...)
				}
			}
		}
	}
	return list, err
}

func (r *gormRepository) Upsert(setting *SecuritySetting) error {
	var existing SecuritySetting
	err := r.db.Where("SettingName = ?", setting.SettingName).First(&existing).Error
	if err == gorm.ErrRecordNotFound {
		return r.db.Create(setting).Error
	}
	if err != nil {
		return err
	}
	setting.ID = existing.ID
	return r.db.Save(setting).Error
}
