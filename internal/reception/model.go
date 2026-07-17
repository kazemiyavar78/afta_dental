package reception

import (
	"time"

	"gorm.io/gorm"
)

// ReceptionStatus وضعیت پذیرش بیمار.
type ReceptionStatus string

const (
	// ReceptionStatusDraft پیش‌نویس — هنگام ایجاد جدید
	ReceptionStatusDraft ReceptionStatus = "draft"
	// ReceptionStatusSaved ذخیره‌شده — پس از کلیک ذخیره
	ReceptionStatusSaved ReceptionStatus = "saved"
)

// TeethDirection جهت دندان (مقادیر ثابت عددی برای ذخیره در DB).
type TeethDirection uint8

const (
	TeethDirectionUpperLeft  TeethDirection = 1
	TeethDirectionUpperRight TeethDirection = 2
	TeethDirectionLowerLeft  TeethDirection = 3
	TeethDirectionLowerRight TeethDirection = 4
)

// Reception مدل پذیرش بیمار.
type Reception struct {
	gorm.Model
	PatientID uint `gorm:"column:PatientID;not null"`
	// بیمه پایه (اختیاری؛ حداقل یکی از پایه یا تکمیلی الزامی است)
	InsuranceID *uint `gorm:"column:InsuranceID"`
	// بیمه تکمیلی
	AdditionalInsuranceID *uint `gorm:"column:AdditionalInsuranceID"`
	// پزشک (الزامی هنگام ذخیره)
	DoctorID *uint `gorm:"column:DoctorID"`
	// دستیار (اختیاری)
	AssistantID *uint `gorm:"column:AssistantID"`
	// تاریخ اعتبار دفترچه
	BookingDate *time.Time `gorm:"column:BookingDate"`
	// تاریخ پذیرش
	ReceptionDate time.Time `gorm:"column:ReceptionDate;not null"`
	// وضعیت پذیرش
	Status string `gorm:"column:Status;size:50;not null;default:draft"`
	// توضیحات
	Description string `gorm:"column:Description;size:200;not null;default:''"`
	// تخفیف پذیرش
	Discount int64 `gorm:"column:Discount;not null;default:0"`
	// کد معرفی‌نامه تکمیلی
	ReferralCode *int64 `gorm:"column:ReferralCode"`
	// سقف پوشش بیمه تکمیلی
	AdditionalInsuranceCoverage *int64 `gorm:"column:AdditionalInsuranceCoverage"`
	// درصد بیمه تکمیلی (۰ تا ۱۰۰)
	AdditionalInsurancePercentage *uint8 `gorm:"column:AdditionalInsurancePercentage"`
	// کاربر ثبت‌کننده
	RegisteredByID *uint `gorm:"column:RegisteredByID"`
	// لیست خدمات پذیرش‌شده
	Services []ReceptionService `gorm:"foreignKey:ReceptionID;references:ID"`
}

// TableName نام جدول در دیتابیس.
func (Reception) TableName() string {
	return "Receptions"
}

// ReceptionService مدل خدمات پذیرش بیمار.
type ReceptionService struct {
	gorm.Model
	ReceptionID uint `gorm:"column:ReceptionID;not null;index"`
	// کد خدمت (شناسه)
	ServiceID uint `gorm:"column:ServiceID;not null"`
	// نام خدمت
	ServiceName string `gorm:"column:ServiceName;size:200;not null"`
	// تعداد خدمت
	Quantity int `gorm:"column:Quantity;not null;default:1"`
	// نرخ خدمت
	ServiceAmount int64 `gorm:"column:ServiceAmount;not null"`
	// تعرفه خدمت
	ServiceTariff int64 `gorm:"column:ServiceTariff;not null"`
	// سهم سازمان خدمت
	ServiceOrganizationShare int64 `gorm:"column:ServiceOrganizationShare;not null"`
	// سهم بیمه تکمیلی خدمت
	ServiceSupplementaryInsuranceShare int64 `gorm:"column:ServiceSupplementaryInsuranceShare;not null"`
	// سهم یارانه
	ServiceSubsidyShare int64 `gorm:"column:ServiceSubsidyShare;not null"`
	// توضیحات خدمت
	ServiceDescription string `gorm:"column:ServiceDescription;size:200;not null;default:''"`
	// شماره دندان
	TeethNumber *uint8 `gorm:"column:TeethNumber"`
	// جهت دندان
	TeethDirection *uint8 `gorm:"column:TeethDirection"`
}

// TableName نام جدول در دیتابیس.
func (ReceptionService) TableName() string {
	return "ReceptionServices"
}
