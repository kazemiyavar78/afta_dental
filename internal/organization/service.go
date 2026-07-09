// این ماژول نمونه است؛ باقی Entity ها و Endpoint ها طبق همین الگو در فازهای بعدی تکمیل می‌شوند.
package organization

import (
	"fmt"
	"time"

	"github.com/tpdenta/afta-reception/internal/platform/apperror"
	"github.com/tpdenta/afta-reception/internal/platform/security/audit"
	"gorm.io/gorm"
)

type Service struct {
	repo  Repository
	audit *audit.Manager
}

func NewService(db *gorm.DB, auditMgr *audit.Manager) *Service {
	return &Service{repo: NewRepository(db), audit: auditMgr}
}

func (s *Service) Create(req CreateRequest, actorID int, ip string) (*Response, error) {
	o := &Organization{Name: req.Name, CreatedAt: time.Now().UTC()}
	if err := s.repo.Create(o); err != nil {
		return nil, apperror.New("DB_ERROR", "خطا در ایجاد سازمان.", err.Error(), 500)
	}
	_ = s.audit.LogEvent(&actorID, ip, audit.EventUserDataChange, fmt.Sprintf("ایجاد سازمان %s", o.Name))
	return &Response{ID: o.ID, Name: o.Name}, nil
}

func (s *Service) Get(id int) (*Response, error) {
	o, err := s.repo.FindByID(id)
	if err == gorm.ErrRecordNotFound {
		return nil, apperror.ErrNotFound
	}
	if err != nil {
		return nil, apperror.New("DB_ERROR", "خطا در خواندن سازمان.", err.Error(), 500)
	}
	return &Response{ID: o.ID, Name: o.Name}, nil
}

func (s *Service) List() ([]Response, error) {
	list, err := s.repo.FindAll()
	if err != nil {
		return nil, apperror.New("DB_ERROR", "خطا در خواندن سازمان‌ها.", err.Error(), 500)
	}
	var result []Response
	for _, o := range list {
		result = append(result, Response{ID: o.ID, Name: o.Name})
	}
	return result, nil
}
