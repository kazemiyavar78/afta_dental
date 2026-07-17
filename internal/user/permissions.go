package user

// مجوزهای پیش‌فرض سیستم — در لحظه اجرا در جدول Permissions همگام‌سازی می‌شوند.
var defaultPermissions = map[string]string{
	"users.read":                   "خواندن کاربران",
	"users.create":                 "ایجاد کاربران",
	"users.update":                 "بروزرسانی کاربران",
	"users.listSessions":           "لیست جلسات کاربران",
	"roles.read":                   "خواندن نقش‌ها",
	"roles.create":                 "ایجاد نقش‌ها",
	"roles.update":                 "بروزرسانی نقش‌ها",
	"roles.delete":                 "حذف نقش‌ها",
	"reception.read":               "خواندن رسیدها",
	"reception.create":             "ایجاد رسیدها",
	"reception.update":             "ویرایش رسیدها",
	"reception.delete":             "حذف رسیدها",
	"reception.restore":            "بازیابی رسیدها",
	"organization.read":            "خواندن سازمانها",
	"organization.create":          "ایجاد سازمانها",
	"organization.update":          "بروزرسانی سازمانها",
	"organization.delete":          "حذف سازمانها",
	"patient.read":                 "خواندن بیماران",
	"patient.create":               "ایجاد بیماران",
	"patient.update":               "بروزرسانی بیماران",
	"patient.delete":               "حذف بیماران",
	"organization_packages.read":   "خواندن بستههای سازمانی",
	"organization_packages.create": "ایجاد بستههای سازمانی",
	"organization_packages.update": "بروزرسانی بستههای سازمانی",
	"organization_packages.delete": "حذف بستههای سازمانی",
	"services.read":                "خواندن خدمات",
	"services.create":              "ایجاد خدمات",
	"services.update":              "بروزرسانی خدمات",
	"services.delete":              "حذف خدمات",
	"fund.read":                    "خواندن صندوق",
	"fund.create":                  "ایجاد صندوق",
	"fund.update":                  "بروزرسانی صندوق",
	"fund.delete":                  "حذف صندوق",
	"tariff.read":                  "خواندن تعرفهها",
	"tariff.create":                "ایجاد تعرفهها",
	"tariff.update":                "بروزرسانی تعرفهها",
	"tariff.delete":                "حذف تعرفهها",
	"security.settings":            "خواندن تنظیمات امنیتی",
	"logs.read":                    "خواندن لاگها",
}

// DefaultPermissionCatalog کپی از کاتالوگ ثابت مجوزهای سیستم را برمی‌گرداند.
func DefaultPermissionCatalog() map[string]string {
	out := make(map[string]string, len(defaultPermissions))
	for k, v := range defaultPermissions {
		out[k] = v
	}
	return out
}

// HasPermission بررسی وجود مجوز در لیست.
func HasPermission(permissions []string, permission string) bool {
	for _, p := range permissions {
		if p == permission {
			return true
		}
	}
	return false
}
