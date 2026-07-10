package session

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Session مدل جدول نشست‌ها.
type Session struct {
	ID                 MSSQLUUID `gorm:"column:Id;type:uniqueidentifier;primaryKey"`
	Ip                 string    `gorm:"column:Ip;size:45;not null"`
	Browser            string    `gorm:"column:Browser;size:512;not null"`
	CreationTime       time.Time `gorm:"column:CreationTime;not null"`
	PersonnelAccountID int       `gorm:"column:PersonnelAccountID;not null"`
}

// TableName نام جدول در دیتابیس.
func (Session) TableName() string {
	return "Sessions"
}

// Repository عملیات CRUD خام نشست‌ها.
type Repository interface {
	Create(session *Session) error
	FindByID(id uuid.UUID) (*Session, error)
	FindByUserID(userID int) ([]Session, error)
	Delete(id uuid.UUID) error
	DeleteByUserID(userID int) error
	CountByUserID(userID int) (int64, error)
	DeleteOldestByUserID(userID int) error
	FindAll() ([]Session, error)
}

type gormRepository struct {
	db *gorm.DB
}

// NewRepository نمونه Repository می‌سازد.
func NewRepository(db *gorm.DB) Repository {
	return &gormRepository{db: db}
}

func (r *gormRepository) Create(session *Session) error {
	return r.db.Create(session).Error
}

func (r *gormRepository) FindByID(id uuid.UUID) (*Session, error) {
	var s Session
	// MSSQLUUID تا Value() بایت‌ها را به فرمت SQL Server تبدیل کند.
	err := r.db.Where("Id = ?", MSSQLUUID(id)).First(&s).Error
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *gormRepository) FindByUserID(userID int) ([]Session, error) {
	var sessions []Session
	err := r.db.Where("PersonnelAccountID = ?", userID).Order("CreationTime DESC").Find(&sessions).Error
	return sessions, err
}

func (r *gormRepository) Delete(id uuid.UUID) error {
	return r.db.Where("Id = ?", MSSQLUUID(id)).Delete(&Session{}).Error
}

func (r *gormRepository) DeleteByUserID(userID int) error {
	return r.db.Where("PersonnelAccountID = ?", userID).Delete(&Session{}).Error
}

func (r *gormRepository) CountByUserID(userID int) (int64, error) {
	var count int64
	err := r.db.Model(&Session{}).Where("PersonnelAccountID = ?", userID).Count(&count).Error
	return count, err
}

func (r *gormRepository) DeleteOldestByUserID(userID int) error {
	var oldest Session
	err := r.db.Where("PersonnelAccountID = ?", userID).Order("CreationTime ASC").First(&oldest).Error
	if err == gorm.ErrRecordNotFound {
		return nil
	}
	if err != nil {
		return err
	}
	return r.db.Where("Id = ?", oldest.ID).Delete(&Session{}).Error
}

func (r *gormRepository) FindAll() ([]Session, error) {
	var sessions []Session
	err := r.db.Order("CreationTime DESC").Find(&sessions).Error
	return sessions, err
}
