package organizationpackage

import (
	"gorm.io/gorm"
)

// OrganizationPackage بسته تعرفه سازمان.
type OrganizationPackage struct {
	gorm.Model
	PackageName        string `gorm:"column:PackageName;size:200;not null"`
	PackageDescription string `gorm:"column:PackageDescription;size:200;not null"`
	// تعرفه ضریب فنی بسته
	TechnicalCoefficient int `gorm:"column:TechnicalCoefficient;not null"`
	// تعرفه ضریب حرفه‌ای بسته
	TechnicalProfessionalCoefficient int `gorm:"column:TechnicalProfessionalCoefficient;not null"`
	// تعرفه ضریب مصرفی بسته
	ConsumptionCoefficient int `gorm:"column:ConsumptionCoefficient;not null"`
	// درصد یارانه
	SubsidyPercentage int `gorm:"column:SubsidyPercentage;not null"`
	// درصد تکمیلی
	SupplementaryPercentage int    `gorm:"column:SupplementaryPercentage;not null"`
	// درصد سهم سازمان در بسته
	OrganizationPercentage int `gorm:"column:OrganizationPercentage;not null"`
	IntegrityHash           string `gorm:"column:IntegrityHash;size:128;not null"`
}

// OrganizationPackageRelation جدول بین سازمان و بسته‌ها (هر سازمان یک بسته؛ هر بسته چند سازمان).
type OrganizationPackageRelation struct {
	gorm.Model
	OrganizationID int `gorm:"column:OrganizationID;not null;index"`
	PackageID      int `gorm:"column:PackageID;not null;index"`
}

// TableName نام جدول بین سازمان و بسته‌ها.
func (OrganizationPackageRelation) TableName() string { return "OrganizationPackageRelations" }

// TableName نام جدول بسته‌های تعرفه برای سازمان.
func (OrganizationPackage) TableName() string { return "OrganizationPackages" }

// Repository اینترفیس CRUD بسته تعرفه.
type Repository interface {
	Create(o *OrganizationPackage) error
	Update(o *OrganizationPackage) error
	Delete(o *OrganizationPackage) error
	FindByID(id int) (*OrganizationPackage, error)
	FindAll() ([]OrganizationPackage, error)
}

type gormRepo struct{ db *gorm.DB }

// NewRepository نمونه Repository مبتنی بر GORM می‌سازد.
func NewRepository(db *gorm.DB) Repository { return &gormRepo{db: db} }

// Create بسته تعرفه جدید را در دیتابیس ذخیره می‌کند.
func (r *gormRepo) Create(o *OrganizationPackage) error { return r.db.Create(o).Error }

// FindByID بسته تعرفه را با شناسه برمی‌گرداند.
func (r *gormRepo) FindByID(id int) (*OrganizationPackage, error) {
	var o OrganizationPackage
	err := r.db.Where("ID = ?", id).First(&o).Error
	return &o, err
}

// FindAll بسته‌های تعرفه را برمی‌گرداند.
func (r *gormRepo) FindAll() ([]OrganizationPackage, error) {
	var list []OrganizationPackage
	err := r.db.Order("ID DESC").Find(&list).Error
	return list, err
}

// Update بسته تعرفه را بروزرسانی می‌کند.
func (r *gormRepo) Update(o *OrganizationPackage) error { return r.db.Save(o).Error }

// Delete بسته تعرفه را به‌صورت soft-delete حذف می‌کند.
func (r *gormRepo) Delete(o *OrganizationPackage) error { return r.db.Delete(o).Error }
