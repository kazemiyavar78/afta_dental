package reception

import (
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Repository اینترفیس دسترسی داده پذیرش.
type Repository interface {
	Create(r *Reception) error
	Update(r *Reception) error
	Delete(id uint) error
	Restore(id uint) error
	FindByID(id uint) (*Reception, error)
	FindByIDUnscoped(id uint) (*Reception, error)
	FindAll() ([]Reception, error)
	FindFirst() (*Reception, error)
	FindLast() (*Reception, error)
	FindPrev(cursor uint) (*Reception, error)
	FindNext(cursor uint) (*Reception, error)
	ReplaceServices(receptionID uint, services []ReceptionService) error
	FindByPatientID(patientID uint) ([]Reception, error)
	FindPreviousForPatient(patientID, currentID uint) (*Reception, error)
	FindPatientReceptionsInRange(patientID uint, from, to time.Time) ([]Reception, error)
	CountPhotos(receptionID uint) (int64, error)
	AddPhoto(photo *ReceptionPhoto) error
}

type gormRepository struct {
	db *gorm.DB
}

// NewRepository نمونه Repository می‌سازد.
func NewRepository(db *gorm.DB) Repository {
	return &gormRepository{db: db}
}

// Create پذیرش جدید را ذخیره می‌کند.
func (r *gormRepository) Create(rec *Reception) error {
	return r.db.Create(rec).Error
}

// Update پذیرش را به‌روزرسانی می‌کند.
func (r *gormRepository) Update(rec *Reception) error {
	return r.db.Session(&gorm.Session{FullSaveAssociations: false}).Save(rec).Error
}

// Delete پذیرش را به‌صورت soft-delete حذف می‌کند.
func (r *gormRepository) Delete(id uint) error {
	return r.db.Delete(&Reception{}, id).Error
}

// Restore پذیرش حذف‌شده را بازیابی می‌کند.
func (r *gormRepository) Restore(id uint) error {
	return r.db.Unscoped().Model(&Reception{}).Where("ID = ?", id).Update("deleted_at", nil).Error
}

// FindByID پذیرش را با خدمات برمی‌گرداند.
func (r *gormRepository) FindByID(id uint) (*Reception, error) {
	var rec Reception
	err := r.db.Preload("Services").Preload("Photos").Where("ID = ?", id).First(&rec).Error
	if err != nil {
		return nil, err
	}
	return &rec, nil
}

// FindByIDUnscoped پذیرش را حتی در صورت حذف‌شده بودن برمی‌گرداند.
func (r *gormRepository) FindByIDUnscoped(id uint) (*Reception, error) {
	var rec Reception
	err := r.db.Unscoped().Preload("Services").Where("ID = ?", id).First(&rec).Error
	if err != nil {
		return nil, err
	}
	return &rec, nil
}

// FindAll لیست پذیرش‌های غیرحذف‌شده را برمی‌گرداند.
func (r *gormRepository) FindAll() ([]Reception, error) {
	var list []Reception
	err := r.db.Preload("Services").Order("ID DESC").Find(&list).Error
	return list, err
}

// FindFirst قدیمی‌ترین پذیرش غیرحذف‌شده را برمی‌گرداند.
// از Take به‌جای First استفاده می‌شود تا در MSSQL ستون ID در ORDER BY تکراری نشود
// (First به‌صورت خودکار primary key را هم به ORDER BY اضافه می‌کند).
func (r *gormRepository) FindFirst() (*Reception, error) {
	var rec Reception
	err := r.db.Preload("Services").Order("ID ASC").Take(&rec).Error
	if err != nil {
		return nil, err
	}
	return &rec, nil
}

// FindLast جدیدترین پذیرش غیرحذف‌شده را برمی‌گرداند.
func (r *gormRepository) FindLast() (*Reception, error) {
	var rec Reception
	err := r.db.Preload("Services").Order("ID DESC").Take(&rec).Error
	if err != nil {
		return nil, err
	}
	return &rec, nil
}

// FindPrev پذیرش قبلی نسبت به cursor را برمی‌گرداند.
func (r *gormRepository) FindPrev(cursor uint) (*Reception, error) {
	var rec Reception
	err := r.db.Preload("Services").Where("ID < ?", cursor).Order("ID DESC").Take(&rec).Error
	if err != nil {
		return nil, err
	}
	return &rec, nil
}

// FindNext پذیرش بعدی نسبت به cursor را برمی‌گرداند.
func (r *gormRepository) FindNext(cursor uint) (*Reception, error) {
	var rec Reception
	err := r.db.Preload("Services").Where("ID > ?", cursor).Order("ID ASC").Take(&rec).Error
	if err != nil {
		return nil, err
	}
	return &rec, nil
}

// ReplaceServices خدمات پذیرش را جایگزین می‌کند (حذف نرم قبلی + درج جدید).
func (r *gormRepository) ReplaceServices(receptionID uint, services []ReceptionService) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("ReceptionID = ?", receptionID).Delete(&ReceptionService{}).Error; err != nil {
			return err
		}
		if len(services) == 0 {
			return nil
		}
		for i := range services {
			services[i].ID = 0
			services[i].ReceptionID = receptionID
		}
		return tx.Omit(clause.Associations).Create(&services).Error
	})
}

// FindByPatientID پذیرش‌های یک پرونده را به ترتیب تاریخ پذیرش برمی‌گرداند.
func (r *gormRepository) FindByPatientID(patientID uint) ([]Reception, error) {
	var list []Reception
	err := r.db.Preload("Services").Preload("Photos").
		Where("PatientID = ?", patientID).
		Order("ReceptionDate ASC, ID ASC").
		Find(&list).Error
	return list, err
}

// FindPreviousForPatient پذیرش قبلی همان پرونده (قبل از currentID) را برمی‌گرداند.
func (r *gormRepository) FindPreviousForPatient(patientID, currentID uint) (*Reception, error) {
	var rec Reception
	err := r.db.Preload("Services").
		Where("PatientID = ? AND ID < ?", patientID, currentID).
		Order("ID DESC").
		Take(&rec).Error
	if err != nil {
		return nil, err
	}
	return &rec, nil
}

// FindPatientReceptionsInRange پذیرش‌های پرونده در بازه زمانی را برمی‌گرداند.
func (r *gormRepository) FindPatientReceptionsInRange(patientID uint, from, to time.Time) ([]Reception, error) {
	var list []Reception
	err := r.db.Preload("Services").
		Where("PatientID = ? AND ReceptionDate >= ? AND ReceptionDate <= ?", patientID, from, to).
		Order("ReceptionDate ASC, ID ASC").
		Find(&list).Error
	return list, err
}

// CountPhotos تعداد عکس‌های آپلودشده پذیرش را برمی‌گرداند.
func (r *gormRepository) CountPhotos(receptionID uint) (int64, error) {
	var count int64
	err := r.db.Model(&ReceptionPhoto{}).Where("ReceptionID = ?", receptionID).Count(&count).Error
	return count, err
}

// AddPhoto عکس پذیرش را ذخیره می‌کند.
func (r *gormRepository) AddPhoto(photo *ReceptionPhoto) error {
	return r.db.Create(photo).Error
}
