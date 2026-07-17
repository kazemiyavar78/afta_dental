package patient

import (
	"fmt"
	"time"

	"github.com/tpdenta/afta-reception/internal/platform/apperror"
	"github.com/tpdenta/afta-reception/internal/platform/security/audit"
	"github.com/tpdenta/afta-reception/internal/platform/security/encryption"
	"gorm.io/gorm"
)

const birthDateLayout = "2006-01-02"

// Service لایه منطق کسب‌وکار بیمار.
type Service struct {
	repo       Repository
	audit      *audit.Manager
	encryptSvc *encryption.PatientEncryptionService
}

// NewService نمونه Service بیمار را می‌سازد.
func NewService(
	db *gorm.DB,
	auditMgr *audit.Manager,
	encryptSvc *encryption.PatientEncryptionService,
) *Service {
	return &Service{
		repo:       NewRepository(db),
		audit:      auditMgr,
		encryptSvc: encryptSvc,
	}
}

// derefString اشاره‌گر رشته را به مقدار خالی یا مقدار واقعی تبدیل می‌کند.
func derefString(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}

// parseBirthDate رشته تاریخ تولد را به time.Time تبدیل می‌کند.
func parseBirthDate(value string) (time.Time, error) {
	t, err := time.Parse(birthDateLayout, value)
	if err != nil {
		return time.Time{}, apperror.New("VALIDATION_ERROR", "فرمت تاریخ تولد نامعتبر است (YYYY-MM-DD).", err.Error(), 400)
	}
	return t, nil
}

// toSensitiveData مدل بیمار را به داده حساس رمزنگاری تبدیل می‌کند.
func toSensitiveData(p *Patient) encryption.PatientSensitiveData {
	return encryption.PatientSensitiveData{
		FirstName:         p.FirstName,
		LastName:          p.LastName,
		NationalCode:      p.NationalCode,
		BirthDate:         p.BirthDate.Format(birthDateLayout),
		Address:           derefString(p.Address),
		HomePhoneNumber:   derefString(p.HomePhoneNumber),
		MobilePhoneNumber: derefString(p.MobilePhoneNumber),
		FileNumber:        p.FileNumber,
		Sex:               p.Sex,
	}
}

// toResponse مدل دامنه را به DTO پاسخ تبدیل می‌کند.
func toResponse(p *Patient) *Response {
	return &Response{
		ID:                p.ID,
		FirstName:         p.FirstName,
		LastName:          p.LastName,
		NationalCode:      p.NationalCode,
		BirthDate:         p.BirthDate.Format(birthDateLayout),
		Address:           p.Address,
		HomePhoneNumber:   p.HomePhoneNumber,
		MobilePhoneNumber: p.MobilePhoneNumber,
		FileNumber:        p.FileNumber,
		Sex:               p.Sex,
	}
}

// verifyIntegrity هش یکپارچگی بیمار را با داده‌های فعلی بررسی می‌کند؛ در صورت نقض، رویداد دستکاری ثبت و خطا برمی‌گرداند.
func (s *Service) verifyIntegrity(p *Patient, actorID int, ip string) error {
	if !s.encryptSvc.CheckSecurityCode(toSensitiveData(p), p.IntegrityHash) {
		_ = s.audit.LogEvent(&actorID, ip, audit.EventDataTampering,
			fmt.Sprintf("نقض یکپارچگی بیمار %d (%s %s)", p.ID, p.FirstName, p.LastName))
		return apperror.ErrIntegrity
	}
	return nil
}

// Create بیمار جدید می‌سازد و هش امنیتی آن را تولید می‌کند.
func (s *Service) Create(req CreateRequest, actorID int, ip string) (*Response, error) {
	birthDate, err := parseBirthDate(req.BirthDate)
	if err != nil {
		return nil, err
	}

	p := &Patient{
		FirstName:         req.FirstName,
		LastName:          req.LastName,
		NationalCode:      req.NationalCode,
		BirthDate:         birthDate,
		Address:           req.Address,
		HomePhoneNumber:   req.HomePhoneNumber,
		MobilePhoneNumber: req.MobilePhoneNumber,
		FileNumber:        req.FileNumber,
		Sex:               req.Sex,
	}

	integrityHash, err := s.encryptSvc.CreateSecurityCode(toSensitiveData(p))
	if err != nil {
		return nil, apperror.New("ENCRYPTION_ERROR", "خطا در ایجاد هش امنیتی بیمار.", err.Error(), 500)
	}
	p.IntegrityHash = integrityHash

	if err := s.repo.Create(p); err != nil {
		return nil, apperror.New("DB_ERROR", "خطا در ایجاد بیمار.", err.Error(), 500)
	}

	created, err := s.repo.FindByID(p.ID)
	if err != nil {
		return toResponse(p), nil
	}
	_ = s.audit.LogEvent(&actorID, ip, audit.EventUserDataChange,
		fmt.Sprintf("ایجاد بیمار %s %s", p.FirstName, p.LastName))
	return toResponse(created), nil
}

// GetByNationalCode بیمار را با کد ملی دقیق برمی‌گرداند.
func (s *Service) GetByNationalCode(nationalCode string) (*Response, error) {
	p, err := s.repo.FindByNationalCode(nationalCode)
	if err == gorm.ErrRecordNotFound {
		return nil, apperror.ErrNotFound
	}
	if err != nil {
		return nil, apperror.New("DB_ERROR", "خطا در خواندن بیمار.", err.Error(), 500)
	}
	return toResponse(p), nil
}

// GetByFileNumber بیمار را با شماره پرونده دقیق برمی‌گرداند.
func (s *Service) GetByFileNumber(fileNumber string) (*Response, error) {
	p, err := s.repo.FindByFileNumber(fileNumber)
	if err == gorm.ErrRecordNotFound {
		return nil, apperror.ErrNotFound
	}
	if err != nil {
		return nil, apperror.New("DB_ERROR", "خطا در خواندن بیمار.", err.Error(), 500)
	}
	return toResponse(p), nil
}

// Get بیمار را با شناسه برمی‌گرداند.
func (s *Service) Get(id uint) (*Response, error) {
	p, err := s.repo.FindByID(id)
	if err == gorm.ErrRecordNotFound {
		return nil, apperror.ErrNotFound
	}
	if err != nil {
		return nil, apperror.New("DB_ERROR", "خطا در خواندن بیمار.", err.Error(), 500)
	}
	return toResponse(p), nil
}

// List لیست همه بیماران را برمی‌گرداند.
func (s *Service) List() ([]Response, error) {
	list, err := s.repo.FindAll()
	if err != nil {
		return nil, apperror.New("DB_ERROR", "خطا در خواندن بیماران.", err.Error(), 500)
	}
	result := make([]Response, 0, len(list))
	for i := range list {
		result = append(result, *toResponse(&list[i]))
	}
	return result, nil
}

// Search بیماران را بر اساس فیلترهای فیلدها برمی‌گرداند.
func (s *Service) Search(req SearchRequest) ([]Response, error) {
	list, err := s.repo.Search(SearchFilter{
		FirstName:         req.FirstName,
		LastName:          req.LastName,
		NationalCode:      req.NationalCode,
		BirthDate:         req.BirthDate,
		Address:           req.Address,
		HomePhoneNumber:   req.HomePhoneNumber,
		MobilePhoneNumber: req.MobilePhoneNumber,
		FileNumber:        req.FileNumber,
		Sex:               req.Sex,
	})
	if err != nil {
		return nil, apperror.New("DB_ERROR", "خطا در جستجوی بیماران.", err.Error(), 500)
	}
	result := make([]Response, 0, len(list))
	for i := range list {
		result = append(result, *toResponse(&list[i]))
	}
	return result, nil
}

// Update پس از تایید هش یکپارچگی، بیمار را بروزرسانی و هش جدید تولید می‌کند.
func (s *Service) Update(id uint, req UpdateRequest, actorID int, ip string) (*Response, error) {
	p, err := s.repo.FindByID(id)
	if err == gorm.ErrRecordNotFound {
		return nil, apperror.ErrNotFound
	}
	if err != nil {
		return nil, apperror.New("DB_ERROR", "خطا در خواندن بیمار.", err.Error(), 500)
	}

	if err := s.verifyIntegrity(p, actorID, ip); err != nil {
		return nil, err
	}

	birthDate, err := parseBirthDate(req.BirthDate)
	if err != nil {
		return nil, err
	}

	p.FirstName = req.FirstName
	p.LastName = req.LastName
	p.NationalCode = req.NationalCode
	p.BirthDate = birthDate
	p.Address = req.Address
	p.HomePhoneNumber = req.HomePhoneNumber
	p.MobilePhoneNumber = req.MobilePhoneNumber
	p.FileNumber = req.FileNumber
	p.Sex = req.Sex

	integrityHash, err := s.encryptSvc.CreateSecurityCode(toSensitiveData(p))
	if err != nil {
		return nil, apperror.New("ENCRYPTION_ERROR", "خطا در ایجاد هش امنیتی بیمار.", err.Error(), 500)
	}
	p.IntegrityHash = integrityHash

	if err := s.repo.Update(p); err != nil {
		return nil, apperror.New("DB_ERROR", "خطا در بروزرسانی بیمار.", err.Error(), 500)
	}

	updated, err := s.repo.FindByID(id)
	if err != nil {
		return toResponse(p), nil
	}
	_ = s.audit.LogEvent(&actorID, ip, audit.EventUserDataChange,
		fmt.Sprintf("بروزرسانی بیمار %s %s", p.FirstName, p.LastName))
	return toResponse(updated), nil
}

// Delete پس از تایید هش یکپارچگی، بیمار را به‌صورت soft-delete حذف می‌کند.
func (s *Service) Delete(id uint, actorID int, ip string) error {
	p, err := s.repo.FindByID(id)
	if err == gorm.ErrRecordNotFound {
		return apperror.ErrNotFound
	}
	if err != nil {
		return apperror.New("DB_ERROR", "خطا در خواندن بیمار.", err.Error(), 500)
	}

	if err := s.verifyIntegrity(p, actorID, ip); err != nil {
		return err
	}

	if err := s.repo.Delete(p); err != nil {
		return apperror.New("DB_ERROR", "خطا در حذف بیمار.", err.Error(), 500)
	}
	_ = s.audit.LogEvent(&actorID, ip, audit.EventUserDataChange,
		fmt.Sprintf("حذف بیمار %s %s", p.FirstName, p.LastName))
	return nil
}
