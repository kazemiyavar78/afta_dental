package audit

// سطح اهمیت رویداد برای نمایش و اولویت بررسی.
const (
	SeverityCritical = "critical"
	SeverityHigh     = "high"
	SeverityNormal   = "normal"
)

// نتیجه ثبت‌شده رویداد. مقدار خالی فقط برای ردیف‌های قدیمی مجاز است.
const (
	ResultSuccess = "success"
	ResultFailure = "failure"
)

// نسخه هش ردیف. نسخه ۱ فیلدهای قدیمی را هش می‌کند تا ردیف‌های موجود هنوز تأیید شوند.
// نسخه ۲ فیلدهای فاعل، هدف، نتیجه، دلیل و جزئیات را هم وارد payload می‌کند.
const (
	HashVersionV1 uint8 = 1
	HashVersionV2 uint8 = 2
)

// نوع هدف رویداد. مقدار در ستون TargetType ذخیره می‌شود.
const (
	TargetUser                = "user"
	TargetRole                = "role"
	TargetSession             = "session"
	TargetSetting             = "setting"
	TargetPatient             = "patient"
	TargetOrganization        = "organization"
	TargetOrganizationPackage = "organization_package"
	TargetService             = "service"
	TargetExcludedService     = "excluded_service"
	TargetTariff              = "tariff"
	TargetRegulation          = "regulation"
	TargetSpecialCode         = "special_code"
	TargetFund                = "fund"
	TargetReception           = "reception"
	TargetWallet              = "wallet"
	TargetBankAccount         = "bank_account"
	TargetCreditAccount       = "credit_account"
)

// انواع جاری. ثبت جدید فقط از همین کدهاست.
const (
	EventSessionCreated    = "SESSION_CREATED"
	EventSessionTerminated = "SESSION_TERMINATED"
	EventLogout            = "LOGOUT"
	EventLoginFailed       = "LOGIN_FAILED"
	EventLoginBlocked      = "LOGIN_BLOCKED"
	EventIPLocked          = "IP_LOCKED"
	EventAccountLocked     = "ACCOUNT_LOCKED"
	EventUserCreated       = "USER_CREATED"
	EventUserUpdated       = "USER_UPDATED"
	EventPasswordChanged   = "PASSWORD_CHANGED"
	EventRoleCreated       = "ROLE_CREATED"
	EventRoleUpdated       = "ROLE_UPDATED"
	EventRoleDeleted       = "ROLE_DELETED"
	EventSettingChanged    = "SETTING_CHANGED"
	EventSettingRemoved    = "SETTING_REMOVED"
	EventAuditStarted      = "AUDIT_STARTED"
	EventAuditStopped      = "AUDIT_STOPPED"
	EventAuditRead         = "AUDIT_READ"
	EventAuditReadDenied   = "AUDIT_READ_DENIED"
	EventAuditChainBroken  = "AUDIT_CHAIN_BROKEN"
	EventEntityCreated     = "ENTITY_CREATED"
	EventEntityUpdated     = "ENTITY_UPDATED"
	EventEntityDeleted     = "ENTITY_DELETED"
	EventEntityRestored    = "ENTITY_RESTORED"
	EventUpload            = "UPLOAD"
)

// انواع منسوخ. برای خواندن ردیف‌های قدیمی می‌مانند و دیگر ثبت نمی‌شوند.
const (
	// Deprecated: ردیف قدیمی ورود موفق. ثبت جدید SESSION_CREATED است.
	EventLoginSuccess = "LOGIN_SUCCESS"
	// Deprecated: ردیف قدیمی تغییر تنظیم. ثبت جدید SETTING_CHANGED است.
	EventSecuritySettingChange = "SECURITY_SETTING_CHANGE"
	// Deprecated: ردیف قدیمی تغییر داده. ثبت جدید USER_* یا ENTITY_* است.
	EventUserDataChange = "USER_DATA_CHANGE"
	// Deprecated: ردیف قدیمی پاک‌سازی لاگ. ثبت جدید در گام نگهداری لاگ جدا می‌شود.
	EventLogCleanup = "LOG_CLEANUP"
	// Deprecated: ردیف قدیمی شکست زنجیره. ثبت جدید AUDIT_CHAIN_BROKEN است.
	EventDataTampering = "DATA_TAMPERING"
	// Deprecated: ردیف قدیمی نقض CSRF. تا مهاجرت فراخوانی در کد می‌ماند.
	EventCSRFViolation = "CSRF_VIOLATION"
	// Deprecated: ردیف قدیمی حذف نشست. ثبت جدید SESSION_TERMINATED با Reason است.
	EventSessionRevoked = "SESSION_REVOKED"
)

// کد دلیل خروج، رد ورود یا شکست یکپارچگی. مقدار در ستون Reason ذخیره می‌شود.
const (
	ReasonUserLogout         = "user_logout"
	ReasonAdminTerminated    = "admin_terminated"
	ReasonMaxSessions        = "max_sessions"
	ReasonSecurityCodeFailed = "security_code_failed"
	ReasonIdleTimeout        = "idle_timeout"
	ReasonSettingsChanged    = "settings_changed"
	ReasonRoleChanged        = "role_changed"
	ReasonIPChanged          = "ip_changed"
	ReasonWrongPassword      = "wrong_password"
	ReasonUserNotFound       = "user_not_found"
	ReasonAccountLocked      = "account_locked"
	ReasonAccountInactive    = "account_inactive"
	ReasonIPLocked           = "ip_locked"
	ReasonOutsideWorkHours   = "outside_work_hours"
	ReasonIntegrityViolation = "integrity_violation"
)

// EventSpec برچسب فارسی و سطح اهمیت یک نوع رویداد است.
// Deprecated یعنی کد فقط برای ردیف‌های قبلی خوانده می‌شود.
type EventSpec struct {
	Label      string
	Severity   string
	Deprecated bool
}

// eventCatalog منبع واحد برچسب و اهمیت است تا UI و توضیح خودکار از یک جا بخوانند.
var eventCatalog = map[string]EventSpec{
	EventSessionCreated:        {Label: "ایجاد نشست", Severity: SeverityNormal},
	EventSessionTerminated:     {Label: "پایان نشست", Severity: SeverityHigh},
	EventLogout:                {Label: "خروج", Severity: SeverityNormal},
	EventLoginFailed:           {Label: "ورود ناموفق", Severity: SeverityHigh},
	EventLoginBlocked:          {Label: "رد ورود", Severity: SeverityHigh},
	EventIPLocked:              {Label: "قفل نشانی IP", Severity: SeverityCritical},
	EventAccountLocked:         {Label: "قفل حساب", Severity: SeverityCritical},
	EventUserCreated:           {Label: "ایجاد کاربر", Severity: SeverityHigh},
	EventUserUpdated:           {Label: "به‌روزرسانی کاربر", Severity: SeverityNormal},
	EventPasswordChanged:       {Label: "تغییر رمز", Severity: SeverityHigh},
	EventRoleCreated:           {Label: "ایجاد نقش", Severity: SeverityNormal},
	EventRoleUpdated:           {Label: "به‌روزرسانی نقش", Severity: SeverityNormal},
	EventRoleDeleted:           {Label: "حذف نقش", Severity: SeverityHigh},
	EventSettingChanged:        {Label: "تغییر تنظیم امنیتی", Severity: SeverityHigh},
	EventSettingRemoved:        {Label: "حذف تنظیم امنیتی", Severity: SeverityHigh},
	EventAuditStarted:          {Label: "آغاز ممیزی", Severity: SeverityHigh},
	EventAuditStopped:          {Label: "اتمام ممیزی", Severity: SeverityHigh},
	EventAuditRead:             {Label: "خواندن ممیزی", Severity: SeverityNormal},
	EventAuditReadDenied:       {Label: "رد خواندن ممیزی", Severity: SeverityHigh},
	EventAuditChainBroken:      {Label: "شکست زنجیره ممیزی", Severity: SeverityCritical},
	EventEntityCreated:         {Label: "ایجاد رکورد", Severity: SeverityNormal},
	EventEntityUpdated:         {Label: "به‌روزرسانی رکورد", Severity: SeverityNormal},
	EventEntityDeleted:         {Label: "حذف رکورد", Severity: SeverityNormal},
	EventEntityRestored:        {Label: "بازیابی رکورد", Severity: SeverityNormal},
	EventUpload:                {Label: "بارگذاری فایل", Severity: SeverityNormal},
	EventLoginSuccess:          {Label: "ورود موفق", Severity: SeverityNormal, Deprecated: true},
	EventSecuritySettingChange: {Label: "تغییر تنظیم امنیتی", Severity: SeverityHigh, Deprecated: true},
	EventUserDataChange:        {Label: "تغییر داده", Severity: SeverityNormal, Deprecated: true},
	EventLogCleanup:            {Label: "پاک‌سازی لاگ", Severity: SeverityNormal, Deprecated: true},
	EventDataTampering:         {Label: "دستکاری داده", Severity: SeverityCritical, Deprecated: true},
	EventCSRFViolation:         {Label: "نقض CSRF", Severity: SeverityHigh, Deprecated: true},
	EventSessionRevoked:        {Label: "حذف نشست", Severity: SeverityHigh, Deprecated: true},
}

// reasonLabels برچسب فارسی کدهای Reason است.
var reasonLabels = map[string]string{
	ReasonUserLogout:         "خروج توسط کاربر",
	ReasonAdminTerminated:    "پایان نشست توسط مدیر",
	ReasonMaxSessions:        "رسیدن به سقف نشست",
	ReasonSecurityCodeFailed: "شکست بررسی کد امنیتی",
	ReasonIdleTimeout:        "پایان به‌خاطر عدم فعالیت",
	ReasonSettingsChanged:    "تغییر تنظیمات امنیتی",
	ReasonRoleChanged:        "تغییر نقش",
	ReasonIPChanged:          "تغییر نشانی IP",
	ReasonWrongPassword:      "رمز نادرست",
	ReasonUserNotFound:       "کاربر پیدا نشد",
	ReasonAccountLocked:      "حساب قفل است",
	ReasonAccountInactive:    "حساب غیرفعال است",
	ReasonIPLocked:           "نشانی IP قفل است",
	ReasonOutsideWorkHours:   "خارج از ساعت کاری",
	ReasonIntegrityViolation: "نقض یکپارچگی داده",
}

// targetLabels برچسب فارسی TargetTypeهای شناخته‌شده است.
var targetLabels = map[string]string{
	TargetUser:                "کاربر",
	TargetRole:                "نقش",
	TargetSession:             "نشست",
	TargetSetting:             "تنظیم",
	TargetPatient:             "بیمار",
	TargetOrganization:        "سازمان",
	TargetOrganizationPackage: "بسته تعرفه",
	TargetService:             "خدمت",
	TargetExcludedService:     "خدمت خارج‌شده",
	TargetTariff:              "تعرفه",
	TargetRegulation:          "ضابطه",
	TargetSpecialCode:         "کد خاص",
	TargetFund:                "صندوق",
	TargetReception:           "پذیرش",
	TargetWallet:              "کیف پول",
	TargetBankAccount:         "حساب بانکی",
	TargetCreditAccount:       "حساب اعتباری",
}

// LookupEvent برچسب و اهمیت یک نوع را برمی‌گرداند.
// eventType همان EventType ذخیره‌شده است. ok نادرست یعنی نوع در کاتالوگ نیست.
// انواع منسوخ هم برمی‌گردند تا ردیف قدیمی بدون برچسب نماند.
func LookupEvent(eventType string) (EventSpec, bool) {
	spec, ok := eventCatalog[eventType]
	return spec, ok
}

// EventLabel برچسب فارسی نوع را برمی‌گرداند.
// اگر نوع ناشناخته باشد همان کد را برمی‌گرداند تا صفحه لاگ خالی نماند.
func EventLabel(eventType string) string {
	spec, ok := LookupEvent(eventType)
	if !ok || spec.Label == "" {
		return eventType
	}
	return spec.Label
}

// ReasonLabel برچسب فارسی کد دلیل را برمی‌گرداند.
// code مقدار ستون Reason است. کد ناشناخته همان رشته ورودی است.
func ReasonLabel(code string) string {
	if code == "" {
		return ""
	}
	if label, ok := reasonLabels[code]; ok {
		return label
	}
	return code
}
