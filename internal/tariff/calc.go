package tariff

import (
	"github.com/tpdenta/afta-reception/internal/organization"
	"github.com/tpdenta/afta-reception/internal/platform/apperror"
	"github.com/tpdenta/afta-reception/internal/services"
	"github.com/tpdenta/afta-reception/internal/user"
)

// hasServiceCoefficients مشخص می‌کند خدمت بر اساس ضرایب محاسبه می‌شود یا نرخ ثابت.
func hasServiceCoefficients(svc services.ServiceItem) bool {
	return svc.TechnicalCoefficient > 0 ||
		svc.ProfessionalCoefficient > 0 
}

// applySpecialistProfessionalBoost برای پزشک متخصص ۱۵٪ به ضریب حرفه‌ای اضافه می‌کند.
func applySpecialistProfessionalBoost(svc services.ServiceItem, userType user.UserType) services.ServiceItem {
	if userType == user.UserTypeSpecialist && hasServiceCoefficients(svc) {
		svc.ProfessionalCoefficient *= 1.15
	}
	return svc
}

// calcAmountFromCenterPackage نرخ خدمت را از ضرایب خدمت و بسته مرکز محاسبه می‌کند.
func calcAmountFromCenterPackage(svc services.ServiceItem, center packageAmounts) int64 {
	amount := int64(svc.TechnicalCoefficient * float64(center.Technical))
	amount += int64(svc.ProfessionalCoefficient * float64(center.Professional))
	amount += int64(svc.ConsumptionCoefficient * float64(center.Consumption))
	return amount
}

// calcTariffFromPackage تعرفه را از ضرایب خدمت و بسته تعرفه سازمان محاسبه می‌کند.
func calcTariffFromPackage(svc services.ServiceItem, pkg packageAmounts) int64 {
	tariffAmount := int64(svc.TechnicalCoefficient * float64(pkg.Technical))
	tariffAmount += int64(svc.ProfessionalCoefficient * float64(pkg.Professional))
	tariffAmount += int64(svc.ConsumptionCoefficient * float64(pkg.Consumption))
	return tariffAmount
}

// packageAmounts مقادیر لازم از یک بسته برای محاسبه نرخ/تعرفه.
type packageAmounts struct {
	Technical    int
	Professional int
	Consumption  int
}

// calcOrganizationShare سهم سازمان را از تعرفه و درصد بسته به‌دست می‌آورد.
func calcOrganizationShare(tariffAmount int64, organizationPercentage int) int64 {
	return int64(float64(tariffAmount) * float64(organizationPercentage) / 100)
}

// calcSubsidyShare یارانه را از (نرخ − سهم سازمان) و درصد یارانه بسته محاسبه می‌کند.
func calcSubsidyShare(amount, organizationShare int64, subsidyPercentage int) int64 {
	return int64(float64(amount-organizationShare) * float64(subsidyPercentage) / 100)
}

// calcSupplementaryShare سهم تکمیلی را از نرخ و درصد تکمیلی بسته محاسبه می‌کند.
func calcSupplementaryShare(amount int64, supplementaryPercentage int) int64 {
	return int64(float64(amount) * float64(supplementaryPercentage) / 100)
}

// hasCenterPackage وجود بسته مرکز معتبر برای سازمان را بررسی می‌کند.
func hasCenterPackage(org *organization.Response) bool {
	return org != nil && org.CenterPackageID > 0 && org.CenterPackage.ID > 0
}

// calculateUncoveredAmount فقط نرخ را برای خدمت خارج‌ازپوشش محاسبه می‌کند؛ تعرفه و سهم‌ها صفر می‌مانند.
func calculateUncoveredAmount(svc services.ServiceItem, org *organization.Response) (int64, error) {
	if hasServiceCoefficients(svc) {
		if !hasCenterPackage(org) {
			return 0, apperror.New(
				"E-015",
				"سازمان برای خدمت خارج‌ازپوشش باید بسته مرکز داشته باشد.",
				"center package required for uncovered service with coefficients",
				400,
			)
		}
		center := packageAmounts{
			Technical:    org.CenterPackage.TechnicalCoefficient,
			Professional: org.CenterPackage.TechnicalProfessionalCoefficient,
			Consumption:  org.CenterPackage.ConsumptionCoefficient,
		}
		return calcAmountFromCenterPackage(svc, center), nil
	}

	if svc.ServiceRate <= 0 {
		return 0, apperror.New(
			"E-016",
			"خدمت خارج‌ازپوشش باید نرخ معتبر داشته باشد.",
			"service rate required for uncovered service without coefficients",
			400,
		)
	}
	return int64(svc.ServiceRate), nil
}

// CalculateServicePrice نرخ، تعرفه و سهم‌ها را برای یک خدمت در لحظه محاسبه می‌کند.
// اگر uncovered باشد فقط نرخ (با بسته مرکز یا نرخ ثابت) محاسبه می‌شود و تعرفه/سهم‌ها صفر هستند.
func CalculateServicePrice(
	svc services.ServiceItem,
	org *organization.Response,
	userType user.UserType,
	uncovered bool,
) (Calculate, error) {
	svc = applySpecialistProfessionalBoost(svc, userType)

	if uncovered {
		amount, err := calculateUncoveredAmount(svc, org)
		if err != nil {
			return Calculate{}, err
		}
		calc := Calculate{TotalAmount: amount}
		calc.FundAmount = fundAmount(calc)
		return calc, nil
	}

	center := packageAmounts{
		Technical:    org.CenterPackage.TechnicalCoefficient,
		Professional: org.CenterPackage.TechnicalProfessionalCoefficient,
		Consumption:  org.CenterPackage.ConsumptionCoefficient,
	}
	pkg := packageAmounts{
		Technical:    org.Package.TechnicalCoefficient,
		Professional: org.Package.TechnicalProfessionalCoefficient,
		Consumption:  org.Package.ConsumptionCoefficient,
	}

	var amount, tariffAmount, organizationAmount int64

	if hasServiceCoefficients(svc) {
		// نرخ از بسته مرکز؛ تعرفه و سهم سازمان از بسته تعرفه (فقط اگر پایه باشد)
		amount = calcAmountFromCenterPackage(svc, center)
		if !org.IsTakmili {
			tariffAmount = calcTariffFromPackage(svc, pkg)
			organizationAmount = calcOrganizationShare(tariffAmount, org.Package.OrganizationPercentage)
		}
	} else {
		// بدون ضریب: نرخ و تعرفه ثابت خدمت؛ از بسته‌ها فقط درصد سهم سازمان مهم است
		if userType == user.UserTypeSpecialist {
			amount = int64(svc.SpecialistRate)
			tariffAmount = int64(svc.SpecialistTariff)
		} else {
			amount = int64(svc.ServiceRate)
			tariffAmount = int64(svc.ServiceTariff)
		}
		if !org.IsTakmili {
			organizationAmount = calcOrganizationShare(tariffAmount, org.Package.OrganizationPercentage)
		}
	}

	supplementaryAmount := calcSupplementaryShare(amount, org.Package.SupplementaryPercentage)
	subsidyAmount := calcSubsidyShare(amount, organizationAmount, org.Package.SubsidyPercentage)

	calc := Calculate{
		TotalAmount:       amount,
		Tariff:            tariffAmount,
		OrganizationShare: organizationAmount,
		SupplementAmount:  supplementaryAmount,
		SubsidyAmount:     subsidyAmount,
	}
	calc.FundAmount = fundAmount(calc)
	return calc, nil
}
