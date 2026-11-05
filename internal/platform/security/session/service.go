package session

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/tpdenta/afta-reception/internal/platform/apperror"
	"github.com/tpdenta/afta-reception/internal/platform/security/audit"
	"github.com/tpdenta/afta-reception/internal/platform/security/settings"
	"gorm.io/gorm"
)

const SessionCookieName = "session_id"

// intSetting فقط سقف نشست را می‌خواند تا تست بدون سرویس کامل تنظیمات بماند.
type intSetting interface {
	GetSettingInt(name string) (int, error)
}

// Service سرویس مدیریت نشست‌ها.
type Service struct {
	repo     Repository
	settings intSetting
	audit    audit.Recorder
	namer    func(userID int) (string, error)
}

// NewService نمونه Service می‌سازد.
// db اتصال پایگاه، settingsSvc سقف نشست و auditMgr ثبت ممیزی است.
// نام صاحب نشست بعداً با SetAccountNamer وصل می‌شود تا چرخه وابستگی با کاربر پیش نیاید.
func NewService(db *gorm.DB, settingsSvc *settings.Service, auditMgr *audit.Manager) *Service {
	return &Service{
		repo:     NewRepository(db),
		settings: settingsSvc,
		audit:    auditMgr,
	}
}

// SetAccountNamer تابع نام کاربری را برای هدف رویداد نشست ذخیره می‌کند.
// fn شناسه کاربر را می‌گیرد و نام را برمی‌گرداند. خروجی ندارد.
// بعد از ساخته شدن سرویس کاربر صدا زده می‌شود.
func (s *Service) SetAccountNamer(fn func(userID int) (string, error)) {
	s.namer = fn
}

// CreateSession نشست جدید می‌سازد و اگر از سقف گذشته باشد قدیمی‌ترین‌ها را با دلیل max_sessions می‌بندد.
// ctx لغو درخواست است. userID صاحب، ip و browser مبدأ نشست تازه‌اند.
// برای هر نشست حذف‌شده فاعل سیستم است. خروجی نشست جدید یا خطا است.
// FTA_MCS / R1-06: رسیدن به سقف بدون رویداد ممیزی نمی‌ماند.
func (s *Service) CreateSession(ctx context.Context, userID int, ip, browser string) (*Session, error) {
	maxSessions, err := s.settings.GetSettingInt(settings.MaximumNumberOfSessionsPerUser)
	if err != nil {
		maxSessions = 3
	}

	count, err := s.repo.CountByUserID(userID)
	if err != nil {
		return nil, apperror.New("DB_ERROR", "خطا در شمارش نشست‌ها.", err.Error(), 500)
	}

	for count >= int64(maxSessions) {
		oldest, err := s.repo.DeleteOldestByUserID(userID)
		if err != nil {
			return nil, apperror.New("DB_ERROR", "خطا در حذف نشست قدیمی.", err.Error(), 500)
		}
		if oldest == nil {
			break
		}
		if err := s.recordSessionEnd(ctx, audit.SystemActor(), oldest, audit.ReasonMaxSessions); err != nil {
			return nil, err
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
// userID صاحب است. خروجی فهرست یا خطای خواندن است.
func (s *Service) GetSessionsByUserID(userID int) ([]Session, error) {
	sessions, err := s.repo.FindByUserID(userID)
	if err != nil {
		return nil, apperror.New("DB_ERROR", "خطا در خواندن سشن ها.", err.Error(), 500)
	}
	return sessions, nil
}

// GetSession نشست را با شناسه برمی‌گرداند.
// id شناسه کوکی است. اگر نباشد ErrUnauthorized برمی‌گردد.
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

// DeleteSession نشست را حذف می‌کند و دو رویداد پایان و خروج می‌نویسد.
// ctx لغو درخواست است. id شناسه نشست است. actor کسی است که حذف را زده.
// canManageOthers اجازه بستن نشست دیگران است. دلیل، user_logout یا admin_terminated است.
// خطا یعنی حذف یا ممیزی انجام نشده و پاسخ باید به کاربر برگردد.
func (s *Service) DeleteSession(ctx context.Context, id uuid.UUID, actor audit.Actor, canManageOthers bool) error {
	sess, err := s.repo.FindByID(id)
	if err == gorm.ErrRecordNotFound {
		return apperror.ErrNotFound
	}
	if err != nil {
		return apperror.New("DB_ERROR", "خطا در خواندن نشست.", err.Error(), 500)
	}
	if !canManageOthers && (actor.UserID == nil || sess.PersonnelAccountID != *actor.UserID) {
		return apperror.ErrForbidden
	}
	reason := closeReason(actor, sess.PersonnelAccountID)
	if err := s.recordSessionEnd(ctx, actor, sess, reason); err != nil {
		return err
	}
	if err := s.repo.Delete(id); err != nil {
		return apperror.New("DB_ERROR", "خطا در حذف نشست.", err.Error(), 500)
	}
	return nil
}

// ListSessions لیست نشست‌ها را برمی‌گرداند (با canListAll همه، وگرنه فقط خودش).
// actorUserID درخواست‌کننده است. خطا خطای خواندن پایگاه است.
func (s *Service) ListSessions(actorUserID int, canListAll bool) ([]Session, error) {
	if canListAll {
		return s.repo.FindAll()
	}
	return s.repo.FindByUserID(actorUserID)
}

// DeleteAllUserSessions همه نشست‌های یک کاربر را با یک دلیل می‌بندد و برای هر کدام رویداد می‌سازد.
// ctx لغو درخواست است. userID صاحب است. actor کسی است که بستن را انجام داده یا سیستم است.
// reason کد دلیل مشترک همه آن نشست‌هاست، مثل security_code_failed.
// اگر ممیزی شکست بخورد خطا برمی‌گردد؛ نشست‌ها قبل از بازگشت حذف شده‌اند تا نشست نامعتبر نماند.
func (s *Service) DeleteAllUserSessions(ctx context.Context, userID int, actor audit.Actor, reason string) error {
	sessions, err := s.repo.FindByUserID(userID)
	if err != nil {
		return apperror.New("DB_ERROR", "خطا در خواندن نشست‌ها.", err.Error(), 500)
	}
	if err := s.repo.DeleteByUserID(userID); err != nil {
		return apperror.New("DB_ERROR", "خطا در حذف نشست‌ها.", err.Error(), 500)
	}
	for i := range sessions {
		if err := s.recordSessionEnd(ctx, actor, &sessions[i], reason); err != nil {
			return err
		}
	}
	return nil
}
