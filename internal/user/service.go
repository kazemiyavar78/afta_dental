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

	if err := s.ensureAssignableRole(req.RoleID, actorUserID, ip); err != nil {
		return nil, err
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
		UserType:          UserTypeUser,
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

	if err := s.verifyRoleIntegrityForUser(&user.Role, user.ID, ip); err != nil {
		return nil, nil, err
	}

	if user.IsLocked || !user.IsActive {
		_ = s.audit.LogEvent(&user.ID, ip, audit.EventLoginFailed,
			fmt.Sprintf("ورود ناموفق — حساب %s قفل یا غیرفعال", user.Username))
		return nil, nil, apperror.ErrAccountLocked
	}

	rolePerms, err := s.repo.FindPermissionsByRoleID(user.RoleID)
	if err != nil {
		return nil, nil, apperror.New("DB_ERROR", "خطا در خواندن مجوزها.", err.Error(), 500)
	}
	permNames := make([]string, 0, len(rolePerms))
	for _, p := range rolePerms {
		permNames = append(permNames, p.Name)
	}
	if err := s.loginGuard.CheckWorkHours(HasAllPermissions(permNames)); err != nil {
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
func (s *Service) GetUserByID(id int, actorUserID int, isAdmin bool) (*UserResponse, error) {
	if !isAdmin && actorUserID != id {
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
		// ادمین باید بتواند کاربر خراب‌شده را ببیند و با ذخیرهٔ مجدد تعمیر کند
		if !isAdmin {
			return nil, apperror.ErrIntegrity
		}
		_ = s.audit.LogEvent(&actorUserID, "", audit.EventDataTampering,
			fmt.Sprintf("خواندن کاربر با یکپارچگی نقض‌شده ID=%d توسط ادمین", user.ID))
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

// ListUsers لیست کاربران را برمی‌گرداند.
func (s *Service) ListUsers() ([]UserResponse, error) {
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

// ListDoctors لیست دکتران
func (s *Service) ListDoctors() ([]UserResponse, error) {
	users, err := s.repo.FindDoctors()
	if err != nil {
		return nil, apperror.New("DB_ERROR", "خطا در خواندن دکتران.", err.Error(), 500)
	}
	result := make([]UserResponse, 0, len(users))
	for _, u := range users {
		result = append(result, *s.toResponse(&u))
	}
	return result, nil
}

// ListAssistants لیست دستیاران
func (s *Service) ListAssistants() ([]UserResponse, error) {
	users, err := s.repo.FindAssistants()
	if err != nil {
		return nil, apperror.New("DB_ERROR", "خطا در خواندن دستیاران.", err.Error(), 500)
	}
	result := make([]UserResponse, 0, len(users))
	for _, u := range users {
		result = append(result, *s.toResponse(&u))
	}
	return result, nil
}

// UpdateUser کاربر را به‌روزرسانی می‌کند.
func (s *Service) UpdateUser(id int, req UpdateUserRequest, actorUserID int, isAdmin bool, ip string) (*UserResponse, error) {
	if !isAdmin && actorUserID != id {
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
	if req.UserType != nil {
		user.UserType = UserType(*req.UserType)
	}
	if req.RoleID != nil {
		if !isAdmin {
			return nil, apperror.ErrForbidden
		}
		if err := s.ensureAssignableRole(*req.RoleID, actorUserID, ip); err != nil {
			return nil, err
		}
		user.RoleID = *req.RoleID
		// association از Preload را خالی کن تا با RoleID جدید هم‌خوان نباشد
		user.Role = Role{}
	}
	if req.IsActive != nil && isAdmin {
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

// GetAuthPermissions نام مجوزهای کاربر و وضعیت ادمین (داشتن تمام مجوزها) را برمی‌گرداند.
func (s *Service) GetAuthPermissions(userID int) (permissions []string, isAdmin bool, err error) {
	user, err := s.repo.FindByID(userID)
	if err != nil {
		return nil, false, err
	}
	perms, err := s.repo.FindPermissionsByRoleID(user.RoleID)
	if err != nil {
		return nil, false, err
	}
	names := make([]string, 0, len(perms))
	for _, p := range perms {
		names = append(names, p.Name)
	}
	return names, HasAllPermissions(names), nil
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
		permIDs, err := s.repo.FindPermissionIDsByRoleID(r.ID)
		if err != nil {
			return false, err
		}
		if !VerifyRoleIntegrity(s.signer, &r, permIDs) {
			_ = s.audit.LogEvent(nil, ip, audit.EventDataTampering,
				fmt.Sprintf("دستکاری نقش %s", r.Name))
			return false, nil
		}
	}
	return true, nil
}

// SyncPermissions مجوزهای ثابت سیستم را در دیتابیس ایجاد یا بروزرسانی می‌کند.
func (s *Service) SyncPermissions() error {
	now := time.Now().UTC()
	for name, description := range DefaultPermissionCatalog() {
		existing, err := s.repo.FindPermissionByName(name)
		if err != nil {
			return err
		}
		if existing == nil {
			p := &Permission{
				Name:          name,
				Description:   description,
				IntegrityHash: "",
				CreatedAt:     now,
			}
			if err := s.repo.CreatePermission(p); err != nil {
				return err
			}
			p.IntegrityHash = SignPermissionIntegrityHash(s.signer, p)
			if err := s.repo.UpdatePermission(p); err != nil {
				return err
			}
			continue
		}
		existing.Description = description
		existing.IntegrityHash = SignPermissionIntegrityHash(s.signer, existing)
		if err := s.repo.UpdatePermission(existing); err != nil {
			return err
		}
	}
	return s.ensureAdminHasAllPermissions()
}

// ensureAdminHasAllPermissions همه مجوزهای سیستم را به نقش Admin منتسب می‌کند و هش را امضا می‌کند.
func (s *Service) ensureAdminHasAllPermissions() error {
	adminRole, err := s.repo.FindRoleByName("Admin")
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return s.ensureEmptyOperationalRoles()
		}
		return err
	}
	allPerms, err := s.repo.FindAllPermissions()
	if err != nil {
		return err
	}
	permIDs := make([]int, 0, len(allPerms))
	rps := make([]RolePermission, 0, len(allPerms))
	for _, p := range allPerms {
		permIDs = append(permIDs, p.ID)
		rp := RolePermission{RoleID: adminRole.ID, PermissionID: p.ID}
		rp.IntegrityHash = SignRolePermissionIntegrityHash(s.signer, &rp)
		rps = append(rps, rp)
	}
	if err := s.repo.ReplaceRolePermissions(adminRole.ID, rps); err != nil {
		return err
	}
	adminRole.IntegrityHash = SignRoleIntegrityHash(s.signer, adminRole, permIDs)
	if err := s.repo.UpdateRole(adminRole); err != nil {
		return err
	}
	return s.ensureEmptyOperationalRoles()
}

// ensureEmptyOperationalRoles برای نقش‌های Reception/Doctor بدون مجوز، مجموعه پیش‌فرض عملیاتی می‌گذارد.
func (s *Service) ensureEmptyOperationalRoles() error {
	defaults := map[string][]string{
		"Reception": {
			"reception.read", "reception.create", "reception.update", "reception.delete",
			"patient.read", "patient.create", "patient.update",
			"organization.read", "organization_packages.read",
			"services.read", "fund.read",
			"wallet.read", "wallet.cash", "wallet.card_to_card", "wallet.card_reader",
			"bank_account.read",
			"tariff.read", "special_code.read", "regulation.read",
		},
		"Doctor": {
			"reception.read", "reception.update",
			"patient.read", "services.read",
		},
	}
	for roleName, permNames := range defaults {
		if err := s.assignPermissionsIfRoleEmpty(roleName, permNames); err != nil {
			return err
		}
	}
	return nil
}

// assignPermissionsIfRoleEmpty در صورتی که نقش وجود داشته و هیچ مجوزی نداشته باشد، مجوزهای داده‌شده را می‌دهد.
func (s *Service) assignPermissionsIfRoleEmpty(roleName string, permissionNames []string) error {
	role, err := s.repo.FindRoleByName(roleName)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil
		}
		return err
	}
	existing, err := s.repo.FindPermissionIDsByRoleID(role.ID)
	if err != nil {
		return err
	}
	if len(existing) > 0 {
		return nil
	}

	permIDs := make([]int, 0, len(permissionNames))
	rps := make([]RolePermission, 0, len(permissionNames))
	for _, name := range permissionNames {
		p, err := s.repo.FindPermissionByName(name)
		if err != nil || p == nil {
			continue
		}
		permIDs = append(permIDs, p.ID)
		rp := RolePermission{RoleID: role.ID, PermissionID: p.ID}
		rp.IntegrityHash = SignRolePermissionIntegrityHash(s.signer, &rp)
		rps = append(rps, rp)
	}
	if len(rps) == 0 {
		return nil
	}
	if err := s.repo.ReplaceRolePermissions(role.ID, rps); err != nil {
		return err
	}
	role.IntegrityHash = SignRoleIntegrityHash(s.signer, role, permIDs)
	return s.repo.UpdateRole(role)
}

// FixRoleIntegrityHashes هش نقش‌های بدون هش یا نقش‌های قدیمی بدون مجوز را به فرمت جدید مهاجرت می‌دهد.
func (s *Service) FixRoleIntegrityHashes() error {
	roles, err := s.repo.FindAllRoles()
	if err != nil {
		return err
	}
	for i := range roles {
		r := &roles[i]
		permIDs, err := s.repo.FindPermissionIDsByRoleID(r.ID)
		if err != nil {
			return err
		}
		if VerifyRoleIntegrity(s.signer, r, permIDs) {
			continue
		}
		// فقط نقش‌های بدون هش یا هش قدیمی بدون انتصاب مجوز مهاجرت می‌شوند؛ دستکاری واقعی حفظ می‌شود.
		legacyOK := r.IntegrityHash != "" && VerifyRoleIntegrityLegacy(s.signer, r)
		if r.IntegrityHash != "" && !(legacyOK && len(permIDs) == 0) {
			continue
		}
		r.IntegrityHash = SignRoleIntegrityHash(s.signer, r, permIDs)
		if err := s.repo.UpdateRole(r); err != nil {
			return err
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

	if err := s.verifyRoleIntegrityForUser(&user.Role, user.ID, ""); err != nil {
		return nil, err
	}

	perms, err := s.repo.FindPermissionsByRoleID(user.RoleID)
	if err != nil {
		return nil, apperror.New("DB_ERROR", "خطا در خواندن مجوزها.", err.Error(), 500)
	}
	names := make([]string, 0, len(perms))
	for _, p := range perms {
		names = append(names, p.Name)
	}

	return &MeResponse{
		User:        *s.toResponse(user),
		Permissions: names,
		IsAdmin:     HasAllPermissions(names),
	}, nil
}

// ListAssignableRoles نقش‌های دارای یکپارچگی معتبر را برای انتصاب برمی‌گرداند.
func (s *Service) ListAssignableRoles() ([]RoleResponse, error) {
	roles, err := s.repo.FindAllRoles()
	if err != nil {
		return nil, apperror.New("DB_ERROR", "خطا در خواندن نقش‌ها.", err.Error(), 500)
	}
	var result []RoleResponse
	for i := range roles {
		permIDs, err := s.repo.FindPermissionIDsByRoleID(roles[i].ID)
		if err != nil {
			return nil, apperror.New("DB_ERROR", "خطا در خواندن مجوزهای نقش.", err.Error(), 500)
		}
		if !VerifyRoleIntegrity(s.signer, &roles[i], permIDs) {
			continue
		}
		result = append(result, RoleResponse{ID: roles[i].ID, Name: roles[i].Name})
	}
	return result, nil
}

// ListRoles لیست نقش‌ها را با جزئیات مجوز و وضعیت یکپارچگی برمی‌گرداند.
func (s *Service) ListRoles() ([]RoleDetailResponse, error) {
	roles, err := s.repo.FindAllRoles()
	if err != nil {
		return nil, apperror.New("DB_ERROR", "خطا در خواندن نقش‌ها.", err.Error(), 500)
	}
	result := make([]RoleDetailResponse, 0, len(roles))
	for i := range roles {
		detail, err := s.toRoleDetail(&roles[i])
		if err != nil {
			return nil, err
		}
		result = append(result, *detail)
	}
	return result, nil
}

// GetRole نقش را با شناسه برمی‌گرداند.
func (s *Service) GetRole(id int) (*RoleDetailResponse, error) {
	role, err := s.repo.FindRoleByID(id)
	if err == gorm.ErrRecordNotFound {
		return nil, apperror.ErrNotFound
	}
	if err != nil {
		return nil, apperror.New("DB_ERROR", "خطا در خواندن نقش.", err.Error(), 500)
	}
	return s.toRoleDetail(role)
}

// CreateRole نقش جدید می‌سازد و مجوزها را منتسب می‌کند.
func (s *Service) CreateRole(req CreateRoleRequest, actorID int, ip string) (*RoleDetailResponse, error) {
	if _, err := s.repo.FindRoleByName(req.Name); err == nil {
		return nil, apperror.New("VALIDATION_ERROR", "نقش با این نام از قبل وجود دارد.", "role name exists", 400)
	} else if err != gorm.ErrRecordNotFound {
		return nil, apperror.New("DB_ERROR", "خطا در بررسی نام نقش.", err.Error(), 500)
	}

	perms, err := s.resolvePermissions(req.PermissionIDs)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	role := &Role{
		Name:          req.Name,
		Description:   req.Description,
		IntegrityHash: "",
		CreatedAt:     now,
	}
	if err := s.repo.CreateRole(role); err != nil {
		return nil, apperror.New("DB_ERROR", "خطا در ایجاد نقش.", err.Error(), 500)
	}

	permIDs := make([]int, 0, len(perms))
	rps := make([]RolePermission, 0, len(perms))
	for _, p := range perms {
		permIDs = append(permIDs, p.ID)
		rp := RolePermission{RoleID: role.ID, PermissionID: p.ID}
		rp.IntegrityHash = SignRolePermissionIntegrityHash(s.signer, &rp)
		rps = append(rps, rp)
	}
	if err := s.repo.ReplaceRolePermissions(role.ID, rps); err != nil {
		return nil, apperror.New("DB_ERROR", "خطا در انتصاب مجوزها.", err.Error(), 500)
	}

	role.IntegrityHash = SignRoleIntegrityHash(s.signer, role, permIDs)
	if err := s.repo.UpdateRole(role); err != nil {
		return nil, apperror.New("DB_ERROR", "خطا در ذخیره هش نقش.", err.Error(), 500)
	}

	_ = s.audit.LogEvent(&actorID, ip, audit.EventUserDataChange, fmt.Sprintf("ایجاد نقش %s", role.Name))
	return s.toRoleDetail(role)
}

// UpdateRole پس از تایید یکپارچگی، نقش و مجوزهای آن را بروزرسانی می‌کند.
func (s *Service) UpdateRole(id int, req UpdateRoleRequest, actorID int, ip string) (*RoleDetailResponse, error) {
	role, err := s.repo.FindRoleByID(id)
	if err == gorm.ErrRecordNotFound {
		return nil, apperror.ErrNotFound
	}
	if err != nil {
		return nil, apperror.New("DB_ERROR", "خطا در خواندن نقش.", err.Error(), 500)
	}

	currentPermIDs, err := s.repo.FindPermissionIDsByRoleID(role.ID)
	if err != nil {
		return nil, apperror.New("DB_ERROR", "خطا در خواندن مجوزهای نقش.", err.Error(), 500)
	}
	if !VerifyRoleIntegrity(s.signer, role, currentPermIDs) {
		_ = s.audit.LogEvent(&actorID, ip, audit.EventDataTampering,
			fmt.Sprintf("نقض یکپارچگی نقش %d (%s)", role.ID, role.Name))
		return nil, apperror.ErrIntegrity
	}

	if req.Name != role.Name {
		if existing, err := s.repo.FindRoleByName(req.Name); err == nil && existing.ID != role.ID {
			return nil, apperror.New("VALIDATION_ERROR", "نقش با این نام از قبل وجود دارد.", "role name exists", 400)
		} else if err != nil && err != gorm.ErrRecordNotFound {
			return nil, apperror.New("DB_ERROR", "خطا در بررسی نام نقش.", err.Error(), 500)
		}
	}

	perms, err := s.resolvePermissions(req.PermissionIDs)
	if err != nil {
		return nil, err
	}

	role.Name = req.Name
	role.Description = req.Description

	permIDs := make([]int, 0, len(perms))
	rps := make([]RolePermission, 0, len(perms))
	for _, p := range perms {
		permIDs = append(permIDs, p.ID)
		rp := RolePermission{RoleID: role.ID, PermissionID: p.ID}
		rp.IntegrityHash = SignRolePermissionIntegrityHash(s.signer, &rp)
		rps = append(rps, rp)
	}
	if err := s.repo.ReplaceRolePermissions(role.ID, rps); err != nil {
		return nil, apperror.New("DB_ERROR", "خطا در انتصاب مجوزها.", err.Error(), 500)
	}

	role.IntegrityHash = SignRoleIntegrityHash(s.signer, role, permIDs)
	if err := s.repo.UpdateRole(role); err != nil {
		return nil, apperror.New("DB_ERROR", "خطا در بروزرسانی نقش.", err.Error(), 500)
	}

	_ = s.audit.LogEvent(&actorID, ip, audit.EventUserDataChange, fmt.Sprintf("بروزرسانی نقش %s", role.Name))
	return s.toRoleDetail(role)
}

// DeleteRole پس از تایید یکپارچگی، نقش را حذف می‌کند.
func (s *Service) DeleteRole(id int, actorID int, ip string) error {
	role, err := s.repo.FindRoleByID(id)
	if err == gorm.ErrRecordNotFound {
		return apperror.ErrNotFound
	}
	if err != nil {
		return apperror.New("DB_ERROR", "خطا در خواندن نقش.", err.Error(), 500)
	}
	permIDs, err := s.repo.FindPermissionIDsByRoleID(role.ID)
	if err != nil {
		return apperror.New("DB_ERROR", "خطا در خواندن مجوزهای نقش.", err.Error(), 500)
	}
	if role.Name == "Admin" || s.permissionIDsCoverCatalog(permIDs) {
		return apperror.New("VALIDATION_ERROR", "حذف نقش ادمین (دارای تمام مجوزها) مجاز نیست.", "cannot delete admin role", 400)
	}
	if !VerifyRoleIntegrity(s.signer, role, permIDs) {
		_ = s.audit.LogEvent(&actorID, ip, audit.EventDataTampering,
			fmt.Sprintf("نقض یکپارچگی نقش %d (%s)", role.ID, role.Name))
		return apperror.ErrIntegrity
	}

	count, err := s.repo.CountUsersByRoleID(role.ID)
	if err != nil {
		return apperror.New("DB_ERROR", "خطا در بررسی کاربران نقش.", err.Error(), 500)
	}
	if count > 0 {
		return apperror.New("VALIDATION_ERROR", "نقش دارای کاربر است و قابل حذف نیست.", "role has users", 400)
	}

	if err := s.repo.DeleteRolePermissionsByRoleID(role.ID); err != nil {
		return apperror.New("DB_ERROR", "خطا در حذف مجوزهای نقش.", err.Error(), 500)
	}
	if err := s.repo.DeleteRole(role.ID); err != nil {
		return apperror.New("DB_ERROR", "خطا در حذف نقش.", err.Error(), 500)
	}
	_ = s.audit.LogEvent(&actorID, ip, audit.EventUserDataChange, fmt.Sprintf("حذف نقش %s", role.Name))
	return nil
}

// ListPermissions لیست مجوزهای سیستم را برمی‌گرداند.
func (s *Service) ListPermissions() ([]PermissionResponse, error) {
	perms, err := s.repo.FindAllPermissions()
	if err != nil {
		return nil, apperror.New("DB_ERROR", "خطا در خواندن مجوزها.", err.Error(), 500)
	}
	result := make([]PermissionResponse, 0, len(perms))
	for _, p := range perms {
		result = append(result, PermissionResponse{ID: p.ID, Name: p.Name, Description: p.Description})
	}
	return result, nil
}

// ensureAssignableRole وجود نقش و صحت یکپارچگی آن را قبل از انتصاب بررسی می‌کند.
func (s *Service) ensureAssignableRole(roleID int, actorID int, ip string) error {
	role, err := s.repo.FindRoleByID(roleID)
	if err == gorm.ErrRecordNotFound {
		return apperror.New("INVALID_ROLE", "نقش انتخاب‌شده معتبر نیست.", "role not found", 400)
	}
	if err != nil {
		return apperror.New("DB_ERROR", "خطا در خواندن نقش.", err.Error(), 500)
	}
	permIDs, err := s.repo.FindPermissionIDsByRoleID(role.ID)
	if err != nil {
		return apperror.New("DB_ERROR", "خطا در خواندن مجوزهای نقش.", err.Error(), 500)
	}
	if !VerifyRoleIntegrity(s.signer, role, permIDs) {
		_ = s.audit.LogEvent(&actorID, ip, audit.EventDataTampering,
			fmt.Sprintf("انتصاب نقش دستکاری‌شده %d (%s) رد شد", role.ID, role.Name))
		return apperror.New("INVALID_ROLE", "نقش انتخاب‌شده به دلیل نقض یکپارچگی قابل انتصاب نیست.", "role integrity failed", 400)
	}
	return nil
}

// verifyRoleIntegrityForUser یکپارچگی نقش کاربر را بررسی می‌کند؛ در صورت نقض ورود/دسترسی را قطع می‌کند.
func (s *Service) verifyRoleIntegrityForUser(role *Role, userID int, ip string) error {
	permIDs, err := s.repo.FindPermissionIDsByRoleID(role.ID)
	if err != nil {
		return apperror.New("DB_ERROR", "خطا در خواندن مجوزهای نقش.", err.Error(), 500)
	}
	if !VerifyRoleIntegrity(s.signer, role, permIDs) {
		_ = s.audit.LogEvent(&userID, ip, audit.EventDataTampering,
			fmt.Sprintf("دستکاری نقش %s — ورود/دسترسی کاربر %d رد شد", role.Name, userID))
		return apperror.ErrIntegrity
	}
	return nil
}

// resolvePermissions شناسه‌های مجوز را اعتبارسنجی و بارگذاری می‌کند.
func (s *Service) resolvePermissions(ids []int) ([]Permission, error) {
	unique := make(map[int]struct{}, len(ids))
	clean := make([]int, 0, len(ids))
	for _, id := range ids {
		if id <= 0 {
			continue
		}
		if _, ok := unique[id]; ok {
			continue
		}
		unique[id] = struct{}{}
		clean = append(clean, id)
	}
	if len(clean) == 0 {
		return []Permission{}, nil
	}
	perms, err := s.repo.FindPermissionsByIDs(clean)
	if err != nil {
		return nil, apperror.New("DB_ERROR", "خطا در خواندن مجوزها.", err.Error(), 500)
	}
	if len(perms) != len(clean) {
		return nil, apperror.New("VALIDATION_ERROR", "یک یا چند مجوز انتخاب‌شده معتبر نیست.", "invalid permission ids", 400)
	}
	return perms, nil
}

// toRoleDetail مدل نقش را به پاسخ جزئیات تبدیل می‌کند.
func (s *Service) toRoleDetail(role *Role) (*RoleDetailResponse, error) {
	perms, err := s.repo.FindPermissionsByRoleID(role.ID)
	if err != nil {
		return nil, apperror.New("DB_ERROR", "خطا در خواندن مجوزهای نقش.", err.Error(), 500)
	}
	permIDs := make([]int, 0, len(perms))
	permResps := make([]PermissionResponse, 0, len(perms))
	for _, p := range perms {
		permIDs = append(permIDs, p.ID)
		permResps = append(permResps, PermissionResponse{ID: p.ID, Name: p.Name, Description: p.Description})
	}
	return &RoleDetailResponse{
		ID:            role.ID,
		Name:          role.Name,
		Description:   role.Description,
		Permissions:   permResps,
		PermissionIDs: permIDs,
		IntegrityOK:   VerifyRoleIntegrity(s.signer, role, permIDs),
	}, nil
}

// UpdateSecuritySetting تنظیم امنیتی را به‌روزرسانی می‌کند.
func (s *Service) UpdateSecuritySetting(id int, value string, actorUserID int, ip string) error {
	return s.settings.UpdateSettingByID(id, value, actorUserID, ip)
}

// GetSecuritySettings تنظیمات امنیتی را برمی‌گرداند.
func (s *Service) GetSecuritySettings() ([]settings.SecuritySetting, error) {
	return s.settings.GetAllSettings(true)
}

// permissionIDsCoverCatalog بررسی می‌کند شناسه‌های مجوز کل کاتالوگ سیستم را پوشش می‌دهند.
func (s *Service) permissionIDsCoverCatalog(permIDs []int) bool {
	catalog := DefaultPermissionCatalog()
	if len(permIDs) < len(catalog) {
		return false
	}
	allPerms, err := s.repo.FindAllPermissions()
	if err != nil {
		return false
	}
	idSet := make(map[int]struct{}, len(permIDs))
	for _, id := range permIDs {
		idSet[id] = struct{}{}
	}
	names := make([]string, 0, len(permIDs))
	for _, p := range allPerms {
		if _, ok := idSet[p.ID]; ok {
			names = append(names, p.Name)
		}
	}
	return HasAllPermissions(names)
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
		UserType:    user.UserType,
	}
	if user.LastLoginAt != nil {
		t := user.LastLoginAt.Format(time.RFC3339)
		resp.LastLoginAt = &t
	}
	return resp
}
