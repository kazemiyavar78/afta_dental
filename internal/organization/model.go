// این ماژول نمونه است؛ باقی Entity ها و Endpoint ها طبق همین الگو در فازهای بعدی تکمیل می‌شوند.
package organization

import (
	"time"

	"gorm.io/gorm"
)

// Organization مدل سازمان (نمونه).
type Organization struct {
	ID        int       `gorm:"column:ID;primaryKey;autoIncrement"`
	Name      string    `gorm:"column:Name;size:200;not null"`
	CreatedAt time.Time `gorm:"column:CreatedAt;not null"`
}

func (Organization) TableName() string { return "Organizations" }

// Repository اینترفیس CRUD.
type Repository interface {
	Create(o *Organization) error
	FindByID(id int) (*Organization, error)
	FindAll() ([]Organization, error)
}

type gormRepo struct{ db *gorm.DB }

func NewRepository(db *gorm.DB) Repository { return &gormRepo{db: db} }

func (r *gormRepo) Create(o *Organization) error { return r.db.Create(o).Error }

func (r *gormRepo) FindByID(id int) (*Organization, error) {
	var o Organization
	err := r.db.Where("ID = ?", id).First(&o).Error
	return &o, err
}

func (r *gormRepo) FindAll() ([]Organization, error) {
	var list []Organization
	err := r.db.Order("ID DESC").Find(&list).Error
	return list, err
}
