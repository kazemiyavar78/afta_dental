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
	"wallet.read":                  "خواندن کیف پول",
	"wallet.cash":                  "افزایش و کاهش اعتبار نقدی",
	"wallet.card_to_card":          "افزایش و کاهش کارت به کارت",
	"wallet.card_reader":           "شارژ از کارتخوان",
	"bank_account.read":            "خواندن حسابهای بانکی",
	"bank_account.create":          "ایجاد حساب بانکی",
	"bank_account.update":          "بروزرسانی حساب بانکی",
	"bank_account.delete":          "حذف حساب بانکی",
	"tariff.read":                  "خواندن تعرفهها",
	"tariff.create":                "ایجاد تعرفهها",
	"tariff.update":                "بروزرسانی تعرفهها",
	"tariff.delete":                "حذف تعرفهها",
	"special_code.read":            "خواندن کد خاص",
	"special_code.create":          "ایجاد کد خاص",
	"special_code.update":          "بروزرسانی کد خاص",
	"special_code.delete":          "حذف کد خاص",
	"regulation.read":              "خواندن ضوابط",
	"regulation.create":            "ایجاد ضوابط",
	"regulation.update":            "بروزرسانی ضوابط",
	"regulation.delete":            "حذف ضوابط",
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

// HasAllPermissions وقتی کاربر تمام مجوزهای کاتالوگ سیستم را دارد true برمی‌گرداند (تعریف ادمین).
func HasAllPermissions(permissions []string) bool {
	catalog := DefaultPermissionCatalog()
	if len(permissions) < len(catalog) {
		return false
	}
	owned := make(map[string]struct{}, len(permissions))
	for _, p := range permissions {
		owned[p] = struct{}{}
	}
	for name := range catalog {
		if _, ok := owned[name]; !ok {
			return false
		}
	}
	return true
}

// PermissionModule کلید ماژول مجوز را از نام کامل استخراج می‌کند (مثلاً bank_account از bank_account.read).
func PermissionModule(permissionName string) string {
	for i := 0; i < len(permissionName); i++ {
		if permissionName[i] == '.' {
			return permissionName[:i]
		}
	}
	return permissionName
}
