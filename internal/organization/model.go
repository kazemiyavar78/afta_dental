// این ماژول نمونه است؛ باقی Entity ها و Endpoint ها طبق همین الگو در فازهای بعدی تکمیل می‌شوند.
package organization

import (
	organizationpackage "github.com/tpdenta/afta-reception/internal/organizationPackage"
	"gorm.io/gorm"
)

// Organization مدل سازمان.
type Organization struct {
	gorm.Model
	Name string `gorm:"column:Name;size:200;not null"`
	// تکمیلی یا پایه
	IsTakmili bool `gorm:"column:IsTakmili;default:false"`
	// فعال یا غیرفعال
	IsActive      bool   `gorm:"column:IsActive;default:true"`
	IntegrityHash string `gorm:"column:IntegrityHash;size:128;not null"`
	// بسته تعرفه منتصب
	PackageID uint `gorm:"column:PackageID;not null;index"`

	Package organizationpackage.OrganizationPackage `gorm:"foreignKey:PackageID"`
}

// TableName نام جدول سازمان‌ها را برمی‌گرداند.
func (Organization) TableName() string { return "Organizations" }

// Repository اینترفیس CRUD.
type Repository interface {
	Create(o *Organization) error
	Update(o *Organization) error
	Delete(o *Organization) error
	FindByID(id uint) (*Organization, error)
	FindAll() ([]Organization, error)
	GetActive() ([]Organization, error)
}

type gormRepo struct{ db *gorm.DB }

// NewRepository نمونه Repository مبتنی بر GORM می‌سازد.
func NewRepository(db *gorm.DB) Repository { return &gormRepo{db: db} }

// Create سازمان جدید را در دیتابیس ذخیره می‌کند.
func (r *gormRepo) Create(o *Organization) error { return r.db.Create(o).Error }

// FindByID سازمان را با شناسه برمی‌گرداند (همراه با بسته).
func (r *gormRepo) FindByID(id uint) (*Organization, error) {
	var o Organization
	err := r.db.Preload("Package").Where("ID = ?", id).First(&o).Error
	return &o, err
}

// FindAll همه سازمان‌ها را به ترتیب نزولی شناسه برمی‌گرداند (همراه با بسته).
func (r *gormRepo) FindAll() ([]Organization, error) {
	var list []Organization
	err := r.db.Preload("Package").Order("ID DESC").Find(&list).Error
	return list, err
}

// GetActive سازمان‌های فعال را برمی‌گرداند.
func (r *gormRepo) GetActive() ([]Organization, error) {
	var list []Organization
	err := r.db.Preload("Package").Where("IsActive = ?", true).Order("ID DESC").Find(&list).Error
	return list, err
}

// Update سازمان را ذخیره می‌کند.
func (r *gormRepo) Update(o *Organization) error {
	return r.db.Save(o).Error
}

// Delete سازمان را به‌صورت soft-delete حذف می‌کند.
func (r *gormRepo) Delete(o *Organization) error {
	return r.db.Delete(o).Error
}
