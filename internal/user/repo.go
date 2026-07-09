package user

import (
	"strconv"

	"github.com/tpdenta/afta-reception/internal/platform/security/integrity"
	"gorm.io/gorm"
)

// Repository اینترفیس CRUD خام کاربران.
type Repository interface {
	Create(user *User) error
	FindByID(id int) (*User, error)
	FindByUsername(username string) (*User, error)
	Update(user *User) error
	Delete(id int) error
	FindAll() ([]User, error)
	FindRoleByID(id int) (*Role, error)
	FindRoleByName(name string) (*Role, error)
	FindAllRoles() ([]Role, error)
	UpdateRoleIntegrity(role *Role) error
	FindAllPermissions() ([]Permission, error)
}

type gormRepository struct {
	db *gorm.DB
}

// NewRepository نمونه Repository می‌سازد.
func NewRepository(db *gorm.DB) Repository {
	return &gormRepository{db: db}
}

func (r *gormRepository) Create(user *User) error {
	return r.db.Create(user).Error
}

func (r *gormRepository) FindByID(id int) (*User, error) {
	var u User
	err := r.db.Preload("Role").Where("ID = ?", id).First(&u).Error
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *gormRepository) FindByUsername(username string) (*User, error) {
	var u User
	err := r.db.Preload("Role").Where("Username = ?", username).First(&u).Error
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *gormRepository) Update(user *User) error {
	return r.db.Save(user).Error
}

func (r *gormRepository) Delete(id int) error {
	return r.db.Where("ID = ?", id).Delete(&User{}).Error
}

func (r *gormRepository) FindAll() ([]User, error) {
	var users []User
	err := r.db.Preload("Role").Find(&users).Error
	return users, err
}

func (r *gormRepository) FindRoleByID(id int) (*Role, error) {
	var role Role
	err := r.db.Where("ID = ?", id).First(&role).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *gormRepository) FindRoleByName(name string) (*Role, error) {
	var role Role
	err := r.db.Where("Name = ?", name).First(&role).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *gormRepository) FindAllRoles() ([]Role, error) {
	var roles []Role
	err := r.db.Find(&roles).Error
	return roles, err
}

func (r *gormRepository) UpdateRoleIntegrity(role *Role) error {
	return r.db.Save(role).Error
}

func (r *gormRepository) FindAllPermissions() ([]Permission, error) {
	var perms []Permission
	err := r.db.Find(&perms).Error
	return perms, err
}

// BuildUserIntegrityFields فیلدهای HMAC کاربر را برمی‌گرداند.
func BuildUserIntegrityFields(u *User) []string {
	medicalCode := ""
	if u.MedicalCode != nil {
		medicalCode = *u.MedicalCode
	}
	twoFactorSecret := ""
	if u.TwoFactorSecret != nil {
		twoFactorSecret = *u.TwoFactorSecret
	}
	return []string{
		strconv.Itoa(u.ID),
		u.Username,
		u.PasswordHash,
		u.Address,
		u.Name,
		u.Family,
		u.PhoneNumber,
		medicalCode,
		strconv.Itoa(u.RoleID),
		strconv.FormatBool(u.IsActive),
		strconv.FormatBool(u.IsLocked),
		u.SecurityCode,
		twoFactorSecret,
		strconv.FormatBool(u.TwoFactorEnabled),
	}
}

// BuildRoleIntegrityFields فیلدهای HMAC نقش را برمی‌گرداند.
func BuildRoleIntegrityFields(role *Role) []string {
	return []string{
		strconv.Itoa(role.ID),
		role.Name,
		role.Description,
	}
}

// SignUserIntegrityHash هش یکپارچگی کاربر را محاسبه می‌کند.
func SignUserIntegrityHash(signer *integrity.Signer, u *User) string {
	return signer.Sign(BuildUserIntegrityFields(u)...)
}

// SignRoleIntegrityHash هش یکپارچگی نقش را محاسبه می‌کند.
func SignRoleIntegrityHash(signer *integrity.Signer, role *Role) string {
	return signer.Sign(BuildRoleIntegrityFields(role)...)
}

// VerifyUserIntegrity یکپارچگی کاربر را بررسی می‌کند.
func VerifyUserIntegrity(signer *integrity.Signer, u *User) bool {
	return signer.Verify(u.IntegrityHash, BuildUserIntegrityFields(u)...)
}

// VerifyRoleIntegrity یکپارچگی نقش را بررسی می‌کند.
func VerifyRoleIntegrity(signer *integrity.Signer, role *Role) bool {
	return signer.Verify(role.IntegrityHash, BuildRoleIntegrityFields(role)...)
}
