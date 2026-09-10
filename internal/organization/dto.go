// این ماژول نمونه است؛ باقی Entity ها و Endpoint ها طبق همین الگو در فازهای بعدی تکمیل می‌شوند.
package organization

import (
	organizationpackage "github.com/tpdenta/afta-reception/internal/organizationPackage"
)

// CreateRequest بدنه ایجاد سازمان.
type CreateRequest struct {
	Name            string `json:"name" binding:"required"`
	IsTakmili       bool   `json:"is_takmili"`
	IsActive        bool   `json:"is_active"`
	IsFree          bool   `json:"is_free"`
	PackageID       uint   `json:"package_id" binding:"required"`
	CenterPackageID uint   `json:"center_package_id" binding:"required"`
}

// UpdateRequest بدنه بروزرسانی سازمان.
type UpdateRequest struct {
	Name            string `json:"name" binding:"required"`
	IsTakmili       bool   `json:"is_takmili"`
	IsActive        bool   `json:"is_active"`
	IsFree          bool   `json:"is_free"`
	PackageID       uint   `json:"package_id" binding:"required"`
	CenterPackageID uint   `json:"center_package_id" binding:"required"`
}

// Response پاسخ API سازمان.
type Response struct {
	ID                uint   `json:"id"`
	Name              string `json:"name"`
	IsTakmili         bool   `json:"is_takmili"`
	IsActive          bool   `json:"is_active"`
	IsFree            bool   `json:"is_free"`
	PackageID         uint   `json:"package_id"`
	PackageName       string `json:"package_name,omitempty"`
	CenterPackageID   uint   `json:"center_package_id"`
	CenterPackageName string `json:"center_package_name,omitempty"`

	// Package و CenterPackage فقط برای مصرف داخلی لایه سرویس؛ در JSON ارسال نمی‌شوند.
	Package       organizationpackage.OrganizationPackage `json:"-"`
	CenterPackage organizationpackage.OrganizationPackage `json:"-"`
}
