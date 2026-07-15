package organizationpackage

// CreateRequest بدنه ایجاد بسته تعرفه.
type CreateRequest struct {
	PackageName                      string `json:"package_name" binding:"required"`
	PackageDescription               string `json:"package_description"`
	TechnicalCoefficient             int    `json:"technical_coefficient"`
	TechnicalProfessionalCoefficient int    `json:"technical_professional_coefficient"`
	ConsumptionCoefficient           int    `json:"consumption_coefficient"`
	SubsidyPercentage                int    `json:"subsidy_percentage"`
	SupplementaryPercentage          int    `json:"supplementary_percentage"`
	OrganizationPercentage           int    `json:"organization_percentage"`
}

// UpdateRequest بدنه بروزرسانی بسته تعرفه.
type UpdateRequest struct {
	PackageName                      string `json:"package_name" binding:"required"`
	PackageDescription               string `json:"package_description"`
	TechnicalCoefficient             int    `json:"technical_coefficient"`
	TechnicalProfessionalCoefficient int    `json:"technical_professional_coefficient"`
	ConsumptionCoefficient           int    `json:"consumption_coefficient"`
	SubsidyPercentage                int    `json:"subsidy_percentage"`
	SupplementaryPercentage          int    `json:"supplementary_percentage"`
	OrganizationPercentage           int    `json:"organization_percentage"`
}

// Response پاسخ API بسته تعرفه.
type Response struct {
	ID                               uint   `json:"id"`
	PackageName                      string `json:"package_name"`
	PackageDescription               string `json:"package_description"`
	TechnicalCoefficient             int    `json:"technical_coefficient"`
	TechnicalProfessionalCoefficient int    `json:"technical_professional_coefficient"`
	ConsumptionCoefficient           int    `json:"consumption_coefficient"`
	SubsidyPercentage                int    `json:"subsidy_percentage"`
	SupplementaryPercentage          int    `json:"supplementary_percentage"`
	OrganizationPercentage           int    `json:"organization_percentage"`
}
