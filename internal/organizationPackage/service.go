package organizationpackage

import (
	"context"

	"github.com/tpdenta/afta-reception/internal/platform/apperror"
	"github.com/tpdenta/afta-reception/internal/platform/security/audit"
	"github.com/tpdenta/afta-reception/internal/platform/security/encryption"
	"gorm.io/gorm"
)

// Service لایه منطق کسب‌وکار بسته‌های تعرفه.
type Service struct {
	repo       Repository
	audit      audit.Recorder
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
// ctx زمینه است. o بسته و actor فاعل است. شکست ممیزی قبل از ErrIntegrity برمی‌گردد.
func (s *Service) verifyIntegrity(ctx context.Context, o *OrganizationPackage, actor audit.Actor) error {
	if !s.encryptSvc.CheckSecurityCode(toSensitiveData(o), o.IntegrityHash) {
		if err := s.recordIntegrity(ctx, actor, o); err != nil {
			return err
		}
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
// ctx زمینه است. req داده فرم و actor سازنده است. خروجی پاسخ یا خطا است.
func (s *Service) Create(ctx context.Context, req CreateRequest, actor audit.Actor) (*Response, error) {
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
	if err := s.recordCreated(ctx, actor, o); err != nil {
		return nil, err
	}
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
// ctx زمینه است. id شناسه، req داده جدید و actor ویرایشگر است.
func (s *Service) Update(ctx context.Context, id int, req UpdateRequest, actor audit.Actor) (*Response, error) {
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

	if err := s.verifyIntegrity(ctx, o, actor); err != nil {
		return nil, err
	}

	before := *o
	applyRequest(o, CreateRequest(req))

	integrityHash, err := s.encryptSvc.CreateSecurityCode(toSensitiveData(o))
	if err != nil {
		return nil, apperror.New("ENCRYPTION_ERROR", "خطا در ایجاد هش امنیتی بسته تعرفه.", err.Error(), 500)
	}
	o.IntegrityHash = integrityHash

	if err := s.repo.Update(o); err != nil {
		return nil, apperror.New("DB_ERROR", "خطا در بروزرسانی بسته تعرفه.", err.Error(), 500)
	}
	if err := s.recordUpdated(ctx, actor, &before, o); err != nil {
		return nil, err
	}
	return toResponse(o), nil
}

// Delete پس از تایید هش یکپارچگی، بسته تعرفه را به‌صورت soft-delete حذف می‌کند.
// ctx زمینه است. id شناسه و actor حذف‌کننده است. خروجی خطا است.
func (s *Service) Delete(ctx context.Context, id int, actor audit.Actor) error {
	o, err := s.repo.FindByID(id)
	if err == gorm.ErrRecordNotFound {
		return apperror.ErrNotFound
	}
	if err != nil {
		return apperror.New("DB_ERROR", "خطا در خواندن بسته تعرفه.", err.Error(), 500)
	}

	if err := s.verifyIntegrity(ctx, o, actor); err != nil {
		return err
	}

	if err := s.repo.Delete(o); err != nil {
		return apperror.New("DB_ERROR", "خطا در حذف بسته تعرفه.", err.Error(), 500)
	}
	return s.recordDeleted(ctx, actor, o)
}
