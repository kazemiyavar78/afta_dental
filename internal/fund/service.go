package fund

import (
	"context"
	"strconv"
	"time"

	"github.com/tpdenta/afta-reception/internal/platform/apperror"
	"github.com/tpdenta/afta-reception/internal/platform/security/audit"
	"gorm.io/gorm"
)

// Service منطق ایجاد و خواندن صندوق‌های نام‌دار است.
type Service struct {
	repo  Repository
	audit audit.Recorder
}

// NewService سرویس صندوق را با مخزن و ثبت‌کننده ممیزی می‌سازد.
// db اتصال پایگاه است و auditMgr ثبت مرکزی است. خروجی سرویس آماده هندلر است.
func NewService(db *gorm.DB, auditMgr *audit.Manager) *Service {
	return &Service{repo: NewRepository(db), audit: auditMgr}
}

// Create صندوق جدید با مانده صفر می‌سازد و ENTITY_CREATED ثبت می‌کند.
// ctx لغو درخواست است. req نام صندوق و actor سازنده است. خروجی پاسخ یا خطا است.
// شکست ثبت ممیزی عملیات را ناموفق می‌کند.
func (s *Service) Create(ctx context.Context, req CreateRequest, actor audit.Actor) (*Response, error) {
	f := &Fund{Name: req.Name, Balance: 0, CreatedAt: time.Now().UTC()}
	if err := s.repo.Create(f); err != nil {
		return nil, apperror.New("DB_ERROR", "خطا در ایجاد صندوق.", err.Error(), 500)
	}
	if err := s.record(ctx, audit.Event{
		Type:        audit.EventEntityCreated,
		Result:      audit.ResultSuccess,
		Actor:       actor,
		Target:      audit.Target{Type: audit.TargetFund, ID: strconv.Itoa(f.ID), Name: f.Name},
		Description: "ایجاد صندوق «" + f.Name + "»",
	}); err != nil {
		return nil, err
	}
	return &Response{ID: f.ID, Name: f.Name, Balance: f.Balance}, nil
}

// Get صندوق را با شناسه برمی‌گرداند.
// id شناسه صندوق است. خروجی پاسخ یا خطای نبودن است.
func (s *Service) Get(id int) (*Response, error) {
	f, err := s.repo.FindByID(id)
	if err == gorm.ErrRecordNotFound {
		return nil, apperror.ErrNotFound
	}
	if err != nil {
		return nil, apperror.New("DB_ERROR", "خطا در خواندن صندوق.", err.Error(), 500)
	}
	return &Response{ID: f.ID, Name: f.Name, Balance: f.Balance}, nil
}

// List همه صندوق‌ها را برمی‌گرداند.
// خروجی فهرست پاسخ یا خطای خواندن است.
func (s *Service) List() ([]Response, error) {
	list, err := s.repo.FindAll()
	if err != nil {
		return nil, apperror.New("DB_ERROR", "خطا در خواندن صندوق‌ها.", err.Error(), 500)
	}
	var result []Response
	for _, f := range list {
		result = append(result, Response{ID: f.ID, Name: f.Name, Balance: f.Balance})
	}
	return result, nil
}

// record رویداد را می‌نویسد و شکست را به AUDIT_UNAVAILABLE تبدیل می‌کند.
// ctx زمینه است. e رویداد است. nil یعنی ثبت شد.
func (s *Service) record(ctx context.Context, e audit.Event) error {
	if s.audit == nil {
		return audit.Unavailable(apperror.ErrInternal)
	}
	if ctx == nil {
		ctx = context.Background()
	}
	return audit.Unavailable(s.audit.Record(ctx, e))
}
