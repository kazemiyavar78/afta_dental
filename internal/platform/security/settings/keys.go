package settings

// نام کلیدهای تنظیمات امنیتی — از رشته دستی در کد استفاده نشود.
const (
	MaximumNumberOfSessionsPerUser    = "MaximumNumberOfSessionsPerUser"
	MaximumNumberOfFailedLogin        = "MaximumNumberOfFailedLogin"
	MaximumTimeOfUserIpBeingLocked    = "MaximumTimeOfUserIpBeingLocked"
	MaximumSizeOfLogTables            = "MaximumSizeOfLogTables"
	LastVerifiedEventID               = "LastVerifiedEventId"
	WorkHoursStart                    = "WorkHours_Start"
	WorkHoursEnd                      = "WorkHours_End"
	IsSaturdayBlocked                 = "IsSaturdayBlocked"
	IsSundayBlocked                   = "IsSundayBlocked"
	IsMondayBlocked                   = "IsMondayBlocked"
	IsTuesdayBlocked                  = "IsTuesdayBlocked"
	IsWednesdayBlocked                = "IsWednesdayBlocked"
	IsThursdayBlocked                 = "IsThursdayBlocked"
	IsFridayBlocked                   = "IsFridayBlocked"
	PasswordMinLength                 = "Password_MinLength"
	PasswordRequireUppercase          = "Password_RequireUppercase"
	PasswordRequireLowercase          = "Password_RequireLowercase"
	PasswordRequireDigit              = "Password_RequireDigit"
	PasswordRequireSpecial            = "Password_RequireSpecial"
	MaximumUploadsPerWindow             = "MaximumUploadsPerWindow"
	UploadWindowMinutes               = "UploadWindowMinutes"
)

// DefaultSettings مقادیر پیش‌فرض تنظیمات امنیتی.
var DefaultSettings = map[string]string{
	MaximumNumberOfSessionsPerUser: "3",
	MaximumNumberOfFailedLogin:     "5",
	MaximumTimeOfUserIpBeingLocked: "30",
	MaximumSizeOfLogTables:         "100",
	LastVerifiedEventID:            "0",
	WorkHoursStart:                 "08:00",
	WorkHoursEnd:                   "18:00",
	IsSaturdayBlocked:              "false",
	IsSundayBlocked:                "true",
	IsMondayBlocked:                "false",
	IsTuesdayBlocked:               "false",
	IsWednesdayBlocked:             "false",
	IsThursdayBlocked:              "false",
	IsFridayBlocked:                "true",
	PasswordMinLength:              "8",
	PasswordRequireUppercase:       "true",
	PasswordRequireLowercase:       "true",
	PasswordRequireDigit:           "true",
	PasswordRequireSpecial:         "true",
	MaximumUploadsPerWindow:        "10",
	UploadWindowMinutes:            "60",
}
