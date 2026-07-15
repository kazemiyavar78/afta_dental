// این ماژول نمونه است؛ باقی Entity ها و Endpoint ها طبق همین الگو در فازهای بعدی تکمیل می‌شوند.
package organization

import (
	"fmt"

	organizationpackage "github.com/tpdenta/afta-reception/internal/organizationPackage"
	"github.com/tpdenta/afta-reception/internal/platform/apperror"
	"github.com/tpdenta/afta-reception/internal/platform/security/audit"
	"github.com/tpdenta/afta-reception/internal/platform/security/encryption"
	"gorm.io/gorm"
)

// Service لایه منطق کسب‌وکار سازمان.
type Service struct {
	repo       Repository
	audit      *audit.Manager
	encryptSvc *encryption.OrganizationEncryptionService
	packageSvc *organizationpackage.Service
}

// NewService نمونه Service سازمان را می‌سازد.
func NewService(
	db *gorm.DB,
	auditMgr *audit.Manager,
	encryptSvc *encryption.OrganizationEncryptionService,
	packageSvc *organizationpackage.Service,
) *Service {
	return &Service{
		repo:       NewRepository(db),
		audit:      auditMgr,
		encryptSvc: encryptSvc,
		packageSvc: packageSvc,
	}
}

// toSensitiveData مدل سازمان را به داده حساس رمزنگاری تبدیل می‌کند.
func toSensitiveData(o *Organization) encryption.OrganizationSensitiveData {
	return encryption.OrganizationSensitiveData{
		Name:      o.Name,
		IsTakmili: o.IsTakmili,
		IsActive:  o.IsActive,
		PackageID: o.PackageID,
	}
}

// toResponse مدل دامنه را به DTO پاسخ تبدیل می‌کند و بسته preload‌شده را برای محاسبات داخلی نگه می‌دارد.
func toResponse(o *Organization) *Response {
	return &Response{
		ID:          o.ID,
		Name:        o.Name,
		IsTakmili:   o.IsTakmili,
		IsActive:    o.IsActive,
		PackageID:   o.PackageID,
		PackageName: o.Package.PackageName,
		Package:     o.Package,
	}
}

// ensurePackageExists وجود بسته تعرفه را برای انتصاب بررسی می‌کند.
func (s *Service) ensurePackageExists(packageID uint) error {
	if packageID == 0 {
		return apperror.New("VALIDATION_ERROR", "انتخاب بسته تعرفه الزامی است.", "package_id required", 400)
	}
	exists, err := s.packageSvc.Exists(int(packageID))
	if err != nil {
		return apperror.New("DB_ERROR", "خطا در بررسی بسته تعرفه.", err.Error(), 500)
	}
	if !exists {
		return apperror.New("VALIDATION_ERROR", "بسته تعرفه انتخاب‌شده یافت نشد.", "package not found", 400)
	}
	return nil
}

// verifyIntegrity هش یکپارچگی سازمان را با داده‌های فعلی بررسی می‌کند؛ در صورت نقض، رویداد دستکاری ثبت و خطا برمی‌گرداند.
func (s *Service) verifyIntegrity(o *Organization, actorID int, ip string) error {
	if !s.encryptSvc.CheckUserSecurityCode(toSensitiveData(o), o.IntegrityHash) {
		_ = s.audit.LogEvent(&actorID, ip, audit.EventDataTampering,
			fmt.Sprintf("نقض یکپارچگی سازمان %d (%s)", o.ID, o.Name))
		return apperror.ErrIntegrity
	}
	return nil
}

// Create سازمان جدید می‌سازد، بسته را منتسب می‌کند و هش امنیتی آن را تولید می‌کند.
func (s *Service) Create(req CreateRequest, actorID int, ip string) (*Response, error) {
	if err := s.ensurePackageExists(req.PackageID); err != nil {
		return nil, err
	}

	o := &Organization{
		Name:      req.Name,
		IsTakmili: req.IsTakmili,
		IsActive:  req.IsActive,
		PackageID: req.PackageID,
	}

	integrityHash, err := s.encryptSvc.CreateSecurityCode(toSensitiveData(o))
	if err != nil {
		return nil, apperror.New("ENCRYPTION_ERROR", "خطا در ایجاد هش امنیتی سازمان.", err.Error(), 500)
	}
	o.IntegrityHash = integrityHash

	if err := s.repo.Create(o); err != nil {
		return nil, apperror.New("DB_ERROR", "خطا در ایجاد سازمان.", err.Error(), 500)
	}

	created, err := s.repo.FindByID(o.ID)
	if err != nil {
		return toResponse(o), nil
	}
	_ = s.audit.LogEvent(&actorID, ip, audit.EventUserDataChange, fmt.Sprintf("ایجاد سازمان %s", o.Name))
	return toResponse(created), nil
}

// Get سازمان را با شناسه برمی‌گرداند.
func (s *Service) Get(id uint) (*Response, error) {
	o, err := s.repo.FindByID(id)
	if err == gorm.ErrRecordNotFound {
		return nil, apperror.ErrNotFound
	}
	if err != nil {
		return nil, apperror.New("DB_ERROR", "خطا در خواندن سازمان.", err.Error(), 500)
	}
	return toResponse(o), nil
}

// List لیست همه سازمان‌ها را برمی‌گرداند.
func (s *Service) List() ([]Response, error) {
	list, err := s.repo.FindAll()
	if err != nil {
		return nil, apperror.New("DB_ERROR", "خطا در خواندن سازمان‌ها.", err.Error(), 500)
	}
	var result []Response
	for i := range list {
		result = append(result, *toResponse(&list[i]))
	}
	return result, nil
}

// Update پس از تایید هش یکپارچگی، سازمان را بروزرسانی (شامل انتصاب بسته) و هش جدید تولید می‌کند.
func (s *Service) Update(id uint, req UpdateRequest, actorID int, ip string) (*Response, error) {
	if err := s.ensurePackageExists(req.PackageID); err != nil {
		return nil, err
	}

	o, err := s.repo.FindByID(id)
	if err == gorm.ErrRecordNotFound {
		return nil, apperror.ErrNotFound
	}
	if err != nil {
		return nil, apperror.New("DB_ERROR", "خطا در خواندن سازمان.", err.Error(), 500)
	}

	if err := s.verifyIntegrity(o, actorID, ip); err != nil {
		return nil, err
	}

	o.Name = req.Name
	o.IsTakmili = req.IsTakmili
	o.IsActive = req.IsActive
	o.PackageID = req.PackageID

	integrityHash, err := s.encryptSvc.CreateSecurityCode(toSensitiveData(o))
	if err != nil {
		return nil, apperror.New("ENCRYPTION_ERROR", "خطا در ایجاد هش امنیتی سازمان.", err.Error(), 500)
	}
	o.IntegrityHash = integrityHash

	if err := s.repo.Update(o); err != nil {
		return nil, apperror.New("DB_ERROR", "خطا در بروزرسانی سازمان.", err.Error(), 500)
	}

	updated, err := s.repo.FindByID(id)
	if err != nil {
		return toResponse(o), nil
	}
	_ = s.audit.LogEvent(&actorID, ip, audit.EventUserDataChange, fmt.Sprintf("بروزرسانی سازمان %s", o.Name))
	return toResponse(updated), nil
}

// Delete پس از تایید هش یکپارچگی، سازمان را به‌صورت soft-delete حذف می‌کند.
func (s *Service) Delete(id uint, actorID int, ip string) error {
	o, err := s.repo.FindByID(id)
	if err == gorm.ErrRecordNotFound {
		return apperror.ErrNotFound
	}
	if err != nil {
		return apperror.New("DB_ERROR", "خطا در خواندن سازمان.", err.Error(), 500)
	}

	if err := s.verifyIntegrity(o, actorID, ip); err != nil {
		return err
	}

	if err := s.repo.Delete(o); err != nil {
		return apperror.New("DB_ERROR", "خطا در حذف سازمان.", err.Error(), 500)
	}
	_ = s.audit.LogEvent(&actorID, ip, audit.EventUserDataChange, fmt.Sprintf("حذف سازمان %s", o.Name))
	return nil
}
