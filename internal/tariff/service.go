// این ماژول نمونه است؛ باقی Entity ها و Endpoint ها طبق همین الگو در فازهای بعدی تکمیل می‌شوند.
package tariff

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
	t := &Tariff{Name: req.Name, Amount: req.Amount, CreatedAt: time.Now().UTC()}
	if err := s.repo.Create(t); err != nil {
		return nil, apperror.New("DB_ERROR", "خطا در ایجاد تعرفه.", err.Error(), 500)
	}
	_ = s.audit.LogEvent(&actorID, ip, audit.EventUserDataChange, fmt.Sprintf("ایجاد تعرفه %s", t.Name))
	return &Response{ID: t.ID, Name: t.Name, Amount: t.Amount}, nil
}

func (s *Service) Get(id int) (*Response, error) {
	t, err := s.repo.FindByID(id)
	if err == gorm.ErrRecordNotFound {
		return nil, apperror.ErrNotFound
	}
	if err != nil {
		return nil, apperror.New("DB_ERROR", "خطا در خواندن تعرفه.", err.Error(), 500)
	}
	return &Response{ID: t.ID, Name: t.Name, Amount: t.Amount}, nil
}

func (s *Service) List() ([]Response, error) {
	list, err := s.repo.FindAll()
	if err != nil {
		return nil, apperror.New("DB_ERROR", "خطا در خواندن تعرفه‌ها.", err.Error(), 500)
	}
	var result []Response
	for _, t := range list {
		result = append(result, Response{ID: t.ID, Name: t.Name, Amount: t.Amount})
	}
	return result, nil
}
