package tariff

import (
	"errors"
	"fmt"

	"github.com/tpdenta/afta-reception/internal/organization"
	"github.com/tpdenta/afta-reception/internal/platform/apperror"
	"github.com/tpdenta/afta-reception/internal/platform/security/audit"
	"github.com/tpdenta/afta-reception/internal/services"
	"gorm.io/gorm"
)

// Service لایه منطق کسب‌وکار تعرفه.
type Service struct {
	db                 *gorm.DB
	repo               Repository
	audit              *audit.Manager
	organizationService *organization.Service
	serviceService     *services.Service
}

// NewService نمونه Service تعرفه را با وابستگی به سازمان و خدمات می‌سازد.
func NewService(
	db *gorm.DB,
	auditMgr *audit.Manager,
	organizationService *organization.Service,
	serviceService *services.Service,
) *Service {
	return &Service{
		db:                  db,
		repo:                NewRepository(db),
		audit:               auditMgr,
		organizationService: organizationService,
		serviceService:      serviceService,
	}
}

// fundAmount صندوق را از مبالغ محاسبه‌شده به‌دست می‌آورد.
func fundAmount(c Calculate) int64 {
	return c.TotalAmount - c.OrganizationShare - c.SupplementAmount - c.SubsidyAmount
}

// CalculateTariffForOrganization تعرفه خدمات (به‌جز exclude) را برای سازمان محاسبه می‌کند؛ صندوق منفی در تست مجاز است.
func (s *Service) CalculateTariffForOrganization(req CalculateTariffForOrganizationRequest, actorID int, ip string) (*CalculateResponse, error) {
	organizationResp, err := s.organizationService.Get(req.OrganizationID)
	if err != nil {
		return nil, apperror.New("ORGANIZATION_NOT_FOUND", "سازمان یافت نشد.", err.Error(), 404)
	}

	for _, serviceID := range req.ExcludeServiceIDs {
		_, err = s.serviceService.Get(serviceID)
		if err != nil {
			return nil, apperror.New("SERVICE_NOT_FOUND", "خدمت مورد نظر یافت نشد.", err.Error(), 404)
		}
	}

	listServices, err := s.serviceService.FindByExcludeServices(req.ExcludeServiceIDs)
	if err != nil {
		return nil, apperror.New("SERVICE_NOT_FOUND", "خدمات مورد نظر یافت نشد.", err.Error(), 404)
	}

	result, err := s.CalculateTariff(listServices, organizationResp, req)
	if err != nil {
		return nil, apperror.New("CALCULATE_TARIFF_ERROR", "خطا در محاسبه تعرفه.", err.Error(), 500)
	}
	return &CalculateResponse{
		Services:     result,
		Organization: *organizationResp,
		SendInfo:     req,
	}, nil
}

// CalculateTariff برای هر خدمت نرخ، تعرفه، سهم‌ها، یارانه و صندوق را محاسبه می‌کند.
func (s *Service) CalculateTariff(
	listServices []services.ServiceItem,
	organizationResp *organization.Response,
	req CalculateTariffForOrganizationRequest,
) ([]ServiceWithPrice, error) {
	var result []ServiceWithPrice
	pkg := organizationResp.Package

	for _, service := range listServices {
		amount := int64(service.TechnicalCoefficient * float64(req.TechnicalAmount))
		amount += int64(service.ProfessionalCoefficient * float64(req.ProfessionalCenterAmount))
		amount += int64(service.ConsumptionCoefficient * float64(req.ConsumptionCenterAmount))

		var tariffAmount int64
		var organizationAmount int64
		var supplementaryAmount int64

		// اگر سازمان تکمیلی نباشد تعرفه و سهم سازمان و تکمیلی محاسبه نمی‌شود.
		if !organizationResp.IsTakmili {
			tariffAmount = int64(service.TechnicalCoefficient * float64(pkg.TechnicalCoefficient))
			tariffAmount += int64(service.ProfessionalCoefficient * float64(pkg.TechnicalProfessionalCoefficient))
			tariffAmount += int64(service.ConsumptionCoefficient * float64(pkg.ConsumptionCoefficient))
			organizationAmount = int64(float64(tariffAmount) * float64(pkg.OrganizationPercentage) / 100)
		}
		
		supplementaryAmount = int64(float64(amount) * float64(pkg.SupplementaryPercentage) / 100)
		subsidyAmount := int64(float64(amount-organizationAmount) * float64(pkg.SubsidyPercentage) / 100)

		calc := Calculate{
			TotalAmount:       amount,
			Tariff:            tariffAmount,
			OrganizationShare: organizationAmount,
			SupplementAmount:  supplementaryAmount,
			SubsidyAmount:     subsidyAmount,
		}
		calc.FundAmount = fundAmount(calc)

		result = append(result, ServiceWithPrice{
			Service:   toServiceResponse(&service),
			Calculate: calc,
		})
	}
	return result, nil
}

// toServiceResponse مدل دامنه خدمت را به DTO پاسخ تبدیل می‌کند.
func toServiceResponse(item *services.ServiceItem) services.Response {
	return services.Response{
		ID:                      item.ID,
		ServiceCode:             item.ServiceCode,
		Name:                    item.Name,
		TechnicalCoefficient:    item.TechnicalCoefficient,
		ProfessionalCoefficient: item.ProfessionalCoefficient,
		ConsumptionCoefficient:  item.ConsumptionCoefficient,
		ServiceRate:             item.ServiceRate,
		ServiceTariff:           item.ServiceTariff,
		InternationalCode:       item.InternationalCode,
		DefaultCount:            item.DefaultCount,
		MaximumCount:            item.MaximumCount,
		ServiceFeatures:         item.ServiceFeatures,
		IsActive:                item.IsActive,
	}
}

// SaveTariffsForOrganization محاسبه را تکرار کرده و در یک تراکنش همه ردیف‌ها را upsert می‌کند؛ اگر هر صندوق منفی باشد کل عملیات rollback می‌شود.
func (s *Service) SaveTariffsForOrganization(req CalculateTariffForOrganizationRequest, actorID int, ip string) (*SaveTariffResponse, error) {
	calcResp, err := s.CalculateTariffForOrganization(req, actorID, ip)
	if err != nil {
		return nil, err
	}
	if len(calcResp.Services) == 0 {
		return nil, apperror.New("VALIDATION_ERROR", "هیچ خدمتی برای ذخیره باقی نمانده است.", "no services to save", 400)
	}

	for _, item := range calcResp.Services {
		if item.Calculate.FundAmount < 0 {
			return nil, apperror.New(
				"VALIDATION_ERROR",
				fmt.Sprintf("صندوق خدمت «%s» منفی است و امکان ذخیره وجود ندارد.", item.Service.Name),
				"negative fund",
				400,
			)
		}
	}

	var saved []Tariff
	err = s.db.Transaction(func(tx *gorm.DB) error {
		txRepo := NewRepository(tx)
		for _, item := range calcResp.Services {
			existing, findErr := txRepo.FindByOrganizationIDAndServiceID(req.OrganizationID, item.Service.ID)
			if findErr != nil && !errors.Is(findErr, gorm.ErrRecordNotFound) {
				return apperror.New("DB_ERROR", "خطا در خواندن تعرفه موجود.", findErr.Error(), 500)
			}

			row := &Tariff{
				OrganizationID:     req.OrganizationID,
				ServiceID:          item.Service.ID,
				Amount:             item.Calculate.TotalAmount,
				TariffAmount:       item.Calculate.Tariff,
				OrganizationShare:  item.Calculate.OrganizationShare,
				SupplementaryShare: item.Calculate.SupplementAmount,
				SubsidyShare:       item.Calculate.SubsidyAmount,
			}

			if findErr == nil {
				row.Model = existing.Model
				if err := txRepo.Update(row); err != nil {
					return apperror.New("DB_ERROR", "خطا در بروزرسانی تعرفه.", err.Error(), 500)
				}
			} else {
				if err := txRepo.Create(row); err != nil {
					return apperror.New("DB_ERROR", "خطا در ایجاد تعرفه.", err.Error(), 500)
				}
			}
			saved = append(saved, *row)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	_ = s.audit.LogEvent(&actorID, ip, audit.EventUserDataChange,
		fmt.Sprintf("ذخیره تعرفه سازمان %d (%d خدمت)", req.OrganizationID, len(saved)))

	items := make([]TariffResponse, 0, len(saved))
	for i, row := range saved {
		svc := calcResp.Services[i].Service
		items = append(items, toTariffResponse(&row, svc.ServiceCode, svc.Name))
	}
	return &SaveTariffResponse{Items: items}, nil
}

// ListByOrganizationID لیست تعرفه‌های ذخیره‌شده یک سازمان را به‌همراه نام خدمت برمی‌گرداند.
func (s *Service) ListByOrganizationID(organizationID uint) ([]TariffResponse, error) {
	if _, err := s.organizationService.Get(organizationID); err != nil {
		return nil, apperror.New("ORGANIZATION_NOT_FOUND", "سازمان یافت نشد.", err.Error(), 404)
	}

	list, err := s.repo.FindAllByOrganizationID(organizationID)
	if err != nil {
		return nil, apperror.New("DB_ERROR", "خطا در خواندن تعرفه‌ها.", err.Error(), 500)
	}

	result := make([]TariffResponse, 0, len(list))
	for i := range list {
		code, name := "", ""
		if svc, svcErr := s.serviceService.Get(list[i].ServiceID); svcErr == nil && svc != nil {
			code, name = svc.ServiceCode, svc.Name
		}
		result = append(result, toTariffResponse(&list[i], code, name))
	}
	return result, nil
}

// GetByOrganizationAndService تعرفه ذخیره‌شده یک خدمت برای سازمان را برمی‌گرداند.
func (s *Service) GetByOrganizationAndService(organizationID, serviceID uint) (*Tariff, error) {
	t, err := s.repo.FindByOrganizationIDAndServiceID(organizationID, serviceID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperror.New("E-005", "خدمت بدون تعرفه است.", "tariff not found for service", 400)
	}
	if err != nil {
		return nil, apperror.New("DB_ERROR", "خطا در خواندن تعرفه.", err.Error(), 500)
	}
	return t, nil
}

// RecalculateTariff یک تعرفه خاص را با سه مبلغ مرکز دوباره محاسبه و در صورت معتبر بودن صندوق ذخیره می‌کند.
func (s *Service) RecalculateTariff(id uint, req RecalculateTariffRequest, actorID int, ip string) (*TariffResponse, error) {
	existing, err := s.repo.FindByID(id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperror.ErrNotFound
	}
	if err != nil {
		return nil, apperror.New("DB_ERROR", "خطا در خواندن تعرفه.", err.Error(), 500)
	}

	organizationResp, err := s.organizationService.Get(existing.OrganizationID)
	if err != nil {
		return nil, apperror.New("ORGANIZATION_NOT_FOUND", "سازمان یافت نشد.", err.Error(), 404)
	}

	serviceItem, err := s.serviceService.FindItemByID(existing.ServiceID)
	if err != nil {
		return nil, apperror.New("SERVICE_NOT_FOUND", "خدمت مورد نظر یافت نشد.", err.Error(), 404)
	}

	calcReq := CalculateTariffForOrganizationRequest{
		OrganizationID:           existing.OrganizationID,
		TechnicalAmount:          req.TechnicalAmount,
		ProfessionalCenterAmount: req.ProfessionalCenterAmount,
		ConsumptionCenterAmount:  req.ConsumptionCenterAmount,
	}
	calculated, err := s.CalculateTariff([]services.ServiceItem{*serviceItem}, organizationResp, calcReq)
	if err != nil || len(calculated) == 0 {
		return nil, apperror.New("CALCULATE_TARIFF_ERROR", "خطا در محاسبه تعرفه.", "calculate failed", 500)
	}

	calc := calculated[0].Calculate
	if calc.FundAmount < 0 {
		return nil, apperror.New(
			"VALIDATION_ERROR",
			fmt.Sprintf("صندوق خدمت «%s» منفی است و امکان ویرایش وجود ندارد.", serviceItem.Name),
			"negative fund",
			400,
		)
	}

	existing.Amount = calc.TotalAmount
	existing.TariffAmount = calc.Tariff
	existing.OrganizationShare = calc.OrganizationShare
	existing.SupplementaryShare = calc.SupplementAmount
	existing.SubsidyShare = calc.SubsidyAmount

	if err := s.repo.Update(existing); err != nil {
		return nil, apperror.New("DB_ERROR", "خطا در بروزرسانی تعرفه.", err.Error(), 500)
	}

	_ = s.audit.LogEvent(&actorID, ip, audit.EventUserDataChange,
		fmt.Sprintf("بازمحاسبه تعرفه %d برای سازمان %d", existing.ID, existing.OrganizationID))

	resp := toTariffResponse(existing, serviceItem.ServiceCode, serviceItem.Name)
	return &resp, nil
}

// toTariffResponse مدل دامنه را به DTO پاسخ تبدیل و صندوق را محاسبه می‌کند.
func toTariffResponse(t *Tariff, serviceCode, serviceName string) TariffResponse {
	fund := t.Amount - t.OrganizationShare - t.SupplementaryShare - t.SubsidyShare
	return TariffResponse{
		ID:                 t.ID,
		OrganizationID:     t.OrganizationID,
		ServiceID:          t.ServiceID,
		ServiceCode:        serviceCode,
		ServiceName:        serviceName,
		Amount:             t.Amount,
		TariffAmount:       t.TariffAmount,
		OrganizationShare:  t.OrganizationShare,
		SupplementaryShare: t.SupplementaryShare,
		SubsidyShare:       t.SubsidyShare,
		FundAmount:         fund,
	}
}
