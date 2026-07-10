package user

import (
	"fmt"
	"time"

	"github.com/tpdenta/afta-reception/internal/platform/apperror"
	"github.com/tpdenta/afta-reception/internal/platform/security/audit"
	"github.com/tpdenta/afta-reception/internal/platform/security/encryption"
	"github.com/tpdenta/afta-reception/internal/platform/security/integrity"
	"github.com/tpdenta/afta-reception/internal/platform/security/loginguard"
	"github.com/tpdenta/afta-reception/internal/platform/security/passwordpolicy"
	"github.com/tpdenta/afta-reception/internal/platform/security/session"
	"github.com/tpdenta/afta-reception/internal/platform/security/settings"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// Service سرویس منطق کسب‌وکار کاربران.
type Service struct {
	repo         Repository
	signer       *integrity.Signer
	encryptSvc   *encryption.UserEncryptionService
	settings     *settings.Service
	audit        *audit.Manager
	session      *session.Service
	loginGuard   *loginguard.Guard
	pwdValidator *passwordpolicy.Validator
}

// NewService نمونه Service می‌سازد.
func NewService(
	db *gorm.DB,
	signer *integrity.Signer,
	encryptSvc *encryption.UserEncryptionService,
	settingsSvc *settings.Service,
	auditMgr *audit.Manager,
	sessionSvc *session.Service,
	loginGuard *loginguard.Guard,
) *Service {
	return &Service{
		repo:         NewRepository(db),
		signer:       signer,
		encryptSvc:   encryptSvc,
		settings:     settingsSvc,
		audit:        auditMgr,
		session:      sessionSvc,
		loginGuard:   loginGuard,
		pwdValidator: passwordpolicy.NewValidator(settingsSvc),
	}
}

// CreateUser کاربر جدید ایجاد می‌کند.
func (s *Service) CreateUser(req CreateUserRequest, actorUserID int, ip string) (*UserResponse, error) {
	valid, errs := s.pwdValidator.IsValidPassword(req.Password)
	if !valid {
		return nil, apperror.New("WEAK_PASSWORD", errs[0], "password policy violation", 400)
	}

	_, err := s.repo.FindRoleByID(req.RoleID)
	if err != nil {
		return nil, apperror.New("INVALID_ROLE", "نقش انتخاب‌شده معتبر نیست.", err.Error(), 400)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, apperror.New("HASH_ERROR", "خطا در پردازش رمز عبور.", err.Error(), 500)
	}

	now := time.Now().UTC()
	user := &User{
		Username:          req.Username,
		PasswordHash:      string(hash),
		Address:           req.Address,
		Name:              req.Name,
		Family:            req.Family,
		PhoneNumber:       req.PhoneNumber,
		MedicalCode:       req.MedicalCode,
		RoleID:            req.RoleID,
		IsActive:          true,
		IsLocked:          false,
		PasswordChangedAt: now,
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	securityCode, err := s.encryptSvc.CreateSecurityCode(user.ToSensitiveData())
	if err != nil {
		return nil, apperror.New("ENCRYPT_ERROR", "خطا در تولید کد امنیتی.", err.Error(), 500)
	}
	user.SecurityCode = securityCode
	user.IntegrityHash = SignUserIntegrityHash(s.signer, user)

	if err := s.repo.Create(user); err != nil {
		return nil, apperror.New("DB_ERROR", "خطا در ایجاد کاربر.", err.Error(), 500)
	}

	// پس از Create، ID موجود است — IntegrityHash را با ID واقعی به‌روز کن
	created, _ := s.repo.FindByID(user.ID)
	if created != nil {
		created.IntegrityHash = SignUserIntegrityHash(s.signer, created)
		_ = s.repo.Update(created)
		user = created
	}

	uid := actorUserID
	_ = s.audit.LogEvent(&uid, ip, audit.EventUserDataChange,
		fmt.Sprintf("ایجاد کاربر %s", user.Username))

	return s.toResponse(user), nil
}

// HandleUserLogin ورود کاربر را مدیریت می‌کند.
func (s *Service) HandleUserLogin(req LoginRequest, ip, browser string) (*User, *session.Session, error) {
	if err := s.loginGuard.CheckIPLock(ip); err != nil {
		return nil, nil, err
	}

	user, err := s.repo.FindByUsername(req.Username)
	if err == gorm.ErrRecordNotFound {
		_ = s.loginGuard.RecordFailedLogin(ip)
		_ = s.audit.LogEvent(nil, ip, audit.EventLoginFailed,
			fmt.Sprintf("ورود ناموفق — کاربر %s یافت نشد", req.Username))
		return nil, nil, apperror.ErrInvalidCreds
	}
	if err != nil {
		return nil, nil, apperror.New("DB_ERROR", "خطا در جستجوی کاربر.", err.Error(), 500)
	}

	if !VerifyUserIntegrity(s.signer, user) {
		_ = s.audit.LogEvent(&user.ID, ip, audit.EventDataTampering,
			fmt.Sprintf("دستکاری داده کاربر %s", user.Username))
		return nil, nil, apperror.ErrIntegrity
	}

	if user.IsLocked || !user.IsActive {
		_ = s.audit.LogEvent(&user.ID, ip, audit.EventLoginFailed,
			fmt.Sprintf("ورود ناموفق — حساب %s قفل یا غیرفعال", user.Username))
		return nil, nil, apperror.ErrAccountLocked
	}

	roleName := user.Role.Name
	if err := s.loginGuard.CheckWorkHours(roleName); err != nil {
		_ = s.audit.LogEvent(&user.ID, ip, audit.EventLoginFailed,
			fmt.Sprintf("ورود خارج ساعات کاری — %s", user.Username))
		return nil, nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		_ = s.loginGuard.RecordFailedLogin(ip)
		_ = s.audit.LogEvent(&user.ID, ip, audit.EventLoginFailed,
			fmt.Sprintf("رمز اشتباه — %s", user.Username))
		return nil, nil, apperror.ErrInvalidCreds
	}

	if !s.encryptSvc.CheckUserSecurityCode(user.ToSensitiveData(), user.SecurityCode) {
		_ = s.audit.LogEvent(&user.ID, ip, audit.EventDataTampering,
			fmt.Sprintf("SecurityCode نامعتبر — %s", user.Username))
		return nil, nil, apperror.ErrIntegrity
	}

	_ = s.loginGuard.ResetLoginAttempts(ip)

	sess, err := s.session.CreateSession(user.ID, ip, browser)
	if err != nil {
		return nil, nil, err
	}

	now := time.Now().UTC()
	user.LastLoginAt = &now
	_ = s.repo.Update(user)

	_ = s.audit.LogEvent(&user.ID, ip, audit.EventLoginSuccess,
		fmt.Sprintf("ورود موفق — %s", user.Username))

	return user, sess, nil
}

// HandleLogout خروج کاربر را مدیریت می‌کند.
func (s *Service) HandleLogout(sessionID string, userID int, ip string) error {
	_ = s.audit.LogEvent(&userID, ip, audit.EventLogout,
		fmt.Sprintf("خروج کاربر ID=%d", userID))
	return nil
}

// GetUserByID کاربر را با شناسه برمی‌گرداند.
func (s *Service) GetUserByID(id int, actorUserID int, actorRole string) (*UserResponse, error) {
	if actorRole != "Admin" && actorUserID != id {
		return nil, apperror.ErrForbidden
	}

	user, err := s.repo.FindByID(id)
	if err == gorm.ErrRecordNotFound {
		return nil, apperror.ErrNotFound
	}
	if err != nil {
		return nil, apperror.New("DB_ERROR", "خطا در خواندن کاربر.", err.Error(), 500)
	}

	if !VerifyUserIntegrity(s.signer, user) {
		return nil, apperror.ErrIntegrity
	}

	return s.toResponse(user), nil
}

// دریافت سشن های کاربر
func (s *Service) GetUserSessions(userID int) ([]session.Session, error) {
	sessions, err := s.session.GetSessionsByUserID(userID)
	if err != nil {
		return nil, apperror.New("DB_ERROR", "خطا در خواندن سشن ها.", err.Error(), 500)
	}
	return sessions, nil
}

// ListUsers لیست کاربران (فقط Admin).
func (s *Service) ListUsers(actorRole string) ([]UserResponse, error) {
	if actorRole != "Admin" {
		return nil, apperror.ErrForbidden
	}

	users, err := s.repo.FindAll()
	if err != nil {
		return nil, apperror.New("DB_ERROR", "خطا در خواندن کاربران.", err.Error(), 500)
	}

	var result []UserResponse
	for _, u := range users {
		result = append(result, *s.toResponse(&u))
	}
	return result, nil
}

// UpdateUser کاربر را به‌روزرسانی می‌کند.
func (s *Service) UpdateUser(id int, req UpdateUserRequest, actorUserID int, actorRole, ip string) (*UserResponse, error) {
	if actorRole != "Admin" && actorUserID != id {
		return nil, apperror.ErrForbidden
	}

	user, err := s.repo.FindByID(id)
	if err == gorm.ErrRecordNotFound {
		return nil, apperror.ErrNotFound
	}
	if err != nil {
		return nil, apperror.New("DB_ERROR", "خطا در خواندن کاربر.", err.Error(), 500)
	}

	if req.Address != nil {
		user.Address = *req.Address
	}
	if req.Name != nil {
		user.Name = *req.Name
	}
	if req.Family != nil {
		user.Family = *req.Family
	}
	if req.PhoneNumber != nil {
		user.PhoneNumber = *req.PhoneNumber
	}
	if req.MedicalCode != nil {
		user.MedicalCode = req.MedicalCode
	}
	if req.RoleID != nil {
		if actorRole != "Admin" {
			return nil, apperror.ErrForbidden
		}
		user.RoleID = *req.RoleID
	}
	if req.IsActive != nil && actorRole == "Admin" {
		user.IsActive = *req.IsActive
	}

	user.UpdatedAt = time.Now().UTC()

	securityCode, err := s.encryptSvc.CreateSecurityCode(user.ToSensitiveData())
	if err != nil {
		return nil, apperror.New("ENCRYPT_ERROR", "خطا در به‌روزرسانی کد امنیتی.", err.Error(), 500)
	}
	user.SecurityCode = securityCode
	user.IntegrityHash = SignUserIntegrityHash(s.signer, user)

	if err := s.repo.Update(user); err != nil {
		return nil, apperror.New("DB_ERROR", "خطا در به‌روزرسانی کاربر.", err.Error(), 500)
	}

	uid := actorUserID
	_ = s.audit.LogEvent(&uid, ip, audit.EventUserDataChange,
		fmt.Sprintf("به‌روزرسانی کاربر %s", user.Username))

	updated, _ := s.repo.FindByID(id)
	return s.toResponse(updated), nil
}

// ChangePassword رمز عبور کاربر را تغییر می‌دهد.
func (s *Service) ChangePassword(userID int, req ChangePasswordRequest, ip string) error {
	user, err := s.repo.FindByID(userID)
	if err != nil {
		return apperror.ErrNotFound
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.OldPassword)); err != nil {
		return apperror.New("INVALID_PASSWORD", "رمز عبور فعلی اشتباه است.", "old password mismatch", 400)
	}

	valid, errs := s.pwdValidator.IsValidPassword(req.NewPassword)
	if !valid {
		return apperror.New("WEAK_PASSWORD", errs[0], "password policy violation", 400)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return apperror.New("HASH_ERROR", "خطا در پردازش رمز عبور.", err.Error(), 500)
	}

	user.PasswordHash = string(hash)
	user.PasswordChangedAt = time.Now().UTC()
	user.UpdatedAt = time.Now().UTC()

	securityCode, err := s.encryptSvc.CreateSecurityCode(user.ToSensitiveData())
	if err != nil {
		return apperror.New("ENCRYPT_ERROR", "خطا در به‌روزرسانی کد امنیتی.", err.Error(), 500)
	}
	user.SecurityCode = securityCode
	user.IntegrityHash = SignUserIntegrityHash(s.signer, user)

	if err := s.repo.Update(user); err != nil {
		return apperror.New("DB_ERROR", "خطا در تغییر رمز.", err.Error(), 500)
	}

	_ = s.audit.LogEvent(&userID, ip, audit.EventUserDataChange, "تغییر رمز عبور")
	return nil
}

// GetRoleName نام نقش کاربر را برمی‌گرداند.
func (s *Service) GetRoleName(userID int) (string, error) {
	user, err := s.repo.FindByID(userID)
	if err != nil {
		return "", err
	}
	return user.Role.Name, nil
}

// CheckCurrentUserSecurityCode SecurityCode کاربر فعلی را بررسی می‌کند.
func (s *Service) CheckCurrentUserSecurityCode(userID int) (bool, error) {
	user, err := s.repo.FindByID(userID)
	if err != nil {
		return false, apperror.ErrNotFound
	}
	return s.encryptSvc.CheckUserSecurityCode(user.ToSensitiveData(), user.SecurityCode), nil
}

// VerifyAllUsers یکپارچگی تمام کاربران را بررسی می‌کند.
func (s *Service) VerifyAllUsers(ip string) (bool, error) {
	users, err := s.repo.FindAll()
	if err != nil {
		return false, err
	}
	for _, u := range users {
		if !VerifyUserIntegrity(s.signer, &u) {
			_ = s.audit.LogEvent(&u.ID, ip, audit.EventDataTampering,
				fmt.Sprintf("دستکاری کاربر %s", u.Username))
			return false, nil
		}
	}
	return true, nil
}

// VerifyAllRoles یکپارچگی تمام نقش‌ها را بررسی می‌کند.
func (s *Service) VerifyAllRoles(ip string) (bool, error) {
	roles, err := s.repo.FindAllRoles()
	if err != nil {
		return false, err
	}
	for _, r := range roles {
		if !VerifyRoleIntegrity(s.signer, &r) {
			_ = s.audit.LogEvent(nil, ip, audit.EventDataTampering,
				fmt.Sprintf("دستکاری نقش %s", r.Name))
			return false, nil
		}
	}
	return true, nil
}

// FixRoleIntegrityHashes هش یکپارچگی نقش‌های بدون هش را اصلاح می‌کند.
func (s *Service) FixRoleIntegrityHashes() error {
	roles, err := s.repo.FindAllRoles()
	if err != nil {
		return err
	}
	for _, r := range roles {
		if r.IntegrityHash == "" {
			r.IntegrityHash = SignRoleIntegrityHash(s.signer, &r)
			if err := s.repo.UpdateRoleIntegrity(&r); err != nil {
				return err
			}
		}
	}
	return nil
}

// GetMe اطلاعات کاربر فعلی همراه نقش و مجوزها.
func (s *Service) GetMe(userID int) (*MeResponse, error) {
	user, err := s.repo.FindByID(userID)
	if err == gorm.ErrRecordNotFound {
		return nil, apperror.ErrNotFound
	}
	if err != nil {
		return nil, apperror.New("DB_ERROR", "خطا در خواندن کاربر.", err.Error(), 500)
	}

	if !VerifyUserIntegrity(s.signer, user) {
		return nil, apperror.ErrIntegrity
	}

	return &MeResponse{
		User:        *s.toResponse(user),
		Permissions: GetPermissionsForRole(user.Role.Name),
	}, nil
}

// ListRoles لیست نقش‌ها را برمی‌گرداند.
func (s *Service) ListRoles() ([]RoleResponse, error) {
	roles, err := s.repo.FindAllRoles()
	if err != nil {
		return nil, apperror.New("DB_ERROR", "خطا در خواندن نقش‌ها.", err.Error(), 500)
	}
	var result []RoleResponse
	for _, r := range roles {
		result = append(result, RoleResponse{ID: r.ID, Name: r.Name})
	}
	return result, nil
}

// UpdateSecuritySetting تنظیم امنیتی را به‌روزرسانی می‌کند (فقط Admin).
func (s *Service) UpdateSecuritySetting(name, value string, actorUserID int, actorRole, ip string) error {
	if actorRole != "Admin" {
		return apperror.ErrForbidden
	}
	return s.settings.UpdateSetting(name, value, actorUserID, ip)
}

// GetSecuritySettings تنظیمات امنیتی را برمی‌گرداند (فقط Admin).
func (s *Service) GetSecuritySettings(actorRole string) ([]settings.SecuritySetting, error) {
	if actorRole != "Admin" {
		return nil, apperror.ErrForbidden
	}
	return s.settings.GetAllSettings()
}

func (s *Service) toResponse(user *User) *UserResponse {
	resp := &UserResponse{
		ID:          user.ID,
		Username:    user.Username,
		Address:     user.Address,
		Name:        user.Name,
		Family:      user.Family,
		PhoneNumber: user.PhoneNumber,
		MedicalCode: user.MedicalCode,
		RoleID:      user.RoleID,
		RoleName:    user.Role.Name,
		IsActive:    user.IsActive,
		IsLocked:    user.IsLocked,
	}
	if user.LastLoginAt != nil {
		t := user.LastLoginAt.Format(time.RFC3339)
		resp.LastLoginAt = &t
	}
	return resp
}
