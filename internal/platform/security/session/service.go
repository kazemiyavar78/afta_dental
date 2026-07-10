package session

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/tpdenta/afta-reception/internal/platform/apperror"
	"github.com/tpdenta/afta-reception/internal/platform/security/audit"
	"github.com/tpdenta/afta-reception/internal/platform/security/settings"
	"gorm.io/gorm"
)

const SessionCookieName = "session_id"

// Service سرویس مدیریت نشست‌ها.
type Service struct {
	repo     Repository
	settings *settings.Service
	audit    *audit.Manager
}

// NewService نمونه Service می‌سازد.
func NewService(db *gorm.DB, settingsSvc *settings.Service, auditMgr *audit.Manager) *Service {
	return &Service{
		repo:     NewRepository(db),
		settings: settingsSvc,
		audit:    auditMgr,
	}
}

// CreateSession نشست جدید می‌سازد و محدودیت تعداد هم‌زمان را اعمال می‌کند.
func (s *Service) CreateSession(userID int, ip, browser string) (*Session, error) {
	maxSessions, err := s.settings.GetSettingInt(settings.MaximumNumberOfSessionsPerUser)
	if err != nil {
		maxSessions = 3
	}

	count, err := s.repo.CountByUserID(userID)
	if err != nil {
		return nil, apperror.New("DB_ERROR", "خطا در شمارش نشست‌ها.", err.Error(), 500)
	}

	for count >= int64(maxSessions) {
		if err := s.repo.DeleteOldestByUserID(userID); err != nil {
			return nil, apperror.New("DB_ERROR", "خطا در حذف نشست قدیمی.", err.Error(), 500)
		}
		count--
	}

	sess := &Session{
		ID:                 NewMSSQLUUID(),
		Ip:                 ip,
		Browser:            browser,
		CreationTime:       time.Now().UTC(),
		PersonnelAccountID: userID,
	}

	if err := s.repo.Create(sess); err != nil {
		return nil, apperror.New("DB_ERROR", "خطا در ایجاد نشست.", err.Error(), 500)
	}

	return sess, nil
}

// GetSessionsByUserID نشست‌های یک کاربر را برمی‌گرداند.
func (s *Service) GetSessionsByUserID(userID int) ([]Session, error) {
	sessions, err := s.repo.FindByUserID(userID)
	if err != nil {
		return nil, apperror.New("DB_ERROR", "خطا در خواندن سشن ها.", err.Error(), 500)
	}
	return sessions, nil
}

// GetSession نشست را با شناسه برمی‌گرداند.
func (s *Service) GetSession(id uuid.UUID) (*Session, error) {
	sess, err := s.repo.FindByID(id)
	if err == gorm.ErrRecordNotFound {
		return nil, apperror.ErrUnauthorized
	}
	if err != nil {
		return nil, apperror.New("DB_ERROR", "خطا در خواندن نشست.", err.Error(), 500)
	}
	return sess, nil
}

// DeleteSession نشست را فیزیکی حذف می‌کند.
func (s *Service) DeleteSession(id uuid.UUID, actorUserID int, actorRole string, ip string) error {
	sess, err := s.repo.FindByID(id)
	if err == gorm.ErrRecordNotFound {
		return apperror.ErrNotFound
	}
	if err != nil {
		return apperror.New("DB_ERROR", "خطا در خواندن نشست.", err.Error(), 500)
	}

	if actorRole != "Admin" && sess.PersonnelAccountID != actorUserID {
		return apperror.ErrForbidden
	}

	if err := s.repo.Delete(id); err != nil {
		return apperror.New("DB_ERROR", "خطا در حذف نشست.", err.Error(), 500)
	}

	uid := actorUserID
	_ = s.audit.LogEvent(&uid, ip, audit.EventSessionRevoked,
		fmt.Sprintf("حذف نشست %s", id.String()))

	return nil
}

// ListSessions لیست نشست‌ها را برمی‌گرداند (Admin همه، کاربر فقط خودش).
func (s *Service) ListSessions(actorUserID int, actorRole string) ([]Session, error) {
	if actorRole == "Admin" {
		return s.repo.FindAll()
	}
	return s.repo.FindByUserID(actorUserID)
}

// DeleteAllUserSessions تمام نشست‌های یک کاربر را حذف می‌کند.
func (s *Service) DeleteAllUserSessions(userID int) error {
	return s.repo.DeleteByUserID(userID)
}
