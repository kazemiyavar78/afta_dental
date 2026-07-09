// این ماژول نمونه است؛ باقی Entity ها و Endpoint ها طبق همین الگو در فازهای بعدی تکمیل می‌شوند.
package reception

import (
	"fmt"
	"time"

	"github.com/tpdenta/afta-reception/internal/platform/apperror"
	"github.com/tpdenta/afta-reception/internal/platform/security/audit"
	"gorm.io/gorm"
)

// Service سرویس منطق کسب‌وکار پذیرش.
type Service struct {
	repo  Repository
	audit *audit.Manager
}

// NewService نمونه Service می‌سازد.
func NewService(db *gorm.DB, auditMgr *audit.Manager) *Service {
	return &Service{
		repo:  NewRepository(db),
		audit: auditMgr,
	}
}

// CreateReception پذیرش جدید ایجاد می‌کند.
func (s *Service) CreateReception(req CreateReceptionRequest, actorUserID int, ip string) (*ReceptionResponse, error) {
	date, err := time.Parse("2006-01-02", req.ReceptionDate)
	if err != nil {
		return nil, apperror.New("INVALID_DATE", "فرمت تاریخ نامعتبر است.", err.Error(), 400)
	}

	rec := &Reception{
		PatientName:   req.PatientName,
		DoctorID:      req.DoctorID,
		ReceptionDate: date,
		Status:        "pending",
		CreatedAt:     time.Now().UTC(),
	}

	if err := s.repo.Create(rec); err != nil {
		return nil, apperror.New("DB_ERROR", "خطا در ایجاد پذیرش.", err.Error(), 500)
	}

	_ = s.audit.LogEvent(&actorUserID, ip, audit.EventUserDataChange,
		fmt.Sprintf("ایجاد پذیرش بیمار %s", rec.PatientName))

	return s.toResponse(rec), nil
}

// GetReception پذیرش را با شناسه برمی‌گرداند.
func (s *Service) GetReception(id int) (*ReceptionResponse, error) {
	rec, err := s.repo.FindByID(id)
	if err == gorm.ErrRecordNotFound {
		return nil, apperror.ErrNotFound
	}
	if err != nil {
		return nil, apperror.New("DB_ERROR", "خطا در خواندن پذیرش.", err.Error(), 500)
	}
	return s.toResponse(rec), nil
}

// ListReceptions لیست پذیرش‌ها.
func (s *Service) ListReceptions() ([]ReceptionResponse, error) {
	list, err := s.repo.FindAll()
	if err != nil {
		return nil, apperror.New("DB_ERROR", "خطا در خواندن پذیرش‌ها.", err.Error(), 500)
	}
	var result []ReceptionResponse
	for _, r := range list {
		result = append(result, *s.toResponse(&r))
	}
	return result, nil
}

func (s *Service) toResponse(rec *Reception) *ReceptionResponse {
	return &ReceptionResponse{
		ID:            rec.ID,
		PatientName:   rec.PatientName,
		DoctorID:      rec.DoctorID,
		ReceptionDate: rec.ReceptionDate.Format("2006-01-02"),
		Status:        rec.Status,
	}
}
