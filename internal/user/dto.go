package user

// CreateUserRequest درخواست ایجاد کاربر.
type CreateUserRequest struct {
	Username    string  `json:"username" binding:"required"`
	Password    string  `json:"password" binding:"required"`
	Address     string  `json:"address"`
	Name        string  `json:"name" binding:"required"`
	Family      string  `json:"family" binding:"required"`
	PhoneNumber string  `json:"phone_number"`
	MedicalCode *string `json:"medical_code"`
	RoleID      int     `json:"role_id" binding:"required"`
}

// UpdateUserRequest درخواست به‌روزرسانی کاربر.
type UpdateUserRequest struct {
	Address     *string `json:"address"`
	Name        *string `json:"name"`
	Family      *string `json:"family"`
	PhoneNumber *string `json:"phone_number"`
	MedicalCode *string `json:"medical_code"`
	RoleID      *int    `json:"role_id"`
	IsActive    *bool   `json:"is_active"`
}

// LoginRequest درخواست ورود.
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// ChangePasswordRequest درخواست تغییر رمز عبور.
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
}

// UserResponse پاسخ اطلاعات کاربر (بدون فیلدهای حساس).
type UserResponse struct {
	ID           int     `json:"id"`
	Username     string  `json:"username"`
	Address      string  `json:"address"`
	Name         string  `json:"name"`
	Family       string  `json:"family"`
	PhoneNumber  string  `json:"phone_number"`
	MedicalCode  *string `json:"medical_code"`
	RoleID       int     `json:"role_id"`
	RoleName     string  `json:"role_name"`
	IsActive     bool    `json:"is_active"`
	IsLocked     bool    `json:"is_locked"`
	LastLoginAt  *string `json:"last_login_at"`
}

// MeResponse پاسخ اطلاعات کاربر لاگین‌شده همراه مجوزها.
type MeResponse struct {
	User        UserResponse `json:"user"`
	Permissions []string     `json:"permissions"`
}

// RoleResponse پاسخ اطلاعات نقش.
type RoleResponse struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// LoginResponse پاسخ ورود موفق.
type LoginResponse struct {
	User      UserResponse `json:"user"`
	SessionID string       `json:"session_id"`
}

// UpdateSecuritySettingRequest درخواست تغییر تنظیم امنیتی.
type UpdateSecuritySettingRequest struct {
	Name  string `json:"name" binding:"required"`
	Value string `json:"value" binding:"required"`
}

// SecuritySettingResponse پاسخ تنظیم امنیتی.
type SecuritySettingResponse struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// SessionResponse پاسخ اطلاعات نشست.
type SessionResponse struct {
	ID           string `json:"id"`
	Ip           string `json:"ip"`
	Browser      string `json:"browser"`
	CreationTime string `json:"creation_time"`
	UserID       int    `json:"user_id"`
}
