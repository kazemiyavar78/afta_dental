package regulation

import (
	"fmt"
	"strings"

	"github.com/tpdenta/afta-reception/internal/platform/apperror"
	"github.com/tpdenta/afta-reception/internal/platform/security/audit"
	"gorm.io/gorm"
)

// Service لایه منطق کسب‌وکار ضوابط.
type Service struct {
	repo  Repository
	audit *audit.Manager
}

// NewService نمونه Service ضوابط را می‌سازد.
func NewService(db *gorm.DB, auditMgr *audit.Manager) *Service {
	return &Service{repo: NewRepository(db), audit: auditMgr}
}

// toResponse مدل دامنه را به DTO پاسخ تبدیل می‌کند.
func toResponse(item *Regulation) *Response {
	ids := make([]uint, 0, len(item.Services))
	for _, s := range item.Services {
		ids = append(ids, s.ServiceID)
	}
	return &Response{
		ID:           item.ID,
		ServiceIDs:   ids,
		DurationDays: item.DurationDays,
		IsActive:     item.IsActive,
		PhotoCount:   item.PhotoCount,
		Description:  item.Description,
	}
}

// validateRequest اعتبارسنجی فیلدهای ضابطه را انجام می‌دهد.
func validateRequest(serviceIDs []uint, durationDays, photoCount int) error {
	if len(serviceIDs) == 0 {
		return apperror.New("VALIDATION_ERROR", "حداقل یک خدمت برای ضابطه الزامی است.", "services required", 400)
	}
	if durationDays <= 0 {
		return apperror.New("VALIDATION_ERROR", "مدت زمان ضابطه باید بزرگ‌تر از صفر باشد.", "invalid duration", 400)
	}
	if photoCount < 0 {
		return apperror.New("VALIDATION_ERROR", "تعداد عکس نمی‌تواند منفی باشد.", "invalid photo count", 400)
	}
	return nil
}

// Create ضابطه جدید می‌سازد.
func (s *Service) Create(req CreateRequest, actorID int, ip string) (*Response, error) {
	if err := validateRequest(req.ServiceIDs, req.DurationDays, req.PhotoCount); err != nil {
		return nil, err
	}
	photoCount := req.PhotoCount
	if photoCount == 0 {
		photoCount = 1
	}
	item := &Regulation{
		DurationDays: req.DurationDays,
		IsActive:     req.IsActive,
		PhotoCount:   photoCount,
		Description:  strings.TrimSpace(req.Description),
	}
	for _, sid := range req.ServiceIDs {
		item.Services = append(item.Services, RegulationService{ServiceID: sid})
	}
	if err := s.repo.Create(item); err != nil {
		return nil, apperror.New("DB_ERROR", "خطا در ایجاد ضابطه.", err.Error(), 500)
	}
	_ = s.audit.LogEvent(&actorID, ip, audit.EventUserDataChange, fmt.Sprintf("ایجاد ضابطه %d", item.ID))
	return toResponse(item), nil
}

// Get ضابطه را با شناسه برمی‌گرداند.
func (s *Service) Get(id uint) (*Response, error) {
	item, err := s.repo.FindByID(id)
	if err == gorm.ErrRecordNotFound {
		return nil, apperror.ErrNotFound
	}
	if err != nil {
		return nil, apperror.New("DB_ERROR", "خطا در خواندن ضابطه.", err.Error(), 500)
	}
	return toResponse(item), nil
}

// List لیست همه ضوابط را برمی‌گرداند.
func (s *Service) List() ([]Response, error) {
	list, err := s.repo.FindAll()
	if err != nil {
		return nil, apperror.New("DB_ERROR", "خطا در خواندن ضوابط.", err.Error(), 500)
	}
	out := make([]Response, 0, len(list))
	for i := range list {
		out = append(out, *toResponse(&list[i]))
	}
	return out, nil
}

// ListActive ضوابط فعال را برمی‌گرداند.
func (s *Service) ListActive() ([]Response, error) {
	list, err := s.repo.FindActive()
	if err != nil {
		return nil, apperror.New("DB_ERROR", "خطا در خواندن ضوابط فعال.", err.Error(), 500)
	}
	out := make([]Response, 0, len(list))
	for i := range list {
		out = append(out, *toResponse(&list[i]))
	}
	return out, nil
}

// Update ضابطه را بروزرسانی می‌کند.
func (s *Service) Update(id uint, req UpdateRequest, actorID int, ip string) (*Response, error) {
	item, err := s.repo.FindByID(id)
	if err == gorm.ErrRecordNotFound {
		return nil, apperror.ErrNotFound
	}
	if err != nil {
		return nil, apperror.New("DB_ERROR", "خطا در خواندن ضابطه.", err.Error(), 500)
	}
	if err := validateRequest(req.ServiceIDs, req.DurationDays, req.PhotoCount); err != nil {
		return nil, err
	}
	photoCount := req.PhotoCount
	if photoCount == 0 {
		photoCount = 1
	}
	item.DurationDays = req.DurationDays
	item.IsActive = req.IsActive
	item.PhotoCount = photoCount
	item.Description = strings.TrimSpace(req.Description)
	if err := s.repo.Update(item); err != nil {
		return nil, apperror.New("DB_ERROR", "خطا در بروزرسانی ضابطه.", err.Error(), 500)
	}
	if err := s.repo.ReplaceServices(id, req.ServiceIDs); err != nil {
		return nil, apperror.New("DB_ERROR", "خطا در بروزرسانی خدمات ضابطه.", err.Error(), 500)
	}
	updated, err := s.repo.FindByID(id)
	if err != nil {
		return nil, apperror.New("DB_ERROR", "خطا در خواندن ضابطه.", err.Error(), 500)
	}
	_ = s.audit.LogEvent(&actorID, ip, audit.EventUserDataChange, fmt.Sprintf("بروزرسانی ضابطه %d", id))
	return toResponse(updated), nil
}

// Delete ضابطه را حذف می‌کند.
func (s *Service) Delete(id uint, actorID int, ip string) error {
	item, err := s.repo.FindByID(id)
	if err == gorm.ErrRecordNotFound {
		return apperror.ErrNotFound
	}
	if err != nil {
		return apperror.New("DB_ERROR", "خطا در خواندن ضابطه.", err.Error(), 500)
	}
	if err := s.repo.Delete(item); err != nil {
		return apperror.New("DB_ERROR", "خطا در حذف ضابطه.", err.Error(), 500)
	}
	_ = s.audit.LogEvent(&actorID, ip, audit.EventUserDataChange, fmt.Sprintf("حذف ضابطه %d", id))
	return nil
}
