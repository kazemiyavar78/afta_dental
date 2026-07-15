package tariff

import (
	"github.com/tpdenta/afta-reception/internal/organization"
	"github.com/tpdenta/afta-reception/internal/services"
)

// CalculateTariffForOrganizationRequest بدنه تست/ذخیره تعرفه برای یک سازمان.
type CalculateTariffForOrganizationRequest struct {
	OrganizationID uint `json:"organization_id" binding:"required"`
	// به جز خدمات (شناسه خدمات حذف‌شده از محاسبه)
	ExcludeServiceIDs []uint `json:"exclude_service_ids"`
	// مبلغ فنی مرکز
	TechnicalAmount int64 `json:"technical_amount"`
	// مبلغ حرفه‌ای مرکز
	ProfessionalCenterAmount int64 `json:"professional_center_amount"`
	// مبلغ مصرفی مرکز
	ConsumptionCenterAmount int64 `json:"consumption_center_amount"`
}

// RecalculateTariffRequest بدنه بازمحاسبه یک تعرفه ذخیره‌شده با سه مبلغ مرکز.
type RecalculateTariffRequest struct {
	TechnicalAmount          int64 `json:"technical_amount"`
	ProfessionalCenterAmount int64 `json:"professional_center_amount"`
	ConsumptionCenterAmount  int64 `json:"consumption_center_amount"`
}

// CalculateResponse پاسخ تست محاسبه تعرفه.
type CalculateResponse struct {
	Services []ServiceWithPrice `json:"services"`
	// سازمان
	Organization organization.Response `json:"organization"`
	// اطلاعات ارسالی
	SendInfo CalculateTariffForOrganizationRequest `json:"send_info"`
}

// ServiceWithPrice خدمت به‌همراه نتیجه محاسبه.
type ServiceWithPrice struct {
	Service   services.Response `json:"service"`
	Calculate Calculate         `json:"calculate"`
}

// Calculate مبالغ محاسبه‌شده برای یک خدمت.
type Calculate struct {
	// مبلغ کل (نرخ)
	TotalAmount int64 `json:"total_amount"`
	// تعرفه
	Tariff int64 `json:"tariff"`
	// سهم سازمان
	OrganizationShare int64 `json:"organization_share"`
	// تکمیلی
	SupplementAmount int64 `json:"supplement_amount"`
	// یارانه
	SubsidyAmount int64 `json:"subsidy_amount"`
	// صندوق = total - organization_share - supplement - subsidy
	FundAmount int64 `json:"fund_amount"`
}

// TariffResponse پاسخ یک ردیف تعرفه ذخیره‌شده به‌همراه نام خدمت.
type TariffResponse struct {
	ID                 uint   `json:"id"`
	OrganizationID     uint   `json:"organization_id"`
	ServiceID          uint   `json:"service_id"`
	ServiceCode        string `json:"service_code"`
	ServiceName        string `json:"service_name"`
	Amount             int64  `json:"amount"`
	TariffAmount       int64  `json:"tariff_amount"`
	OrganizationShare  int64  `json:"organization_share"`
	SupplementaryShare int64  `json:"supplementary_share"`
	SubsidyShare       int64  `json:"subsidy_share"`
	FundAmount         int64  `json:"fund_amount"`
}

// SaveTariffResponse پاسخ ذخیره گروهی تعرفه‌ها.
type SaveTariffResponse struct {
	Items []TariffResponse `json:"items"`
}
