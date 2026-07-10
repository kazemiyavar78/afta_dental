package user

// مجوزهای پیش‌فرض هر نقش — در فاز بعدی از جدول RolePermissions خوانده می‌شود.
var defaultRolePermissions = map[string][]string{
	"Admin": {
		"users.read", "users.create", "users.update",
		"reception.read", "reception.create",
		"organization.read", "organization.create",
		"fund.read", "fund.create",
		"tariff.read", "tariff.create","security.settings", 
	},
	"Reception": {
		"reception.read", "reception.create",
		"organization.read",
		"fund.read",
		"tariff.read",
	},
	"Doctor": {
		"reception.read",
		"tariff.read",
	},
}

// GetPermissionsForRole لیست نام مجوزهای یک نقش را برمی‌گرداند.
func GetPermissionsForRole(roleName string) []string {
	if perms, ok := defaultRolePermissions[roleName]; ok {
		return perms
	}
	return []string{}
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
