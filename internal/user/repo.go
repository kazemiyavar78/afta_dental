package user

import (
	"sort"
	"strconv"
	"strings"

	"github.com/tpdenta/afta-reception/internal/platform/security/integrity"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Repository اینترفیس CRUD خام کاربران، نقش‌ها و مجوزها.
type Repository interface {
	Create(user *User) error
	FindByID(id int) (*User, error)
	FindByUsername(username string) (*User, error)
	Update(user *User) error
	Delete(id int) error
	FindAll() ([]User, error)
	CountUsersByRoleID(roleID int) (int64, error)
	FindDoctors() ([]User, error)
	FindAssistants() ([]User, error)

	// role
	CreateRole(role *Role) error
	FindRoleByID(id int) (*Role, error)
	FindRoleByName(name string) (*Role, error)
	FindAllRoles() ([]Role, error)
	UpdateRole(role *Role) error
	DeleteRole(id int) error

	// permission
	FindAllPermissions() ([]Permission, error)
	CreatePermission(permission *Permission) error
	UpdatePermission(permission *Permission) error
	FindPermissionByName(name string) (*Permission, error)
	FindPermissionsByIDs(ids []int) ([]Permission, error)

	// role permission
	ReplaceRolePermissions(roleID int, rolePermissions []RolePermission) error
	FindRolePermissionByRoleID(roleID int) ([]RolePermission, error)
	FindPermissionIDsByRoleID(roleID int) ([]int, error)
	FindPermissionsByRoleID(roleID int) ([]Permission, error)
	DeleteRolePermissionsByRoleID(roleID int) error
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

// Update فقط فیلدهای خود کاربر را ذخیره می‌کند و associationها (مثل Role از پیش‌بارگذاری‌شده)
// را عمداً نادیده می‌گیرد تا GORM مقدار RoleID را با Role.ID قدیمی بازنویسی نکند.
func (r *gormRepository) Update(user *User) error {
	return r.db.Omit(clause.Associations).Save(user).Error
}

func (r *gormRepository) Delete(id int) error {
	return r.db.Where("ID = ?", id).Delete(&User{}).Error
}

func (r *gormRepository) FindAll() ([]User, error) {
	var users []User
	err := r.db.Preload("Role").Find(&users).Error
	return users, err
}

// CountUsersByRoleID تعداد کاربران دارای نقش مشخص را برمی‌گرداند.
func (r *gormRepository) CountUsersByRoleID(roleID int) (int64, error) {
	var count int64
	err := r.db.Model(&User{}).Where("RoleID = ?", roleID).Count(&count).Error
	return count, err
}

func (r *gormRepository) FindDoctors() ([]User, error) {
	var users []User
	err := r.db.Where("UserType = ?", UserTypeDoctor).Find(&users).Error
	return users, err
}

func (r *gormRepository) FindAssistants() ([]User, error) {
	var users []User
	err := r.db.Where("UserType = ?", UserTypeAssistant).Find(&users).Error
	return users, err
}

// CreateRole نقش جدید را ذخیره می‌کند.
func (r *gormRepository) CreateRole(role *Role) error {
	return r.db.Create(role).Error
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

// UpdateRole نقش را بروزرسانی می‌کند.
func (r *gormRepository) UpdateRole(role *Role) error {
	return r.db.Save(role).Error
}

// DeleteRole نقش را حذف می‌کند.
func (r *gormRepository) DeleteRole(id int) error {
	return r.db.Where("ID = ?", id).Delete(&Role{}).Error
}

func (r *gormRepository) FindAllPermissions() ([]Permission, error) {
	var perms []Permission
	err := r.db.Order("Name ASC").Find(&perms).Error
	return perms, err
}

// CreatePermission مجوز جدید را ذخیره می‌کند.
func (r *gormRepository) CreatePermission(permission *Permission) error {
	return r.db.Create(permission).Error
}

// UpdatePermission مجوز را بروزرسانی می‌کند.
func (r *gormRepository) UpdatePermission(permission *Permission) error {
	return r.db.Save(permission).Error
}

func (r *gormRepository) FindPermissionByName(name string) (*Permission, error) {
	var permission Permission
	err := r.db.Where("Name = ?", name).First(&permission).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &permission, nil
}

// FindPermissionsByIDs مجوزها را با لیست شناسه برمی‌گرداند.
func (r *gormRepository) FindPermissionsByIDs(ids []int) ([]Permission, error) {
	if len(ids) == 0 {
		return []Permission{}, nil
	}
	var perms []Permission
	err := r.db.Where("ID IN ?", ids).Find(&perms).Error
	return perms, err
}

// ReplaceRolePermissions تمام مجوزهای نقش را جایگزین می‌کند.
func (r *gormRepository) ReplaceRolePermissions(roleID int, rolePermissions []RolePermission) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("RoleID = ?", roleID).Delete(&RolePermission{}).Error; err != nil {
			return err
		}
		if len(rolePermissions) == 0 {
			return nil
		}
		return tx.Create(&rolePermissions).Error
	})
}

// FindRolePermissionByRoleID ارتباط‌های نقش-مجوز یک نقش را برمی‌گرداند.
func (r *gormRepository) FindRolePermissionByRoleID(roleID int) ([]RolePermission, error) {
	var rps []RolePermission
	err := r.db.Where("RoleID = ?", roleID).Find(&rps).Error
	return rps, err
}

// FindPermissionIDsByRoleID شناسه مجوزهای یک نقش را برمی‌گرداند.
func (r *gormRepository) FindPermissionIDsByRoleID(roleID int) ([]int, error) {
	var ids []int
	err := r.db.Model(&RolePermission{}).Where("RoleID = ?", roleID).Pluck("PermissionID", &ids).Error
	return ids, err
}

// FindPermissionsByRoleID مجوزهای یک نقش را برمی‌گرداند.
func (r *gormRepository) FindPermissionsByRoleID(roleID int) ([]Permission, error) {
	ids, err := r.FindPermissionIDsByRoleID(roleID)
	if err != nil {
		return nil, err
	}
	perms, err := r.FindPermissionsByIDs(ids)
	if err != nil {
		return nil, err
	}
	sort.Slice(perms, func(i, j int) bool {
		return perms[i].Name < perms[j].Name
	})
	return perms, nil
}

// DeleteRolePermissionsByRoleID تمام مجوزهای منتسب به نقش را حذف می‌کند.
func (r *gormRepository) DeleteRolePermissionsByRoleID(roleID int) error {
	return r.db.Where("RoleID = ?", roleID).Delete(&RolePermission{}).Error
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

// BuildRoleIntegrityFields فیلدهای HMAC نقش شامل مجوزهای منتسب را برمی‌گرداند.
func BuildRoleIntegrityFields(role *Role, permissionIDs []int) []string {
	sorted := append([]int(nil), permissionIDs...)
	sort.Ints(sorted)
	parts := make([]string, len(sorted))
	for i, id := range sorted {
		parts[i] = strconv.Itoa(id)
	}
	return []string{
		strconv.Itoa(role.ID),
		role.Name,
		role.Description,
		strings.Join(parts, ","),
	}
}

// BuildPermissionIntegrityFields فیلدهای HMAC مجوز را برمی‌گرداند.
func BuildPermissionIntegrityFields(p *Permission) []string {
	return []string{
		strconv.Itoa(p.ID),
		p.Name,
		p.Description,
	}
}

// BuildRolePermissionIntegrityFields فیلدهای HMAC ارتباط نقش-مجوز را برمی‌گرداند.
func BuildRolePermissionIntegrityFields(rp *RolePermission) []string {
	return []string{
		strconv.Itoa(rp.RoleID),
		strconv.Itoa(rp.PermissionID),
	}
}

// SignUserIntegrityHash هش یکپارچگی کاربر را محاسبه می‌کند.
func SignUserIntegrityHash(signer *integrity.Signer, u *User) string {
	return signer.Sign(BuildUserIntegrityFields(u)...)
}

// SignRoleIntegrityHash هش یکپارچگی نقش را با لیست مجوزها محاسبه می‌کند.
func SignRoleIntegrityHash(signer *integrity.Signer, role *Role, permissionIDs []int) string {
	return signer.Sign(BuildRoleIntegrityFields(role, permissionIDs)...)
}

// SignPermissionIntegrityHash هش یکپارچگی مجوز را محاسبه می‌کند.
func SignPermissionIntegrityHash(signer *integrity.Signer, p *Permission) string {
	return signer.Sign(BuildPermissionIntegrityFields(p)...)
}

// SignRolePermissionIntegrityHash هش یکپارچگی ارتباط نقش-مجوز را محاسبه می‌کند.
func SignRolePermissionIntegrityHash(signer *integrity.Signer, rp *RolePermission) string {
	return signer.Sign(BuildRolePermissionIntegrityFields(rp)...)
}

// VerifyUserIntegrity یکپارچگی کاربر را بررسی می‌کند.
func VerifyUserIntegrity(signer *integrity.Signer, u *User) bool {
	return signer.Verify(u.IntegrityHash, BuildUserIntegrityFields(u)...)
}

// VerifyRoleIntegrity یکپارچگی نقش را با مجوزهای منتسب بررسی می‌کند.
func VerifyRoleIntegrity(signer *integrity.Signer, role *Role, permissionIDs []int) bool {
	return signer.Verify(role.IntegrityHash, BuildRoleIntegrityFields(role, permissionIDs)...)
}

// VerifyRoleIntegrityLegacy یکپارچگی نقش را با فرمت قدیمی (بدون مجوزها) بررسی می‌کند — فقط برای مهاجرت.
func VerifyRoleIntegrityLegacy(signer *integrity.Signer, role *Role) bool {
	return signer.Verify(role.IntegrityHash, strconv.Itoa(role.ID), role.Name, role.Description)
}

// VerifyPermissionIntegrity یکپارچگی مجوز را بررسی می‌کند.
func VerifyPermissionIntegrity(signer *integrity.Signer, p *Permission) bool {
	return signer.Verify(p.IntegrityHash, BuildPermissionIntegrityFields(p)...)
}
