// این ماژول نمونه است؛ باقی Entity ها و Endpoint ها طبق همین الگو در فازهای بعدی تکمیل می‌شوند.
package fund

import (
	"time"

	"gorm.io/gorm"
)

type Fund struct {
	ID        int       `gorm:"column:ID;primaryKey;autoIncrement"`
	Name      string    `gorm:"column:Name;size:200;not null"`
	Balance   float64   `gorm:"column:Balance;not null;default:0"`
	CreatedAt time.Time `gorm:"column:CreatedAt;not null"`
}

func (Fund) TableName() string { return "Funds" }

type Repository interface {
	Create(f *Fund) error
	FindByID(id int) (*Fund, error)
	FindAll() ([]Fund, error)
}

type gormRepo struct{ db *gorm.DB }

func NewRepository(db *gorm.DB) Repository { return &gormRepo{db: db} }

func (r *gormRepo) Create(f *Fund) error { return r.db.Create(f).Error }

func (r *gormRepo) FindByID(id int) (*Fund, error) {
	var f Fund
	err := r.db.Where("ID = ?", id).First(&f).Error
	return &f, err
}

func (r *gormRepo) FindAll() ([]Fund, error) {
	var list []Fund
	err := r.db.Order("ID DESC").Find(&list).Error
	return list, err
}
