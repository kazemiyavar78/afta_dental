// این ماژول نمونه است؛ باقی Entity ها و Endpoint ها طبق همین الگو در فازهای بعدی تکمیل می‌شوند.
package organization

import (
	organizationpackage "github.com/tpdenta/afta-reception/internal/organizationPackage"
)

// CreateRequest بدنه ایجاد سازمان.
type CreateRequest struct {
	Name      string `json:"name" binding:"required"`
	IsTakmili bool   `json:"is_takmili"`
	IsActive  bool   `json:"is_active"`
	PackageID uint   `json:"package_id" binding:"required"`
}

// UpdateRequest بدنه بروزرسانی سازمان.
type UpdateRequest struct {
	Name      string `json:"name" binding:"required"`
	IsTakmili bool   `json:"is_takmili"`
	IsActive  bool   `json:"is_active"`
	PackageID uint   `json:"package_id" binding:"required"`
}

// Response پاسخ API سازمان.
type Response struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	IsTakmili   bool   `json:"is_takmili"`
	IsActive    bool   `json:"is_active"`
	PackageID   uint   `json:"package_id"`
	PackageName string `json:"package_name,omitempty"`

	// Package فقط برای مصرف داخلی لایه سرویس (مثلاً محاسبه تعرفه)؛ در JSON ارسال نمی‌شود.
	Package organizationpackage.OrganizationPackage `json:"-"`
}
