// این ماژول نمونه است؛ باقی Entity ها و Endpoint ها طبق همین الگو در فازهای بعدی تکمیل می‌شوند.
package tariff

import (
	"time"

	"gorm.io/gorm"
)

type Tariff struct {
	ID        int       `gorm:"column:ID;primaryKey;autoIncrement"`
	Name      string    `gorm:"column:Name;size:200;not null"`
	Amount    float64   `gorm:"column:Amount;not null"`
	CreatedAt time.Time `gorm:"column:CreatedAt;not null"`
}

func (Tariff) TableName() string { return "Tariffs" }

type Repository interface {
	Create(t *Tariff) error
	FindByID(id int) (*Tariff, error)
	FindAll() ([]Tariff, error)
}

type gormRepo struct{ db *gorm.DB }

func NewRepository(db *gorm.DB) Repository { return &gormRepo{db: db} }

func (r *gormRepo) Create(t *Tariff) error { return r.db.Create(t).Error }

func (r *gormRepo) FindByID(id int) (*Tariff, error) {
	var t Tariff
	err := r.db.Where("ID = ?", id).First(&t).Error
	return &t, err
}

func (r *gormRepo) FindAll() ([]Tariff, error) {
	var list []Tariff
	err := r.db.Order("ID DESC").Find(&list).Error
	return list, err
}
