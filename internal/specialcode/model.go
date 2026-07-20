package specialcode

import (
	"strconv"

	"github.com/tpdenta/afta-reception/internal/platform/security/integrity"
	"gorm.io/gorm"
)

// SpecialCode مدل کد خاص (سازمان سوم در پذیرش).
type SpecialCode struct {
	gorm.Model
	// کد خاص
	Code string `gorm:"column:Code;size:50;not null;uniqueIndex"`
	// نام کد خاص
	Name string `gorm:"column:Name;size:200;not null"`
	// توضیحات کد خاص
	Description string `gorm:"column:Description;size:500;not null;default:''"`
	// درصد کد خاص (۰ تا ۱۰۰)
	Percentage uint8 `gorm:"column:Percentage;not null;default:0"`
	// وضعیت فعال بودن
	IsActive bool `gorm:"column:IsActive;not null;default:true"`
	// هش یکپارچگی
	IntegrityHash string `gorm:"column:IntegrityHash;size:128;not null"`
}

// TableName نام جدول کدهای خاص.
func (SpecialCode) TableName() string { return "SpecialCodes" }

// Repository اینترفیس CRUD کد خاص.
type Repository interface {
	Create(item *SpecialCode) error
	Update(item *SpecialCode) error
	Delete(item *SpecialCode) error
	FindByID(id uint) (*SpecialCode, error)
	FindByCode(code string) (*SpecialCode, error)
	FindAll() ([]SpecialCode, error)
}

type gormRepo struct{ db *gorm.DB }

// NewRepository نمونه Repository مبتنی بر GORM می‌سازد.
func NewRepository(db *gorm.DB) Repository { return &gormRepo{db: db} }

// Create کد خاص جدید را ذخیره می‌کند.
func (r *gormRepo) Create(item *SpecialCode) error { return r.db.Create(item).Error }

// Update کد خاص را ذخیره می‌کند.
func (r *gormRepo) Update(item *SpecialCode) error { return r.db.Save(item).Error }

// Delete کد خاص را به‌صورت soft-delete حذف می‌کند.
func (r *gormRepo) Delete(item *SpecialCode) error { return r.db.Delete(item).Error }

// FindByID کد خاص را با شناسه برمی‌گرداند.
func (r *gormRepo) FindByID(id uint) (*SpecialCode, error) {
	var item SpecialCode
	err := r.db.Where("ID = ?", id).First(&item).Error
	return &item, err
}

// FindByCode کد خاص را با کد برمی‌گرداند.
func (r *gormRepo) FindByCode(code string) (*SpecialCode, error) {
	var item SpecialCode
	err := r.db.Where("Code = ?", code).First(&item).Error
	return &item, err
}

// FindAll همه کدهای خاص را برمی‌گرداند.
func (r *gormRepo) FindAll() ([]SpecialCode, error) {
	var list []SpecialCode
	err := r.db.Order("ID DESC").Find(&list).Error
	return list, err
}

// BuildIntegrityFields فیلدهای HMAC کد خاص را برمی‌گرداند.
func BuildIntegrityFields(item *SpecialCode) []string {
	active := "0"
	if item.IsActive {
		active = "1"
	}
	return []string{
		strconv.FormatUint(uint64(item.ID), 10),
		item.Code,
		item.Name,
		item.Description,
		strconv.FormatUint(uint64(item.Percentage), 10),
		active,
	}
}

// SignIntegrityHash هش یکپارچگی کد خاص را محاسبه می‌کند.
func SignIntegrityHash(signer *integrity.Signer, item *SpecialCode) string {
	return signer.Sign(BuildIntegrityFields(item)...)
}

// VerifyIntegrity یکپارچگی کد خاص را بررسی می‌کند.
func VerifyIntegrity(signer *integrity.Signer, item *SpecialCode) bool {
	return signer.Verify(item.IntegrityHash, BuildIntegrityFields(item)...)
}
