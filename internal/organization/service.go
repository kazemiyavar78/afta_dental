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
		Name:            o.Name,
		IsTakmili:       o.IsTakmili,
		IsActive:        o.IsActive,
		IsFree:          o.IsFree,
		PackageID:       o.PackageID,
		CenterPackageID: o.CenterPackageID,
	}
}

// toResponse مدل دامنه را به DTO پاسخ تبدیل می‌کند و بسته‌های preload‌شده را برای محاسبات داخلی نگه می‌دارد.
func toResponse(o *Organization) *Response {
	return &Response{
		ID:                o.ID,
		Name:              o.Name,
		IsTakmili:         o.IsTakmili,
		IsActive:          o.IsActive,
		IsFree:            o.IsFree,
		PackageID:         o.PackageID,
		PackageName:       o.Package.PackageName,
		CenterPackageID:   o.CenterPackageID,
		CenterPackageName: o.CenterPackage.PackageName,
		Package:           o.Package,
		CenterPackage:     o.CenterPackage,
	}
}

// ensurePackageExists وجود بسته را برای انتصاب بررسی می‌کند.
func (s *Service) ensurePackageExists(packageID uint, requiredMsg, notFoundMsg string) error {
	if packageID == 0 {
		return apperror.New("VALIDATION_ERROR", requiredMsg, "package_id required", 400)
	}
	exists, err := s.packageSvc.Exists(int(packageID))
	if err != nil {
		return apperror.New("DB_ERROR", "خطا در بررسی بسته.", err.Error(), 500)
	}
	if !exists {
		return apperror.New("VALIDATION_ERROR", notFoundMsg, "package not found", 400)
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

// applyFreeFlag تیک سازمان آزاد را اعمال می‌کند؛ فقط یک سازمان پایه می‌تواند آزاد باشد.
func (s *Service) applyFreeFlag(o *Organization, isFree bool) error {
	if !isFree {
		o.IsFree = false
		return nil
	}
	if o.IsTakmili {
		return apperror.New("VALIDATION_ERROR", "سازمان آزاد فقط برای بیمه پایه مجاز است.", "is_free on takmili", 400)
	}
	if err := s.clearOtherFreeOrganizations(o.ID); err != nil {
		return apperror.New("DB_ERROR", "خطا در بروزرسانی سازمان آزاد.", err.Error(), 500)
	}
	o.IsFree = true
	return nil
}

// clearOtherFreeOrganizations تیک آزاد را از سایر سازمان‌ها برمی‌دارد و هش آن‌ها را بازمحاسبه می‌کند.
func (s *Service) clearOtherFreeOrganizations(exceptID uint) error {
	list, err := s.repo.FindAll()
	if err != nil {
		return err
	}
	for i := range list {
		o := &list[i]
		if !o.IsFree || o.ID == exceptID {
			continue
		}
		o.IsFree = false
		if err := s.recalculateIntegrityHash(o); err != nil {
			return err
		}
		if err := s.repo.Update(o); err != nil {
			return err
		}
	}
	return nil
}

// recalculateIntegrityHash هش یکپارچگی سازمان را با فرمول فعلی بازمحاسبه می‌کند.
func (s *Service) recalculateIntegrityHash(o *Organization) error {
	integrityHash, err := s.encryptSvc.CreateSecurityCode(toSensitiveData(o))
	if err != nil {
		return apperror.New("ENCRYPTION_ERROR", "خطا در ایجاد هش امنیتی سازمان.", err.Error(), 500)
	}
	o.IntegrityHash = integrityHash
	return nil
}

// GetFreeOrganization سازمان آزاد را برمی‌گرداند.
func (s *Service) GetFreeOrganization() (*Response, error) {
	o, err := s.repo.FindFree()
	if err == gorm.ErrRecordNotFound {
		return nil, apperror.New("VALIDATION_ERROR", "سازمان آزاد تعریف نشده است.", "free organization not found", 400)
	}
	if err != nil {
		return nil, apperror.New("DB_ERROR", "خطا در خواندن سازمان آزاد.", err.Error(), 500)
	}
	full, err := s.repo.FindByID(o.ID)
	if err != nil {
		return nil, apperror.New("DB_ERROR", "خطا در خواندن سازمان آزاد.", err.Error(), 500)
	}
	return toResponse(full), nil
}

// Create سازمان جدید می‌سازد، بسته‌ها را منتسب می‌کند و هش امنیتی آن را تولید می‌کند.
func (s *Service) Create(req CreateRequest, actorID int, ip string) (*Response, error) {
	if err := s.ensurePackageExists(req.PackageID, "انتخاب بسته تعرفه الزامی است.", "بسته تعرفه انتخاب‌شده یافت نشد."); err != nil {
		return nil, err
	}
	if err := s.ensurePackageExists(req.CenterPackageID, "انتخاب بسته مرکز الزامی است.", "بسته مرکز انتخاب‌شده یافت نشد."); err != nil {
		return nil, err
	}

	o := &Organization{
		Name:            req.Name,
		IsTakmili:       req.IsTakmili,
		IsActive:        req.IsActive,
		PackageID:       req.PackageID,
		CenterPackageID: req.CenterPackageID,
	}

	if err := s.applyFreeFlag(o, req.IsFree); err != nil {
		return nil, err
	}

	if err := s.recalculateIntegrityHash(o); err != nil {
		return nil, err
	}

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

// Update پس از تایید هش یکپارچگی، سازمان را بروزرسانی (شامل انتصاب بسته‌ها) و هش جدید تولید می‌کند.
func (s *Service) Update(id uint, req UpdateRequest, actorID int, ip string) (*Response, error) {
	if err := s.ensurePackageExists(req.PackageID, "انتخاب بسته تعرفه الزامی است.", "بسته تعرفه انتخاب‌شده یافت نشد."); err != nil {
		return nil, err
	}
	if err := s.ensurePackageExists(req.CenterPackageID, "انتخاب بسته مرکز الزامی است.", "بسته مرکز انتخاب‌شده یافت نشد."); err != nil {
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
	o.CenterPackageID = req.CenterPackageID

	if err := s.applyFreeFlag(o, req.IsFree); err != nil {
		return nil, err
	}

	if err := s.recalculateIntegrityHash(o); err != nil {
		return nil, err
	}

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

// FixIntegrityHashes پس از تغییر فرمول هش (مثلاً افزودن فیلد)، رکوردهای معتبر قدیمی را با فرمول جدید به‌روز می‌کند.
// فقط سازمان‌هایی مهاجرت می‌شوند که هش خالی دارند یا هنوز با فرمول legacy معتبرند؛ دستکاری واقعی دست‌نخورده می‌ماند.
func (s *Service) FixIntegrityHashes() error {
	list, err := s.repo.FindAll()
	if err != nil {
		return err
	}

	for i := range list {
		o := &list[i]
		centerFilled := false

		// پر کردن CenterPackageID برای رکوردهای قبل از افزودن ستون
		if o.CenterPackageID == 0 && o.PackageID != 0 {
			o.CenterPackageID = o.PackageID
			centerFilled = true
		}

		// اگر CenterPackageID تازه پر شده، هش قبلی دیگر معتبر نیست و باید از نو ساخته شود
		if !centerFilled && s.encryptSvc.CheckUserSecurityCode(toSensitiveData(o), o.IntegrityHash) {
			continue
		}

		legacyV2OK := o.IntegrityHash != "" && s.encryptSvc.CheckUserSecurityCodeLegacyV2(toSensitiveData(o), o.IntegrityHash)
		legacyOK := o.IntegrityHash != "" && s.encryptSvc.CheckUserSecurityCodeLegacy(toSensitiveData(o), o.IntegrityHash)
		if !centerFilled && o.IntegrityHash != "" && !legacyV2OK && !legacyOK {
			// هش نه با فرمول جدید و نه قدیمی جور است → احتمال دستکاری؛ مهاجرت نکن
			continue
		}

		if err := s.recalculateIntegrityHash(o); err != nil {
			return err
		}
		if err := s.repo.Update(o); err != nil {
			return err
		}
	}
	return nil
}

// RecalculateAllIntegrityHashes همه هش‌های سازمان را با فرمول فعلی از نو می‌سازد (مهاجرت یک‌باره بعد از تغییر مدل).
func (s *Service) RecalculateAllIntegrityHashes() (int, error) {
	list, err := s.repo.FindAll()
	if err != nil {
		return 0, err
	}

	updated := 0
	for i := range list {
		o := &list[i]
		if o.CenterPackageID == 0 && o.PackageID != 0 {
			o.CenterPackageID = o.PackageID
		}

		if err := s.recalculateIntegrityHash(o); err != nil {
			return updated, err
		}
		if err := s.repo.Update(o); err != nil {
			return updated, err
		}
		updated++
	}
	return updated, nil
}
