package user

import (
	"time"

	"github.com/tpdenta/afta-reception/internal/platform/security/encryption"
)

// User مدل جدول کاربران (PersonnelAccount).
type User struct {
	ID                int        `gorm:"column:ID;primaryKey;autoIncrement"`
	Username          string     `gorm:"column:Username;uniqueIndex;size:100;not null"`
	PasswordHash      string     `gorm:"column:PasswordHash;size:255;not null"`
	Address           string     `gorm:"column:Address;size:500"`
	Name              string     `gorm:"column:Name;size:100;not null"`
	Family            string     `gorm:"column:Family;size:100;not null"`
	PhoneNumber       string     `gorm:"column:PhoneNumber;size:20"`
	MedicalCode       *string    `gorm:"column:MedicalCode;size:50"`
	RoleID            int        `gorm:"column:RoleID;not null"`
	IsActive          bool       `gorm:"column:IsActive;not null;default:1"`
	IsLocked          bool       `gorm:"column:IsLocked;not null;default:0"`
	SecurityCode      string     `gorm:"column:SecurityCode;size:512;not null"`
	IntegrityHash     string     `gorm:"column:IntegrityHash;size:128;not null"`
	PasswordChangedAt time.Time  `gorm:"column:PasswordChangedAt;not null"`
	LastLoginAt       *time.Time `gorm:"column:LastLoginAt"`
	TwoFactorEnabled  bool       `gorm:"column:TwoFactorEnabled;not null;default:0"`
	TwoFactorSecret   *string    `gorm:"column:TwoFactorSecret;size:512"`
	CreatedAt         time.Time  `gorm:"column:CreatedAt;not null"`
	UpdatedAt         time.Time  `gorm:"column:UpdatedAt;not null"`
	Role              Role       `gorm:"foreignKey:RoleID"`
}

// TableName نام جدول در دیتابیس.
func (User) TableName() string {
	return "Users"
}

// Role مدل جدول نقش‌ها.
type Role struct {
	ID            int       `gorm:"column:ID;primaryKey;autoIncrement"`
	Name          string    `gorm:"column:Name;uniqueIndex;size:100;not null"`
	Description   string    `gorm:"column:Description;size:255"`
	IntegrityHash string    `gorm:"column:IntegrityHash;size:128;not null"`
	CreatedAt     time.Time `gorm:"column:CreatedAt;not null"`
}

// TableName نام جدول در دیتابیس.
func (Role) TableName() string {
	return "Roles"
}

// Permission مدل جدول مجوزها.
type Permission struct {
	ID            int       `gorm:"column:ID;primaryKey;autoIncrement"`
	Name          string    `gorm:"column:Name;uniqueIndex;size:100;not null"`
	Description   string    `gorm:"column:Description;size:255"`
	IntegrityHash string    `gorm:"column:IntegrityHash;size:128;not null"`
	CreatedAt     time.Time `gorm:"column:CreatedAt;not null"`
}

// TableName نام جدول در دیتابیس.
func (Permission) TableName() string {
	return "Permissions"
}

// RolePermission ارتباط نقش و مجوز.
type RolePermission struct {
	RoleID        int    `gorm:"column:RoleID;primaryKey"`
	PermissionID  int    `gorm:"column:PermissionID;primaryKey"`
	IntegrityHash string `gorm:"column:IntegrityHash;size:128;not null"`
}

// TableName نام جدول در دیتابیس.
func (RolePermission) TableName() string {
	return "RolePermissions"
}

// ToSensitiveData تبدیل به ساختار حساس برای SecurityCode.
func (u *User) ToSensitiveData() encryption.SensitiveUserData {
	medicalCode := ""
	if u.MedicalCode != nil {
		medicalCode = *u.MedicalCode
	}
	return encryption.SensitiveUserData{
		Name:         u.Name,
		Family:       u.Family,
		PhoneNumber:  u.PhoneNumber,
		PasswordHash: u.PasswordHash,
		Address:      u.Address,
		RoleID:       u.RoleID,
		MedicalCode:  medicalCode,
	}
}
