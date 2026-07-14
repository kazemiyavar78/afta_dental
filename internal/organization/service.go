// این ماژول نمونه است؛ باقی Entity ها و Endpoint ها طبق همین الگو در فازهای بعدی تکمیل می‌شوند.
package organization

import (
	"fmt"

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
}

// NewService نمونه Service سازمان را می‌سازد.
func NewService(db *gorm.DB, auditMgr *audit.Manager, encryptSvc *encryption.OrganizationEncryptionService) *Service {
	return &Service{repo: NewRepository(db), audit: auditMgr, encryptSvc: encryptSvc}
}

// toResponse مدل دامنه را به DTO پاسخ تبدیل می‌کند.
func toResponse(o *Organization) *Response {
	return &Response{
		ID:        o.ID,
		Name:      o.Name,
		IsTakmili: o.IsTakmili,
		IsActive:  o.IsActive,
	}
}

// verifyIntegrity هش یکپارچگی سازمان را با داده‌های فعلی بررسی می‌کند؛ در صورت نقض، رویداد دستکاری ثبت و خطا برمی‌گرداند.
func (s *Service) verifyIntegrity(o *Organization, actorID int, ip string) error {
	data := encryption.OrganizationSensitiveData{
		Name:      o.Name,
		IsTakmili: o.IsTakmili,
		IsActive:  o.IsActive,
	}
	if !s.encryptSvc.CheckUserSecurityCode(data, o.IntegrityHash) {
		_ = s.audit.LogEvent(&actorID, ip, audit.EventDataTampering,
			fmt.Sprintf("نقض یکپارچگی سازمان %d (%s)", o.ID, o.Name))
		return apperror.ErrIntegrity
	}
	return nil
}

// Create سازمان جدید می‌سازد و هش امنیتی آن را تولید می‌کند.
func (s *Service) Create(req CreateRequest, actorID int, ip string) (*Response, error) {
	integrityHash, err := s.encryptSvc.CreateSecurityCode(encryption.OrganizationSensitiveData{
		Name:      req.Name,
		IsTakmili: req.IsTakmili,
		IsActive:  req.IsActive,
	})
	if err != nil {
		return nil, apperror.New("ENCRYPTION_ERROR", "خطا در ایجاد هش امنیتی سازمان.", err.Error(), 500)
	}
	o := &Organization{
		Name:          req.Name,
		IsTakmili:     req.IsTakmili,
		IsActive:      req.IsActive,
		IntegrityHash: integrityHash,
	}
	if err := s.repo.Create(o); err != nil {
		return nil, apperror.New("DB_ERROR", "خطا در ایجاد سازمان.", err.Error(), 500)
	}
	_ = s.audit.LogEvent(&actorID, ip, audit.EventUserDataChange, fmt.Sprintf("ایجاد سازمان %s", o.Name))
	return toResponse(o), nil
}

// Get سازمان را با شناسه برمی‌گرداند.
func (s *Service) Get(id int) (*Response, error) {
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
	for _, o := range list {
		result = append(result, *toResponse(&o))
	}
	return result, nil
}

// Update پس از تایید هش یکپارچگی، سازمان را بروزرسانی و هش جدید تولید می‌کند.
func (s *Service) Update(id int, req UpdateRequest, actorID int, ip string) (*Response, error) {
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

	integrityHash, err := s.encryptSvc.CreateSecurityCode(encryption.OrganizationSensitiveData{
		Name:      o.Name,
		IsTakmili: o.IsTakmili,
		IsActive:  o.IsActive,
	})
	if err != nil {
		return nil, apperror.New("ENCRYPTION_ERROR", "خطا در ایجاد هش امنیتی سازمان.", err.Error(), 500)
	}
	o.IntegrityHash = integrityHash

	if err := s.repo.Update(o); err != nil {
		return nil, apperror.New("DB_ERROR", "خطا در بروزرسانی سازمان.", err.Error(), 500)
	}
	_ = s.audit.LogEvent(&actorID, ip, audit.EventUserDataChange, fmt.Sprintf("بروزرسانی سازمان %s", o.Name))
	return toResponse(o), nil
}

// Delete پس از تایید هش یکپارچگی، سازمان را به‌صورت soft-delete حذف می‌کند.
func (s *Service) Delete(id int, actorID int, ip string) error {
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
