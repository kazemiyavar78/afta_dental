package regulation

import (
	"gorm.io/gorm"
)

// Regulation مدل ضوابط خدمات.
type Regulation struct {
	gorm.Model
	// مدت زمان به روز
	DurationDays int `gorm:"column:DurationDays;not null;default:30"`
	// وضعیت فعال بودن
	IsActive bool `gorm:"column:IsActive;not null;default:true"`
	// تعداد عکس مورد نیاز در صورت نقض ضابطه
	PhotoCount int `gorm:"column:PhotoCount;not null;default:1"`
	// توضیحات ضابطه
	Description string `gorm:"column:Description;size:500;not null;default:''"`
	// لیست خدمات مرتبط
	Services []RegulationService `gorm:"foreignKey:RegulationID;references:ID"`
}

// TableName نام جدول ضوابط.
func (Regulation) TableName() string { return "Regulations" }

// RegulationService پیوند خدمت با ضابطه.
type RegulationService struct {
	gorm.Model
	RegulationID uint `gorm:"column:RegulationID;not null;index"`
	ServiceID    uint `gorm:"column:ServiceID;not null;index"`
}

// TableName نام جدول پیوند خدمات ضابطه.
func (RegulationService) TableName() string { return "RegulationServices" }

// Repository اینترفیس CRUD ضوابط.
type Repository interface {
	Create(item *Regulation) error
	Update(item *Regulation) error
	Delete(item *Regulation) error
	ReplaceServices(regulationID uint, serviceIDs []uint) error
	FindByID(id uint) (*Regulation, error)
	FindAll() ([]Regulation, error)
	FindActive() ([]Regulation, error)
}

type gormRepo struct{ db *gorm.DB }

// NewRepository نمونه Repository مبتنی بر GORM می‌سازد.
func NewRepository(db *gorm.DB) Repository { return &gormRepo{db: db} }

// Create ضابطه جدید را ذخیره می‌کند.
func (r *gormRepo) Create(item *Regulation) error { return r.db.Create(item).Error }

// Update ضابطه را ذخیره می‌کند.
func (r *gormRepo) Update(item *Regulation) error { return r.db.Save(item).Error }

// Delete ضابطه را به‌صورت soft-delete حذف می‌کند.
func (r *gormRepo) Delete(item *Regulation) error { return r.db.Delete(item).Error }

// ReplaceServices خدمات ضابطه را جایگزین می‌کند.
func (r *gormRepo) ReplaceServices(regulationID uint, serviceIDs []uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("RegulationID = ?", regulationID).Delete(&RegulationService{}).Error; err != nil {
			return err
		}
		for _, sid := range serviceIDs {
			row := RegulationService{RegulationID: regulationID, ServiceID: sid}
			if err := tx.Create(&row).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// FindByID ضابطه را با خدمات برمی‌گرداند.
func (r *gormRepo) FindByID(id uint) (*Regulation, error) {
	var item Regulation
	err := r.db.Preload("Services").Where("ID = ?", id).First(&item).Error
	return &item, err
}

// FindAll همه ضوابط را با خدمات برمی‌گرداند.
func (r *gormRepo) FindAll() ([]Regulation, error) {
	var list []Regulation
	err := r.db.Preload("Services").Order("ID DESC").Find(&list).Error
	return list, err
}

// FindActive ضوابط فعال را برمی‌گرداند.
func (r *gormRepo) FindActive() ([]Regulation, error) {
	var list []Regulation
	err := r.db.Preload("Services").Where("IsActive = ?", true).Order("ID DESC").Find(&list).Error
	return list, err
}
