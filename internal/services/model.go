package services

import (
	"gorm.io/gorm"
)

// ServiceItem مدل خدمت.
type ServiceItem struct {
	gorm.Model
	// کد خدمت
	ServiceCode string `gorm:"column:ServiceCode;size:20;not null"`
	// نام خدمت
	Name string `gorm:"column:Name;size:200;not null"`
	// ضریب فنی
	TechnicalCoefficient float64 `gorm:"column:TechnicalCoefficient;not null;default:0"`
	// ضریب حرفه‌ای
	ProfessionalCoefficient float64 `gorm:"column:ProfessionalCoefficient;not null;default:0"`
	// ضریب مصرفی
	ConsumptionCoefficient float64 `gorm:"column:ConsumptionCoefficient;not null;default:0"`
	// نرخ خدمت
	ServiceRate float64 `gorm:"column:ServiceRate;not null;default:0"`
	// تعرفه خدمت
	ServiceTariff float64 `gorm:"column:ServiceTariff;not null;default:0"`
	// کد بین‌المللی
	InternationalCode string `gorm:"column:InternationalCode;size:20;not null"`
	// تعداد پیش‌فرض
	DefaultCount int `gorm:"column:DefaultCount;not null;default:0"`
	// حداکثر تعداد
	MaximumCount int `gorm:"column:MaximumCount;not null;default:0"`
	// ویژگی‌های خدمت: خالی، #، *، #*
	ServiceFeatures string `gorm:"column:ServiceFeatures;size:5;not null;default:''"`
	IsActive        bool   `gorm:"column:IsActive;default:true"`
	IntegrityHash   string `gorm:"column:IntegrityHash;size:128;not null"`
}

// TableName نام جدول خدمات را برمی‌گرداند.
func (ServiceItem) TableName() string { return "Services" }

// Repository اینترفیس CRUD خدمت.
type Repository interface {
	Create(item *ServiceItem) error
	Update(item *ServiceItem) error
	Delete(item *ServiceItem) error
	FindByID(id int) (*ServiceItem, error)
	FindAll() ([]ServiceItem, error)
}

type gormRepo struct{ db *gorm.DB }

// NewRepository نمونه Repository مبتنی بر GORM می‌سازد.
func NewRepository(db *gorm.DB) Repository { return &gormRepo{db: db} }

// Create خدمت جدید را در دیتابیس ذخیره می‌کند.
func (r *gormRepo) Create(item *ServiceItem) error { return r.db.Create(item).Error }

// FindByID خدمت را با شناسه برمی‌گرداند.
func (r *gormRepo) FindByID(id int) (*ServiceItem, error) {
	var item ServiceItem
	err := r.db.Where("ID = ?", id).First(&item).Error
	return &item, err
}

// FindAll همه خدمات را به ترتیب نزولی شناسه برمی‌گرداند.
func (r *gormRepo) FindAll() ([]ServiceItem, error) {
	var list []ServiceItem
	err := r.db.Order("ID DESC").Find(&list).Error
	return list, err
}

// Update خدمت را ذخیره می‌کند.
func (r *gormRepo) Update(item *ServiceItem) error {
	return r.db.Save(item).Error
}

// Delete خدمت را به‌صورت soft-delete حذف می‌کند.
func (r *gormRepo) Delete(item *ServiceItem) error {
	return r.db.Delete(item).Error
}
