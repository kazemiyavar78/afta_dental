package organizationpackage

import (
	"fmt"

	"github.com/tpdenta/afta-reception/internal/platform/apperror"
	"github.com/tpdenta/afta-reception/internal/platform/security/audit"
	"github.com/tpdenta/afta-reception/internal/platform/security/encryption"
	"gorm.io/gorm"
)

// Service لایه منطق کسب‌وکار بسته‌های تعرفه.
type Service struct {
	repo       Repository
	audit      *audit.Manager
	encryptSvc *encryption.OrganizationPackageEncryptionService
}

// NewService نمونه Service بسته‌های تعرفه را می‌سازد.
func NewService(db *gorm.DB, auditMgr *audit.Manager, encryptSvc *encryption.OrganizationPackageEncryptionService) *Service {
	return &Service{repo: NewRepository(db), audit: auditMgr, encryptSvc: encryptSvc}
}

// toSensitiveData مدل بسته را به داده حساس رمزنگاری تبدیل می‌کند.
func toSensitiveData(o *OrganizationPackage) encryption.OrganizationPackageSensitiveData {
	return encryption.OrganizationPackageSensitiveData{
		PackageName:                      o.PackageName,
		PackageDescription:               o.PackageDescription,
		TechnicalCoefficient:             o.TechnicalCoefficient,
		TechnicalProfessionalCoefficient: o.TechnicalProfessionalCoefficient,
		ConsumptionCoefficient:           o.ConsumptionCoefficient,
		SubsidyPercentage:                o.SubsidyPercentage,
		SupplementaryPercentage:          o.SupplementaryPercentage,
		OrganizationPercentage:           o.OrganizationPercentage,
	}
}

// toResponse مدل دامنه را به DTO پاسخ تبدیل می‌کند.
func toResponse(o *OrganizationPackage) *Response {
	return &Response{
		ID:                               o.ID,
		PackageName:                      o.PackageName,
		PackageDescription:               o.PackageDescription,
		TechnicalCoefficient:             o.TechnicalCoefficient,
		TechnicalProfessionalCoefficient: o.TechnicalProfessionalCoefficient,
		ConsumptionCoefficient:           o.ConsumptionCoefficient,
		SubsidyPercentage:                o.SubsidyPercentage,
		SupplementaryPercentage:          o.SupplementaryPercentage,
		OrganizationPercentage:           o.OrganizationPercentage,
	}
}

// validatePercentages محدوده درصدها را بررسی می‌کند.
func validatePercentages(subsidy, supplementary int) error {
	if subsidy < 0 || subsidy > 100 {
		return apperror.New("VALIDATION_ERROR", "درصد یارانه باید بین ۰ تا ۱۰۰ باشد.", "invalid subsidy percentage", 400)
	}
	if supplementary < 0 || supplementary > 100 {
		return apperror.New("VALIDATION_ERROR", "درصد تکمیلی باید بین ۰ تا ۱۰۰ باشد.", "invalid supplementary percentage", 400)
	}
	return nil
}

// applyRequest فیلدهای درخواست را روی مدل بسته اعمال می‌کند.
func applyRequest(o *OrganizationPackage, req CreateRequest) {
	o.PackageName = req.PackageName
	o.PackageDescription = req.PackageDescription
	o.TechnicalCoefficient = req.TechnicalCoefficient
	o.TechnicalProfessionalCoefficient = req.TechnicalProfessionalCoefficient
	o.ConsumptionCoefficient = req.ConsumptionCoefficient
	o.SubsidyPercentage = req.SubsidyPercentage
	o.SupplementaryPercentage = req.SupplementaryPercentage
	o.OrganizationPercentage = req.OrganizationPercentage
}

// verifyIntegrity هش یکپارچگی بسته را بررسی می‌کند؛ در صورت نقض، رویداد دستکاری ثبت و خطا برمی‌گرداند.
func (s *Service) verifyIntegrity(o *OrganizationPackage, actorID int, ip string) error {
	if !s.encryptSvc.CheckSecurityCode(toSensitiveData(o), o.IntegrityHash) {
		_ = s.audit.LogEvent(&actorID, ip, audit.EventDataTampering,
			fmt.Sprintf("نقض یکپارچگی بسته تعرفه %d (%s)", o.ID, o.PackageName))
		return apperror.ErrIntegrity
	}
	return nil
}

// Exists بررسی وجود بسته با شناسه داده‌شده.
func (s *Service) Exists(id int) (bool, error) {
	_, err := s.repo.FindByID(id)
	if err == gorm.ErrRecordNotFound {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

// Create بسته تعرفه جدید می‌سازد و هش امنیتی آن را تولید می‌کند.
func (s *Service) Create(req CreateRequest, actorID int, ip string) (*Response, error) {
	if err := validatePercentages(req.SubsidyPercentage, req.SupplementaryPercentage); err != nil {
		return nil, err
	}

	o := &OrganizationPackage{}
	applyRequest(o, req)

	integrityHash, err := s.encryptSvc.CreateSecurityCode(toSensitiveData(o))
	if err != nil {
		return nil, apperror.New("ENCRYPTION_ERROR", "خطا در ایجاد هش امنیتی بسته تعرفه.", err.Error(), 500)
	}
	o.IntegrityHash = integrityHash

	if err := s.repo.Create(o); err != nil {
		return nil, apperror.New("DB_ERROR", "خطا در ایجاد بسته تعرفه.", err.Error(), 500)
	}
	_ = s.audit.LogEvent(&actorID, ip, audit.EventUserDataChange, fmt.Sprintf("ایجاد بسته تعرفه %s", o.PackageName))
	return toResponse(o), nil
}

// Get بسته تعرفه را با شناسه برمی‌گرداند.
func (s *Service) Get(id int) (*Response, error) {
	o, err := s.repo.FindByID(id)
	if err == gorm.ErrRecordNotFound {
		return nil, apperror.ErrNotFound
	}
	if err != nil {
		return nil, apperror.New("DB_ERROR", "خطا در خواندن بسته تعرفه.", err.Error(), 500)
	}
	return toResponse(o), nil
}

// List لیست همه بسته‌های تعرفه را برمی‌گرداند.
func (s *Service) List() ([]Response, error) {
	list, err := s.repo.FindAll()
	if err != nil {
		return nil, apperror.New("DB_ERROR", "خطا در خواندن بسته‌های تعرفه.", err.Error(), 500)
	}
	result := make([]Response, 0, len(list))
	for i := range list {
		result = append(result, *toResponse(&list[i]))
	}
	return result, nil
}

// Update پس از تایید هش یکپارچگی، بسته تعرفه را بروزرسانی و هش جدید تولید می‌کند.
func (s *Service) Update(id int, req UpdateRequest, actorID int, ip string) (*Response, error) {
	if err := validatePercentages(req.SubsidyPercentage, req.SupplementaryPercentage); err != nil {
		return nil, err
	}

	o, err := s.repo.FindByID(id)
	if err == gorm.ErrRecordNotFound {
		return nil, apperror.ErrNotFound
	}
	if err != nil {
		return nil, apperror.New("DB_ERROR", "خطا در خواندن بسته تعرفه.", err.Error(), 500)
	}

	if err := s.verifyIntegrity(o, actorID, ip); err != nil {
		return nil, err
	}

	applyRequest(o, CreateRequest(req))

	integrityHash, err := s.encryptSvc.CreateSecurityCode(toSensitiveData(o))
	if err != nil {
		return nil, apperror.New("ENCRYPTION_ERROR", "خطا در ایجاد هش امنیتی بسته تعرفه.", err.Error(), 500)
	}
	o.IntegrityHash = integrityHash

	if err := s.repo.Update(o); err != nil {
		return nil, apperror.New("DB_ERROR", "خطا در بروزرسانی بسته تعرفه.", err.Error(), 500)
	}
	_ = s.audit.LogEvent(&actorID, ip, audit.EventUserDataChange, fmt.Sprintf("بروزرسانی بسته تعرفه %s", o.PackageName))
	return toResponse(o), nil
}

// Delete پس از تایید هش یکپارچگی، بسته تعرفه را به‌صورت soft-delete حذف می‌کند.
func (s *Service) Delete(id int, actorID int, ip string) error {
	o, err := s.repo.FindByID(id)
	if err == gorm.ErrRecordNotFound {
		return apperror.ErrNotFound
	}
	if err != nil {
		return apperror.New("DB_ERROR", "خطا در خواندن بسته تعرفه.", err.Error(), 500)
	}

	if err := s.verifyIntegrity(o, actorID, ip); err != nil {
		return err
	}

	if err := s.repo.Delete(o); err != nil {
		return apperror.New("DB_ERROR", "خطا در حذف بسته تعرفه.", err.Error(), 500)
	}
	_ = s.audit.LogEvent(&actorID, ip, audit.EventUserDataChange, fmt.Sprintf("حذف بسته تعرفه %s", o.PackageName))
	return nil
}
