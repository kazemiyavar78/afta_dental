// این ماژول نمونه است؛ باقی Entity ها و Endpoint ها طبق همین الگو در فازهای بعدی تکمیل می‌شوند.
package fund

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
	f := &Fund{Name: req.Name, Balance: 0, CreatedAt: time.Now().UTC()}
	if err := s.repo.Create(f); err != nil {
		return nil, apperror.New("DB_ERROR", "خطا در ایجاد صندوق.", err.Error(), 500)
	}
	_ = s.audit.LogEvent(&actorID, ip, audit.EventUserDataChange, fmt.Sprintf("ایجاد صندوق %s", f.Name))
	return &Response{ID: f.ID, Name: f.Name, Balance: f.Balance}, nil
}

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
