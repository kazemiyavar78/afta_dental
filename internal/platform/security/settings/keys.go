package settings

// نام کلیدهای تنظیمات امنیتی — از رشته دستی در کد استفاده نشود.
const (
	MaximumNumberOfSessionsPerUser = "MaximumNumberOfSessionsPerUser"
	MaximumNumberOfFailedLogin     = "MaximumNumberOfFailedLogin"
	MaximumTimeOfUserIpBeingLocked = "MaximumTimeOfUserIpBeingLocked"
	MaximumSizeOfLogTables         = "MaximumSizeOfLogTables"
	LastVerifiedEventID            = "LastVerifiedEventId"
	WorkHoursStart                 = "WorkHours_Start"
	WorkHoursEnd                   = "WorkHours_End"
	IsSaturdayBlocked              = "IsSaturdayBlocked"
	IsSundayBlocked                = "IsSundayBlocked"
	IsMondayBlocked                = "IsMondayBlocked"
	IsTuesdayBlocked               = "IsTuesdayBlocked"
	IsWednesdayBlocked             = "IsWednesdayBlocked"
	IsThursdayBlocked              = "IsThursdayBlocked"
	IsFridayBlocked                = "IsFridayBlocked"
	PasswordMinLength              = "Password_MinLength"
	PasswordRequireUppercase       = "Password_RequireUppercase"
	PasswordRequireLowercase       = "Password_RequireLowercase"
	PasswordRequireDigit           = "Password_RequireDigit"
	PasswordRequireSpecial         = "Password_RequireSpecial"
	MaximumUploadsPerWindow        = "MaximumUploadsPerWindow"
	UploadWindowMinutes            = "UploadWindowMinutes"
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

//دیکشنری فارسی

var SettingsDicFN = map[string]string{
	"MaximumNumberOfSessionsPerUser": "حداکثر تعداد نشست همزمان",
	"MaximumNumberOfFailedLogin":     "بیشترین تعداد ورود اشتباه",
	"MaximumTimeOfUserIpBeingLocked": "بیشترین مدت زمان قفل بودن اکانت",
	"MaximumSizeOfLogTables":         "حداکثر تعداد لاگ",
	"LastVerifiedEventId":            "آخرین رویداد ثبت شده",
	"WorkHours_Start":                "ساعت شروع کاری",
	"WorkHours_End":                  "ساعت پایان کاری",
	"IsSaturdayBlocked":              "تعطیلی شنبه",
	"IsSundayBlocked":                "تعطیلی یکشنبه",
	"IsMondayBlocked":                "تعطیلی دوشنبه",
	"IsTuesdayBlocked":               "تعطیلی سه شنبه",
	"IsWednesdayBlocked":             "تعطیلی چهارشنبه",
	"IsThursdayBlocked":              "تعطیلی پنجشنبه",
	"IsFridayBlocked":                "تعطیلی جمعه",
	"Password_MinLength":             "کمترین مقدار رمزعبود",
	"Password_RequireUppercase":      "حروف بزرگ استفاده شود ",
	"Password_RequireLowercase":      "حروف کوچک مجاز است",
	"Password_RequireDigit":          "عدد مجاز است",
	"Password_RequireSpecial":        "حروف علامت دار مجاز است",
	"MaximumUploadsPerWindow":        "بیشترین تعداد آپلود",
	"UploadWindowMinutes":            "بیشترین زمان آپلود",
}
