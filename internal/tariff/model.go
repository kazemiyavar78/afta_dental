package tariff

import (
	"gorm.io/gorm"
)

// Tariff مدل ذخیره تعرفه هر خدمت برای یک سازمان.
type Tariff struct {
	gorm.Model
	// سازمان
	OrganizationID uint `gorm:"column:OrganizationID;not null"`
	// خدمت
	ServiceID uint `gorm:"column:ServiceID;not null"`
	// نرخ
	Amount int64 `gorm:"column:Amount;not null"`
	// تعرفه
	TariffAmount int64 `gorm:"column:TariffAmount;not null"`
	// سهم سازمان
	OrganizationShare int64 `gorm:"column:OrganizationShare;not null"`
	// سهم تکمیلی
	SupplementaryShare int64 `gorm:"column:SupplementaryShare;not null"`
	// سهم یارانه
	SubsidyShare int64 `gorm:"column:SubsidyShare;not null"`
}

// TableName نام جدول تعرفه‌ها را برمی‌گرداند.
func (Tariff) TableName() string { return "Tariffs" }

// Repository اینترفیس دسترسی داده تعرفه.
type Repository interface {
	Create(t *Tariff) error
	FindByOrganizationID(organizationID uint) (*Tariff, error)
	FindAllByOrganizationID(organizationID uint) ([]Tariff, error)
	FindByServiceID(serviceID uint) (*Tariff, error)
	FindByOrganizationIDAndServiceID(organizationID uint, serviceID uint) (*Tariff, error)
	FindAll() ([]Tariff, error)
	FindByID(id uint) (*Tariff, error)
	Update(t *Tariff) error
	Delete(id uint) error
}

type gormRepo struct{ db *gorm.DB }

// NewRepository نمونه Repository مبتنی بر GORM می‌سازد.
func NewRepository(db *gorm.DB) Repository { return &gormRepo{db: db} }

// Create تعرفه جدید را در دیتابیس ذخیره می‌کند.
func (r *gormRepo) Create(t *Tariff) error { return r.db.Create(t).Error }

// FindByOrganizationID اولین تعرفه مربوط به سازمان را برمی‌گرداند.
func (r *gormRepo) FindByOrganizationID(organizationID uint) (*Tariff, error) {
	var t Tariff
	err := r.db.Where("OrganizationID = ?", organizationID).First(&t).Error
	return &t, err
}

// FindAllByOrganizationID همه تعرفه‌های یک سازمان را برمی‌گرداند.
func (r *gormRepo) FindAllByOrganizationID(organizationID uint) ([]Tariff, error) {
	var list []Tariff
	err := r.db.Where("OrganizationID = ?", organizationID).Order("ID DESC").Find(&list).Error
	return list, err
}

// FindByServiceID اولین تعرفه مربوط به خدمت را برمی‌گرداند.
func (r *gormRepo) FindByServiceID(serviceID uint) (*Tariff, error) {
	var t Tariff
	err := r.db.Where("ServiceID = ?", serviceID).First(&t).Error
	return &t, err
}

// FindByOrganizationIDAndServiceID تعرفه یک سازمان و خدمت را برمی‌گرداند.
func (r *gormRepo) FindByOrganizationIDAndServiceID(organizationID uint, serviceID uint) (*Tariff, error) {
	var t Tariff
	err := r.db.Where("OrganizationID = ? AND ServiceID = ?", organizationID, serviceID).First(&t).Error
	return &t, err
}

// FindByID تعرفه را با شناسه برمی‌گرداند.
func (r *gormRepo) FindByID(id uint) (*Tariff, error) {
	var t Tariff
	err := r.db.Where("ID = ?", id).First(&t).Error
	return &t, err
}

// FindAll همه تعرفه‌ها را برمی‌گرداند.
func (r *gormRepo) FindAll() ([]Tariff, error) {
	var list []Tariff
	err := r.db.Order("ID DESC").Find(&list).Error
	return list, err
}

// Update تعرفه را ذخیره می‌کند.
func (r *gormRepo) Update(t *Tariff) error { return r.db.Save(t).Error }

// Delete تعرفه را به‌صورت soft-delete حذف می‌کند.
func (r *gormRepo) Delete(id uint) error { return r.db.Delete(&Tariff{}, id).Error }
