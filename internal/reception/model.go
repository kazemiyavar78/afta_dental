// این ماژول نمونه است؛ باقی Entity ها و Endpoint ها طبق همین الگو در فازهای بعدی تکمیل می‌شوند.
package reception

import (
	"time"
)

// Reception مدل پذیرش بیمار (نمونه).
type Reception struct {
	ID            int       `gorm:"column:ID;primaryKey;autoIncrement"`
	PatientName   string    `gorm:"column:PatientName;size:200;not null"`
	DoctorID      int       `gorm:"column:DoctorID;not null"`
	ReceptionDate time.Time `gorm:"column:ReceptionDate;not null"`
	Status        string    `gorm:"column:Status;size:50;not null;default:pending"`
	CreatedAt     time.Time `gorm:"column:CreatedAt;not null"`
}

// TableName نام جدول در دیتابیس.
func (Reception) TableName() string {
	return "Receptions"
}
