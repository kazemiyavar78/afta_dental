// این ماژول نمونه است؛ باقی Entity ها و Endpoint ها طبق همین الگو در فازهای بعدی تکمیل می‌شوند.
package reception

import "gorm.io/gorm"

// Repository اینترفیس CRUD خام پذیرش.
type Repository interface {
	Create(r *Reception) error
	FindByID(id int) (*Reception, error)
	FindAll() ([]Reception, error)
}

type gormRepository struct {
	db *gorm.DB
}

// NewRepository نمونه Repository می‌سازد.
func NewRepository(db *gorm.DB) Repository {
	return &gormRepository{db: db}
}

func (r *gormRepository) Create(rec *Reception) error {
	return r.db.Create(rec).Error
}

func (r *gormRepository) FindByID(id int) (*Reception, error) {
	var rec Reception
	err := r.db.Where("ID = ?", id).First(&rec).Error
	if err != nil {
		return nil, err
	}
	return &rec, nil
}

func (r *gormRepository) FindAll() ([]Reception, error) {
	var list []Reception
	err := r.db.Order("ID DESC").Find(&list).Error
	return list, err
}
