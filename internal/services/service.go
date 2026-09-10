package services

import (
	"fmt"

	"github.com/tpdenta/afta-reception/internal/platform/apperror"
	"github.com/tpdenta/afta-reception/internal/platform/security/audit"
	"github.com/tpdenta/afta-reception/internal/platform/security/encryption"
	"github.com/tpdenta/afta-reception/internal/platform/security/integrity"
	"gorm.io/gorm"
)

var allowedServiceFeatures = map[string]bool{
	"":   true,
	"#":  true,
	"*":  true,
	"#*": true,
}

// Service لایه منطق کسب‌وکار خدمات.
type Service struct {
	repo       Repository
	audit      *audit.Manager
	encryptSvc *encryption.ServiceEncryptionService
	signer     *integrity.Signer
}

// NewService نمونه Service خدمات را می‌سازد.
func NewService(
	db *gorm.DB,
	auditMgr *audit.Manager,
	encryptSvc *encryption.ServiceEncryptionService,
	signer *integrity.Signer,
) *Service {
	return &Service{repo: NewRepository(db), audit: auditMgr, encryptSvc: encryptSvc, signer: signer}
}

// toSensitiveData مدل خدمت را به داده حساس رمزنگاری تبدیل می‌کند.
func toSensitiveData(item *ServiceItem) encryption.ServiceSensitiveData {
	return encryption.ServiceSensitiveData{
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

// toResponse مدل دامنه را به DTO پاسخ تبدیل می‌کند.
func toResponse(item *ServiceItem) *Response {
	return &Response{
		ID:                      item.ID,
		ServiceCode:             item.ServiceCode,
		Name:                    item.Name,
		TechnicalCoefficient:    item.TechnicalCoefficient,
		ProfessionalCoefficient: item.ProfessionalCoefficient,
		ConsumptionCoefficient:  item.ConsumptionCoefficient,
		ServiceRate:             item.ServiceRate,
		ServiceTariff:           item.ServiceTariff,
		SpecialistRate:          item.SpecialistRate,
		SpecialistTariff:        item.SpecialistTariff,
		InternationalCode:       item.InternationalCode,
		DefaultCount:            item.DefaultCount,
		MaximumCount:            item.MaximumCount,
		ServiceFeatures:         item.ServiceFeatures,
		IsActive:                item.IsActive,
		IsDentalDirection:       item.HasDentalDirection,
		HasTooth:                item.HasTooth,
		AllowMultipleUse:        item.AllowMultipleUse,
	}
}

// validateFeatures معتبر بودن ویژگی خدمت را بررسی می‌کند.
func validateFeatures(features string) error {
	if !allowedServiceFeatures[features] {
		return apperror.New("VALIDATION_ERROR", "ویژگی خدمت نامعتبر است. مقادیر مجاز: خالی، #، *، #*", "invalid service features", 400)
	}
	return nil
}

// verifyIntegrity هش یکپارچگی خدمت را بررسی می‌کند؛ در صورت نقض، رویداد دستکاری ثبت و خطا برمی‌گرداند.
func (s *Service) verifyIntegrity(item *ServiceItem, actorID int, ip string) error {
	if !s.encryptSvc.CheckSecurityCode(toSensitiveData(item), item.IntegrityHash) {
		_ = s.audit.LogEvent(&actorID, ip, audit.EventDataTampering,
			fmt.Sprintf("نقض یکپارچگی خدمت %d (%s)", item.ID, item.Name))
		return apperror.ErrIntegrity
	}
	return nil
}

// applyRequest فیلدهای درخواست را روی مدل خدمت اعمال می‌کند.
// SpecialistRate و SpecialistTariff در هش یکپارچگی لحاظ نمی‌شوند.
func applyRequest(item *ServiceItem, code, name, intlCode,
	features string, tech, prof, cons float64, rate, tariff, specialistRate, specialistTariff int, defCount, maxCount int, isActive bool, hasDentalDirection, hasTooth, allowMultipleUse bool) {
	item.ServiceCode = code
	item.Name = name
	item.TechnicalCoefficient = tech
	item.ProfessionalCoefficient = prof
	item.ConsumptionCoefficient = cons
	item.ServiceRate = rate
	item.ServiceTariff = tariff
	item.SpecialistRate = specialistRate
	item.SpecialistTariff = specialistTariff
	item.InternationalCode = intlCode
	item.DefaultCount = defCount
	item.MaximumCount = maxCount
	item.ServiceFeatures = features
	item.IsActive = isActive
	item.HasDentalDirection = hasDentalDirection
	item.HasTooth = hasTooth
	item.AllowMultipleUse = allowMultipleUse
}

// Create خدمت جدید می‌سازد و هش امنیتی آن را تولید می‌کند.
func (s *Service) Create(req CreateRequest, actorID int, ip string) (*Response, error) {
	if err := validateFeatures(req.ServiceFeatures); err != nil {
		return nil, err
	}

	item := &ServiceItem{}
	applyRequest(item, req.ServiceCode, req.Name, req.InternationalCode, req.ServiceFeatures,
		req.TechnicalCoefficient, req.ProfessionalCoefficient, req.ConsumptionCoefficient,
		req.ServiceRate, req.ServiceTariff, req.SpecialistRate, req.SpecialistTariff,
		req.DefaultCount, req.MaximumCount, req.IsActive, req.IsDentalDirection, req.HasTooth, req.AllowMultipleUse)

	integrityHash, err := s.encryptSvc.CreateSecurityCode(toSensitiveData(item))
	if err != nil {
		return nil, apperror.New("ENCRYPTION_ERROR", "خطا در ایجاد هش امنیتی خدمت.", err.Error(), 500)
	}
	item.IntegrityHash = integrityHash

	if err := s.repo.Create(item); err != nil {
		return nil, apperror.New("DB_ERROR", "خطا در ایجاد خدمت.", err.Error(), 500)
	}
	_ = s.audit.LogEvent(&actorID, ip, audit.EventUserDataChange, fmt.Sprintf("ایجاد خدمت %s", item.Name))
	return toResponse(item), nil
}

// Get خدمت را با شناسه برمی‌گرداند.
func (s *Service) Get(id uint) (*Response, error) {
	item, err := s.repo.FindByID(id)
	if err == gorm.ErrRecordNotFound {
		return nil, apperror.ErrNotFound
	}
	if err != nil {
		return nil, apperror.New("DB_ERROR", "خطا در خواندن خدمت.", err.Error(), 500)
	}
	return toResponse(item), nil
}

// List لیست همه خدمات را برمی‌گرداند.
func (s *Service) List() ([]Response, error) {
	list, err := s.repo.FindAll()
	if err != nil {
		return nil, apperror.New("DB_ERROR", "خطا در خواندن خدمات.", err.Error(), 500)
	}
	var result []Response
	for i := range list {
		result = append(result, *toResponse(&list[i]))
	}
	return result, nil
}

// FindByExcludeServices خدمات را به جز شناسه‌های ارسالی برمی‌گرداند.
func (s *Service) FindByExcludeServices(excludeServices []uint) ([]ServiceItem, error) {
	list, err := s.repo.FindByExcludeServices(excludeServices)
	if err != nil {
		return nil, apperror.New("DB_ERROR", "خطا در خواندن خدمات.", err.Error(), 500)
	}

	return list, nil
}

// FindItemByCode مدل دامنه خدمت را با کد خدمت برمی‌گرداند.
func (s *Service) FindItemByCode(code string) (*ServiceItem, error) {
	item, err := s.repo.FindByServiceCode(code)
	if err == gorm.ErrRecordNotFound {
		return nil, apperror.ErrNotFound
	}
	if err != nil {
		return nil, apperror.New("DB_ERROR", "خطا در خواندن خدمت.", err.Error(), 500)
	}
	return item, nil
}

// FindItemByID مدل دامنه خدمت را با شناسه برمی‌گرداند (برای محاسبات داخلی مثل تعرفه).
func (s *Service) FindItemByID(id uint) (*ServiceItem, error) {
	item, err := s.repo.FindByID(id)
	if err == gorm.ErrRecordNotFound {
		return nil, apperror.ErrNotFound
	}
	if err != nil {
		return nil, apperror.New("DB_ERROR", "خطا در خواندن خدمت.", err.Error(), 500)
	}
	return item, nil
}

// Update پس از تایید هش یکپارچگی، خدمت را بروزرسانی و هش جدید تولید می‌کند.
func (s *Service) Update(id uint, req UpdateRequest, actorID int, ip string) (*Response, error) {
	if err := validateFeatures(req.ServiceFeatures); err != nil {
		return nil, err
	}

	item, err := s.repo.FindByID(id)
	if err == gorm.ErrRecordNotFound {
		return nil, apperror.ErrNotFound
	}
	if err != nil {
		return nil, apperror.New("DB_ERROR", "خطا در خواندن خدمت.", err.Error(), 500)
	}

	if err := s.verifyIntegrity(item, actorID, ip); err != nil {
		return nil, err
	}

	applyRequest(item, req.ServiceCode, req.Name, req.InternationalCode, req.ServiceFeatures,
		req.TechnicalCoefficient, req.ProfessionalCoefficient, req.ConsumptionCoefficient,
		req.ServiceRate, req.ServiceTariff, req.SpecialistRate, req.SpecialistTariff,
		req.DefaultCount, req.MaximumCount, req.IsActive, req.IsDentalDirection, req.HasTooth, req.AllowMultipleUse)

	integrityHash, err := s.encryptSvc.CreateSecurityCode(toSensitiveData(item))
	if err != nil {
		return nil, apperror.New("ENCRYPTION_ERROR", "خطا در ایجاد هش امنیتی خدمت.", err.Error(), 500)
	}
	item.IntegrityHash = integrityHash

	if err := s.repo.Update(item); err != nil {
		return nil, apperror.New("DB_ERROR", "خطا در بروزرسانی خدمت.", err.Error(), 500)
	}
	_ = s.audit.LogEvent(&actorID, ip, audit.EventUserDataChange, fmt.Sprintf("بروزرسانی خدمت %s", item.Name))
	return toResponse(item), nil
}

// Delete پس از تایید هش یکپارچگی، خدمت را به‌صورت soft-delete حذف می‌کند.
func (s *Service) Delete(id uint, actorID int, ip string) error {
	item, err := s.repo.FindByID(id)
	if err == gorm.ErrRecordNotFound {
		return apperror.ErrNotFound
	}
	if err != nil {
		return apperror.New("DB_ERROR", "خطا در خواندن خدمت.", err.Error(), 500)
	}

	if err := s.verifyIntegrity(item, actorID, ip); err != nil {
		return err
	}

	if err := s.repo.Delete(item); err != nil {
		return apperror.New("DB_ERROR", "خطا در حذف خدمت.", err.Error(), 500)
	}
	_ = s.audit.LogEvent(&actorID, ip, audit.EventUserDataChange, fmt.Sprintf("حذف خدمت %s", item.Name))
	return nil
}

// toExcludedResponse مدل دامنه خدمت خارج‌شده را به DTO پاسخ تبدیل می‌کند.
func toExcludedResponse(item *ExcludedService) ExcludedServiceResponse {
	return ExcludedServiceResponse{
		ID:             item.ID,
		OrganizationID: item.OrganizationID,
		ServiceID:      item.ServiceID,
	}
}

// ListExcludedServicesByOrganization خدمات خارج‌شده یک سازمان را برمی‌گرداند.
func (s *Service) ListExcludedServicesByOrganization(organizationID uint) ([]ExcludedServiceResponse, error) {
	list, err := s.repo.FindByOrganizationID(organizationID)
	if err != nil {
		return nil, apperror.New("DB_ERROR", "خطا در خواندن خدمات خارج‌شده.", err.Error(), 500)
	}
	result := make([]ExcludedServiceResponse, 0, len(list))
	for i := range list {
		result = append(result, toExcludedResponse(&list[i]))
	}
	return result, nil
}

// ExcludedServiceIDSet شناسه خدمات خارج‌ازپوشش یک سازمان را به‌صورت set برمی‌گرداند.
func (s *Service) ExcludedServiceIDSet(organizationID uint) (map[uint]struct{}, error) {
	list, err := s.repo.FindByOrganizationID(organizationID)
	if err != nil {
		return nil, apperror.New("DB_ERROR", "خطا در خواندن خدمات خارج‌شده.", err.Error(), 500)
	}
	set := make(map[uint]struct{}, len(list))
	for i := range list {
		set[list[i].ServiceID] = struct{}{}
	}
	return set, nil
}

// AddExcludedServices خدمات را به لیست خارج‌شده سازمان اضافه می‌کند و هش یکپارچگی را ثبت می‌کند.
func (s *Service) AddExcludedServices(req ExcludedServicesRequest, actorID int, ip string) ([]ExcludedServiceResponse, error) {
	if req.OrganizationID == 0 {
		return nil, apperror.New("VALIDATION_ERROR", "شناسه سازمان الزامی است.", "organization_id required", 400)
	}
	if len(req.ServiceIDs) == 0 {
		return nil, apperror.New("VALIDATION_ERROR", "حداقل یک خدمت باید انتخاب شود.", "service_ids required", 400)
	}

	result := make([]ExcludedServiceResponse, 0, len(req.ServiceIDs))
	for _, serviceID := range req.ServiceIDs {
		if serviceID == 0 {
			return nil, apperror.New("VALIDATION_ERROR", "شناسه خدمت نامعتبر است.", "invalid service_id", 400)
		}
		if _, err := s.repo.FindByID(serviceID); err != nil {
			if err == gorm.ErrRecordNotFound {
				return nil, apperror.New("VALIDATION_ERROR", "خدمت انتخاب‌شده یافت نشد.", "service not found", 400)
			}
			return nil, apperror.New("DB_ERROR", "خطا در بررسی خدمت.", err.Error(), 500)
		}

		existing, err := s.repo.FindExcludedIncludingDeleted(req.OrganizationID, serviceID)
		if err != nil && err != gorm.ErrRecordNotFound {
			return nil, apperror.New("DB_ERROR", "خطا در بررسی خدمت خارج‌شده.", err.Error(), 500)
		}

		if err == nil {
			if existing.DeletedAt.Valid {
				existing.OrganizationID = req.OrganizationID
				existing.ServiceID = serviceID
				existing.IntegrityHash = SignExcludedServiceIntegrityHash(s.signer, existing)
				if err := s.repo.RestoreExcludedService(existing); err != nil {
					return nil, apperror.New("DB_ERROR", "خطا در بازیابی خدمت خارج‌شده.", err.Error(), 500)
				}
			} else if !VerifyExcludedServiceIntegrity(s.signer, existing) {
				_ = s.audit.LogEvent(&actorID, ip, audit.EventDataTampering,
					fmt.Sprintf("نقض یکپارچگی خدمت خارج‌شده سازمان %d خدمت %d", req.OrganizationID, serviceID))
				return nil, apperror.ErrIntegrity
			}
			result = append(result, toExcludedResponse(existing))
			continue
		}

		item := &ExcludedService{
			OrganizationID: req.OrganizationID,
			ServiceID:      serviceID,
		}
		item.IntegrityHash = SignExcludedServiceIntegrityHash(s.signer, item)
		if err := s.repo.CreateExcludedService(item); err != nil {
			return nil, apperror.New("DB_ERROR", "خطا در افزودن خدمت خارج‌شده.", err.Error(), 500)
		}
		result = append(result, toExcludedResponse(item))
	}

	_ = s.audit.LogEvent(&actorID, ip, audit.EventUserDataChange,
		fmt.Sprintf("افزودن %d خدمت خارج‌شده برای سازمان %d", len(req.ServiceIDs), req.OrganizationID))
	return result, nil
}

// RemoveExcludedServices خدمات را از لیست خارج‌شده سازمان حذف می‌کند.
func (s *Service) RemoveExcludedServices(req ExcludedServicesRequest, actorID int, ip string) error {
	if req.OrganizationID == 0 {
		return apperror.New("VALIDATION_ERROR", "شناسه سازمان الزامی است.", "organization_id required", 400)
	}
	if len(req.ServiceIDs) == 0 {
		return apperror.New("VALIDATION_ERROR", "حداقل یک خدمت باید انتخاب شود.", "service_ids required", 400)
	}

	for _, serviceID := range req.ServiceIDs {
		item, err := s.repo.FindByOrganizationIDAndServiceID(req.OrganizationID, serviceID)
		if err == gorm.ErrRecordNotFound {
			continue
		}
		if err != nil {
			return apperror.New("DB_ERROR", "خطا در خواندن خدمت خارج‌شده.", err.Error(), 500)
		}
		if !VerifyExcludedServiceIntegrity(s.signer, item) {
			_ = s.audit.LogEvent(&actorID, ip, audit.EventDataTampering,
				fmt.Sprintf("نقض یکپارچگی خدمت خارج‌شده سازمان %d خدمت %d", req.OrganizationID, serviceID))
			return apperror.ErrIntegrity
		}
		if err := s.repo.RemoveExcludedService(req.OrganizationID, serviceID); err != nil {
			return apperror.New("DB_ERROR", "خطا در حذف خدمت خارج‌شده.", err.Error(), 500)
		}
	}

	_ = s.audit.LogEvent(&actorID, ip, audit.EventUserDataChange,
		fmt.Sprintf("حذف %d خدمت خارج‌شده برای سازمان %d", len(req.ServiceIDs), req.OrganizationID))
	return nil
}
