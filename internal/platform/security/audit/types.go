package audit

// انواع رویدادهای امنیتی
const (
	EventLoginSuccess          = "LOGIN_SUCCESS"
	EventLoginFailed           = "LOGIN_FAILED"
	EventLogout                = "LOGOUT"
	EventSecuritySettingChange = "SECURITY_SETTING_CHANGE"
	EventUpload                = "UPLOAD"
	EventDownload              = "DOWNLOAD"
	EventUserDataChange        = "USER_DATA_CHANGE"
	EventRoleChange            = "ROLE_CHANGE"
	EventLogCleanup            = "LOG_CLEANUP"
	EventDataTampering         = "DATA_TAMPERING"
	EventCSRFViolation         = "CSRF_VIOLATION"
	EventSessionRevoked        = "SESSION_REVOKED"
)
