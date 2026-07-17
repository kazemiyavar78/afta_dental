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
	ServiceRate int `gorm:"column:ServiceRate;not null;default:0"`
	// تعرفه خدمت
	ServiceTariff int `gorm:"column:ServiceTariff;not null;default:0"`
	// کد بین‌المللی
	InternationalCode string `gorm:"column:InternationalCode;size:20;not null"`
	// تعداد پیش‌فرض
	DefaultCount int `gorm:"column:DefaultCount;not null;default:0"`
	// حداکثر تعداد
	MaximumCount int `gorm:"column:MaximumCount;not null;default:0"`
	// ویژگی‌های خدمت: خالی، #، *، #*
	ServiceFeatures string `gorm:"column:ServiceFeatures;size:5;not null;default:''"`
	// شماره جهت دندان دارد ؟
	HasDentalDirection bool `gorm:"column:HasDentalDirection;not null;default:false"`
	// نیاز به انتخاب شماره دندان دارد؟
	HasTooth bool `gorm:"column:HasTooth;not null;default:false"`
	// اجازه استفاده بیش از یکبار در پرونده
	AllowMultipleUse bool `gorm:"column:AllowMultipleUse;not null;default:false"`
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
	FindByID(id uint) (*ServiceItem, error)
	FindByServiceCode(code string) (*ServiceItem, error)
	FindAll() ([]ServiceItem, error)
	FindByExcludeServices(excludeServices []uint) ([]ServiceItem, error)
}

type gormRepo struct{ db *gorm.DB }

// NewRepository نمونه Repository مبتنی بر GORM می‌سازد.
func NewRepository(db *gorm.DB) Repository { return &gormRepo{db: db} }

// Create خدمت جدید را در دیتابیس ذخیره می‌کند.
func (r *gormRepo) Create(item *ServiceItem) error { return r.db.Create(item).Error }

// FindByID خدمت را با شناسه برمی‌گرداند.
func (r *gormRepo) FindByID(id uint) (*ServiceItem, error) {
	var item ServiceItem
	err := r.db.Where("ID = ?", id).First(&item).Error
	return &item, err
}

// FindByServiceCode خدمت را با کد خدمت برمی‌گرداند.
func (r *gormRepo) FindByServiceCode(code string) (*ServiceItem, error) {
	var item ServiceItem
	err := r.db.Where("ServiceCode = ?", code).First(&item).Error
	return &item, err
}

// FindAll همه خدمات را به ترتیب نزولی شناسه برمی‌گرداند.
func (r *gormRepo) FindAll() ([]ServiceItem, error) {
	var list []ServiceItem
	err := r.db.Order("ID DESC").Find(&list).Error
	return list, err
}
// FindByExcludeServices همه خدمات را به جز خدمات ارسالی برمی‌گرداند؛ اگر لیست exclude خالی باشد همه خدمات را برمی‌گرداند.
func (r *gormRepo) FindByExcludeServices(excludeServices []uint) ([]ServiceItem, error) {
	if len(excludeServices) == 0 {
		return r.FindAll()
	}
	var list []ServiceItem
	err := r.db.Where("ID NOT IN ?", excludeServices).Find(&list).Error
	if err != nil {
		return nil, err
	}
	return list, nil
}

// Update خدمت را ذخیره می‌کند.
func (r *gormRepo) Update(item *ServiceItem) error {
	return r.db.Save(item).Error
}

// Delete خدمت را به‌صورت soft-delete حذف می‌کند.
func (r *gormRepo) Delete(item *ServiceItem) error {
	return r.db.Delete(item).Error
}
