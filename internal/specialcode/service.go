package specialcode

import (
	"fmt"
	"strings"

	"github.com/tpdenta/afta-reception/internal/platform/apperror"
	"github.com/tpdenta/afta-reception/internal/platform/security/audit"
	"github.com/tpdenta/afta-reception/internal/platform/security/integrity"
	"gorm.io/gorm"
)

// Service لایه منطق کسب‌وکار کد خاص.
type Service struct {
	repo   Repository
	audit  *audit.Manager
	signer *integrity.Signer
}

// NewService نمونه Service کد خاص را می‌سازد.
func NewService(db *gorm.DB, auditMgr *audit.Manager, signer *integrity.Signer) *Service {
	return &Service{repo: NewRepository(db), audit: auditMgr, signer: signer}
}

// toResponse مدل دامنه را به DTO پاسخ تبدیل می‌کند.
func toResponse(item *SpecialCode) *Response {
	return &Response{
		ID:          item.ID,
		Code:        item.Code,
		Name:        item.Name,
		Description: item.Description,
		Percentage:  item.Percentage,
		IsActive:    item.IsActive,
	}
}

// verifyIntegrity هش یکپارچگی را بررسی می‌کند؛ در صورت نقض خطا برمی‌گرداند.
func (s *Service) verifyIntegrity(item *SpecialCode, actorID int, ip string) error {
	if !VerifyIntegrity(s.signer, item) {
		_ = s.audit.LogEvent(&actorID, ip, audit.EventDataTampering,
			fmt.Sprintf("نقض یکپارچگی کد خاص %d (%s)", item.ID, item.Code))
		return apperror.ErrIntegrity
	}
	return nil
}

// Create کد خاص جدید می‌سازد و هش یکپارچگی آن را تولید می‌کند.
func (s *Service) Create(req CreateRequest, actorID int, ip string) (*Response, error) {
	code := strings.TrimSpace(req.Code)
	if code == "" || code == "0" {
		return nil, apperror.New("VALIDATION_ERROR", "کد خاص نامعتبر است.", "invalid special code", 400)
	}
	if req.Percentage > 100 {
		return nil, apperror.New("VALIDATION_ERROR", "درصد کد خاص باید بین ۰ تا ۱۰۰ باشد.", "invalid percentage", 400)
	}

	item := &SpecialCode{
		Code:        code,
		Name:        strings.TrimSpace(req.Name),
		Description: strings.TrimSpace(req.Description),
		Percentage:  req.Percentage,
		IsActive:    req.IsActive,
	}
	item.IntegrityHash = SignIntegrityHash(s.signer, item)

	if err := s.repo.Create(item); err != nil {
		return nil, apperror.New("DB_ERROR", "خطا در ایجاد کد خاص.", err.Error(), 500)
	}
	item.IntegrityHash = SignIntegrityHash(s.signer, item)
	if err := s.repo.Update(item); err != nil {
		return nil, apperror.New("DB_ERROR", "خطا در ذخیره هش کد خاص.", err.Error(), 500)
	}

	_ = s.audit.LogEvent(&actorID, ip, audit.EventUserDataChange, fmt.Sprintf("ایجاد کد خاص %s", item.Code))
	return toResponse(item), nil
}

// Get کد خاص را با شناسه برمی‌گرداند.
func (s *Service) Get(id uint) (*Response, error) {
	item, err := s.repo.FindByID(id)
	if err == gorm.ErrRecordNotFound {
		return nil, apperror.ErrNotFound
	}
	if err != nil {
		return nil, apperror.New("DB_ERROR", "خطا در خواندن کد خاص.", err.Error(), 500)
	}
	return toResponse(item), nil
}

// GetByCode کد خاص را با کد برمی‌گرداند و یکپارچگی را بررسی می‌کند.
func (s *Service) GetByCode(code string, actorID int, ip string) (*Response, error) {
	code = strings.TrimSpace(code)
	if code == "" || code == "0" {
		return nil, apperror.ErrNotFound
	}
	item, err := s.repo.FindByCode(code)
	if err == gorm.ErrRecordNotFound {
		return nil, apperror.ErrNotFound
	}
	if err != nil {
		return nil, apperror.New("DB_ERROR", "خطا در خواندن کد خاص.", err.Error(), 500)
	}
	if err := s.verifyIntegrity(item, actorID, ip); err != nil {
		return nil, err
	}
	return toResponse(item), nil
}

// GetActiveForCalc کد خاص فعال را برای محاسبه برمی‌گرداند؛ در صورت نقض هش یا غیرفعال بودن خطا می‌دهد.
func (s *Service) GetActiveForCalc(id uint, actorID int, ip string) (*SpecialCode, error) {
	item, err := s.repo.FindByID(id)
	if err == gorm.ErrRecordNotFound {
		return nil, apperror.New("E-014", "کد خاص یافت نشد.", "special code not found", 404)
	}
	if err != nil {
		return nil, apperror.New("DB_ERROR", "خطا در خواندن کد خاص.", err.Error(), 500)
	}
	if err := s.verifyIntegrity(item, actorID, ip); err != nil {
		return nil, err
	}
	if !item.IsActive || item.Percentage == 0 {
		return nil, nil
	}
	return item, nil
}

// List لیست همه کدهای خاص را برمی‌گرداند.
func (s *Service) List() ([]Response, error) {
	list, err := s.repo.FindAll()
	if err != nil {
		return nil, apperror.New("DB_ERROR", "خطا در خواندن کدهای خاص.", err.Error(), 500)
	}
	out := make([]Response, 0, len(list))
	for i := range list {
		out = append(out, *toResponse(&list[i]))
	}
	return out, nil
}

// Update کد خاص را بروزرسانی می‌کند.
func (s *Service) Update(id uint, req UpdateRequest, actorID int, ip string) (*Response, error) {
	item, err := s.repo.FindByID(id)
	if err == gorm.ErrRecordNotFound {
		return nil, apperror.ErrNotFound
	}
	if err != nil {
		return nil, apperror.New("DB_ERROR", "خطا در خواندن کد خاص.", err.Error(), 500)
	}
	if err := s.verifyIntegrity(item, actorID, ip); err != nil {
		return nil, err
	}

	code := strings.TrimSpace(req.Code)
	if code == "" || code == "0" {
		return nil, apperror.New("VALIDATION_ERROR", "کد خاص نامعتبر است.", "invalid special code", 400)
	}
	if req.Percentage > 100 {
		return nil, apperror.New("VALIDATION_ERROR", "درصد کد خاص باید بین ۰ تا ۱۰۰ باشد.", "invalid percentage", 400)
	}

	item.Code = code
	item.Name = strings.TrimSpace(req.Name)
	item.Description = strings.TrimSpace(req.Description)
	item.Percentage = req.Percentage
	item.IsActive = req.IsActive
	item.IntegrityHash = SignIntegrityHash(s.signer, item)

	if err := s.repo.Update(item); err != nil {
		return nil, apperror.New("DB_ERROR", "خطا در بروزرسانی کد خاص.", err.Error(), 500)
	}
	_ = s.audit.LogEvent(&actorID, ip, audit.EventUserDataChange, fmt.Sprintf("بروزرسانی کد خاص %s", item.Code))
	return toResponse(item), nil
}

// Delete کد خاص را حذف می‌کند.
func (s *Service) Delete(id uint, actorID int, ip string) error {
	item, err := s.repo.FindByID(id)
	if err == gorm.ErrRecordNotFound {
		return apperror.ErrNotFound
	}
	if err != nil {
		return apperror.New("DB_ERROR", "خطا در خواندن کد خاص.", err.Error(), 500)
	}
	if err := s.repo.Delete(item); err != nil {
		return apperror.New("DB_ERROR", "خطا در حذف کد خاص.", err.Error(), 500)
	}
	_ = s.audit.LogEvent(&actorID, ip, audit.EventUserDataChange, fmt.Sprintf("حذف کد خاص %s", item.Code))
	return nil
}
