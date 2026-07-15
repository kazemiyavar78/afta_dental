package patient

import (
	"strings"
	"time"

	"gorm.io/gorm"
)

// Patient مدل بیمار.
type Patient struct {
	gorm.Model
	//نام
	FirstName string `gorm:"column:first_name;not null"`
	//نام خانوادگی
	LastName string `gorm:"column:last_name;not null"`
	//کد ملی
	NationalCode string `gorm:"column:national_code;not null;unique"`
	//تاریخ تولد
	BirthDate time.Time `gorm:"column:birth_date;not null"`
	//آدرس
	Address *string `gorm:"column:address;nullable"`
	//تلفن منزل
	HomePhoneNumber *string `gorm:"column:home_phone_number;nullable"`
	//تلفن همراه
	MobilePhoneNumber *string `gorm:"column:mobile_phone_number;nullable"`
	//شماره پرونده منحصر به فرد
	FileNumber string `gorm:"column:file_number;not null;unique"`
	//جنسیت (مرد=true ، زن=false)
	Sex bool `gorm:"column:sex;not null"`
	//هش یکپارچگی داده‌ها
	IntegrityHash string `gorm:"column:integrity_hash;size:128;not null"`
}

// TableName نام جدول بیماران را برمی‌گرداند.
func (Patient) TableName() string { return "patients" }

// SearchFilter فیلترهای اختیاری جستجوی بیمار بر اساس فیلدها.
type SearchFilter struct {
	FirstName         string
	LastName          string
	NationalCode      string
	BirthDate         string
	Address           string
	HomePhoneNumber   string
	MobilePhoneNumber string
	FileNumber        string
	Sex               *bool
}

// Repository اینترفیس CRUD و جستجوی بیمار.
type Repository interface {
	Create(p *Patient) error
	Update(p *Patient) error
	Delete(p *Patient) error
	FindByID(id uint) (*Patient, error)
	FindAll() ([]Patient, error)
	Search(filter SearchFilter) ([]Patient, error)
	FindByNationalCode(nationalCode string) (*Patient, error)
	FindByFileNumber(fileNumber string) (*Patient, error)
	FindByFirstNameAndLastName(firstName string, lastName string) (*Patient, error)
}

type gormRepo struct{ db *gorm.DB }

// NewRepository نمونه Repository مبتنی بر GORM می‌سازد.
func NewRepository(db *gorm.DB) Repository { return &gormRepo{db: db} }

// Create بیمار جدید را در دیتابیس ذخیره می‌کند.
func (r *gormRepo) Create(p *Patient) error { return r.db.Create(p).Error }

// FindByID بیمار را با شناسه برمی‌گرداند.
func (r *gormRepo) FindByID(id uint) (*Patient, error) {
	var p Patient
	err := r.db.Where("id = ?", id).First(&p).Error
	return &p, err
}

// FindAll همه بیماران را به ترتیب نزولی شناسه برمی‌گرداند.
func (r *gormRepo) FindAll() ([]Patient, error) {
	var list []Patient
	err := r.db.Order("id DESC").Find(&list).Error
	return list, err
}

// Search بیماران را بر اساس فیلترهای اختیاری (LIKE روی رشته‌ها) برمی‌گرداند.
func (r *gormRepo) Search(filter SearchFilter) ([]Patient, error) {
	q := r.db.Model(&Patient{})

	if v := strings.TrimSpace(filter.FirstName); v != "" {
		q = q.Where("first_name LIKE ?", "%"+v+"%")
	}
	if v := strings.TrimSpace(filter.LastName); v != "" {
		q = q.Where("last_name LIKE ?", "%"+v+"%")
	}
	if v := strings.TrimSpace(filter.NationalCode); v != "" {
		q = q.Where("national_code LIKE ?", "%"+v+"%")
	}
	if v := strings.TrimSpace(filter.BirthDate); v != "" {
		if t, err := time.Parse("2006-01-02", v); err == nil {
			q = q.Where("CAST(birth_date AS date) = ?", t)
		} else {
			q = q.Where("CONVERT(varchar(10), birth_date, 23) LIKE ?", "%"+v+"%")
		}
	}
	if v := strings.TrimSpace(filter.Address); v != "" {
		q = q.Where("address LIKE ?", "%"+v+"%")
	}
	if v := strings.TrimSpace(filter.HomePhoneNumber); v != "" {
		q = q.Where("home_phone_number LIKE ?", "%"+v+"%")
	}
	if v := strings.TrimSpace(filter.MobilePhoneNumber); v != "" {
		q = q.Where("mobile_phone_number LIKE ?", "%"+v+"%")
	}
	if v := strings.TrimSpace(filter.FileNumber); v != "" {
		q = q.Where("file_number LIKE ?", "%"+v+"%")
	}
	if filter.Sex != nil {
		q = q.Where("sex = ?", *filter.Sex)
	}

	var list []Patient
	err := q.Order("id DESC").Find(&list).Error
	return list, err
}

// FindByNationalCode بیمار را با کد ملی برمی‌گرداند.
func (r *gormRepo) FindByNationalCode(nationalCode string) (*Patient, error) {
	var p Patient
	err := r.db.Where("national_code = ?", nationalCode).First(&p).Error
	return &p, err
}

// FindByFileNumber بیمار را با شماره پرونده برمی‌گرداند.
func (r *gormRepo) FindByFileNumber(fileNumber string) (*Patient, error) {
	var p Patient
	err := r.db.Where("file_number = ?", fileNumber).First(&p).Error
	return &p, err
}

// FindByFirstNameAndLastName بیمار را با نام و نام خانوادگی برمی‌گرداند.
func (r *gormRepo) FindByFirstNameAndLastName(firstName string, lastName string) (*Patient, error) {
	var p Patient
	err := r.db.Where("first_name = ? AND last_name = ?", firstName, lastName).First(&p).Error
	return &p, err
}

// Update بیمار را ذخیره می‌کند.
func (r *gormRepo) Update(p *Patient) error {
	return r.db.Save(p).Error
}

// Delete بیمار را به‌صورت soft-delete حذف می‌کند.
func (r *gormRepo) Delete(p *Patient) error {
	return r.db.Delete(p).Error
}
