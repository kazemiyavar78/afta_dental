package loginguard

import (
	"fmt"
	"time"

	"github.com/tpdenta/afta-reception/internal/platform/apperror"
	"github.com/tpdenta/afta-reception/internal/platform/security/settings"
	"gorm.io/gorm"
)

// LoginAttempt مدل جدول تلاش‌های ورود.
type LoginAttempt struct {
	Ip            string     `gorm:"column:Ip;primaryKey;size:45"`
	FailedCount   int        `gorm:"column:FailedCount;not null"`
	LastAttemptAt time.Time  `gorm:"column:LastAttemptAt;not null"`
	LockedUntil   *time.Time `gorm:"column:LockedUntil"`
}

// TableName نام جدول در دیتابیس.
func (LoginAttempt) TableName() string {
	return "LoginAttempts"
}

// Repository عملیات اتمیک روی تلاش‌های ورود.
type Repository interface {
	RecordFailedAttempt(ip string, maxFailed, lockMinutes int) (*LoginAttempt, error)
	ResetAttempts(ip string) error
	IsIPLocked(ip string) (bool, error)
	CleanupStale(olderThan time.Duration) error
}

type gormRepository struct {
	db *gorm.DB
}

// NewRepository نمونه Repository می‌سازد.
func NewRepository(db *gorm.DB) Repository {
	return &gormRepository{db: db}
}

// RecordFailedAttempt با MERGE اتمیک شمارنده را افزایش می‌دهد.
func (r *gormRepository) RecordFailedAttempt(ip string, maxFailed, lockMinutes int) (*LoginAttempt, error) {
	var attempt LoginAttempt
	now := time.Now().UTC()

	err := r.db.Transaction(func(tx *gorm.DB) error {
		result := tx.Exec(`
		MERGE LoginAttempts WITH (HOLDLOCK) AS target
		USING (SELECT ? AS Ip) AS source
		ON target.Ip = source.Ip
		WHEN MATCHED THEN
			UPDATE SET
				FailedCount = target.FailedCount + 1,
				LastAttemptAt = ?,
				LockedUntil = CASE
					WHEN target.FailedCount + 1 >= ? THEN DATEADD(MINUTE, ?, ?)
					ELSE target.LockedUntil
				END
		WHEN NOT MATCHED THEN
			INSERT (Ip, FailedCount, LastAttemptAt, LockedUntil)
			VALUES (?, 1, ?, CASE WHEN 1 >= ? THEN DATEADD(MINUTE, ?, ?) ELSE NULL END);
	`, ip, now, maxFailed, lockMinutes, now, ip, now, maxFailed, lockMinutes, now)

		if result.Error != nil {
			return result.Error
		}

		return tx.Where("Ip = ?", ip).First(&attempt).Error
	})

	if err != nil {
		return nil, err
	}
	return &attempt, nil
}

func (r *gormRepository) ResetAttempts(ip string) error {
	return r.db.Where("Ip = ?", ip).Delete(&LoginAttempt{}).Error
}

func (r *gormRepository) IsIPLocked(ip string) (bool, error) {
	var attempt LoginAttempt
	err := r.db.Where("Ip = ?", ip).First(&attempt).Error
	if err == gorm.ErrRecordNotFound {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	if attempt.LockedUntil != nil && attempt.LockedUntil.After(time.Now().UTC()) {
		return true, nil
	}
	return false, nil
}

func (r *gormRepository) CleanupStale(olderThan time.Duration) error {
	cutoff := time.Now().UTC().Add(-olderThan)
	return r.db.Where("LastAttemptAt < ? AND (LockedUntil IS NULL OR LockedUntil < ?)", cutoff, time.Now().UTC()).
		Delete(&LoginAttempt{}).Error
}

// Guard سرویس محدودیت ورود و ساعات کاری.
type Guard struct {
	repo     Repository
	settings *settings.Service
}

// NewGuard نمونه Guard می‌سازد.
func NewGuard(db *gorm.DB, settingsSvc *settings.Service) *Guard {
	return &Guard{
		repo:     NewRepository(db),
		settings: settingsSvc,
	}
}

// CheckIPLock بررسی قفل بودن IP.
func (g *Guard) CheckIPLock(ip string) error {
	locked, err := g.repo.IsIPLocked(ip)
	if err != nil {
		return apperror.New("DB_ERROR", "خطا در بررسی قفل IP.", err.Error(), 500)
	}
	if locked {
		return apperror.ErrIPLocked
	}
	return nil
}

// RecordFailedLogin تلاش ناموفق را ثبت می‌کند.
func (g *Guard) RecordFailedLogin(ip string) error {
	maxFailed, _ := g.settings.GetSettingInt(settings.MaximumNumberOfFailedLogin)
	if maxFailed <= 0 {
		maxFailed = 5
	}
	lockMinutes, _ := g.settings.GetSettingInt(settings.MaximumTimeOfUserIpBeingLocked)
	if lockMinutes <= 0 {
		lockMinutes = 30
	}

	_, err := g.repo.RecordFailedAttempt(ip, maxFailed, lockMinutes)
	if err != nil {
		return apperror.New("DB_ERROR", "خطا در ثبت تلاش ناموفق.", err.Error(), 500)
	}
	return nil
}

// ResetLoginAttempts پس از ورود موفق شمارنده را صفر می‌کند.
func (g *Guard) ResetLoginAttempts(ip string) error {
	return g.repo.ResetAttempts(ip)
}

// CheckWorkHours بررسی ساعات کاری (Admin مستثنی).
func (g *Guard) CheckWorkHours(roleName string) error {
	if roleName == "Admin" {
		return nil
	}

	now := time.Now()
	dayBlocked, err := g.isDayBlocked(now.Weekday())
	if err != nil {
		return err
	}
	if dayBlocked {
		return apperror.ErrWorkHours
	}

	startStr, _ := g.settings.GetSettingValue(settings.WorkHoursStart)
	endStr, _ := g.settings.GetSettingValue(settings.WorkHoursEnd)

	fmt.Println("startStr", startStr)
	fmt.Println("endStr", endStr)

	start, err := parseTimeOfDay(startStr)
	if err != nil {
		start = 8 * 60 // 08:00
	}
	end, err := parseTimeOfDay(endStr)
	if err != nil {
		end = 18 * 60 // 18:00
	}

	currentMinutes := now.Hour()*60 + now.Minute()
	if currentMinutes < start || currentMinutes >= end {
		return apperror.ErrWorkHours
	}

	return nil
}

func (g *Guard) isDayBlocked(weekday time.Weekday) (bool, error) {
	keyMap := map[time.Weekday]string{
		time.Saturday:  settings.IsSaturdayBlocked,
		time.Sunday:    settings.IsSundayBlocked,
		time.Monday:    settings.IsMondayBlocked,
		time.Tuesday:   settings.IsTuesdayBlocked,
		time.Wednesday: settings.IsWednesdayBlocked,
		time.Thursday:  settings.IsThursdayBlocked,
		time.Friday:    settings.IsFridayBlocked,
	}

	key, ok := keyMap[weekday]
	if !ok {
		return false, nil
	}

	val, err := g.settings.GetSettingValue(key)
	if err != nil {
		return false, err
	}
	return val == "true", nil
}

func parseTimeOfDay(s string) (int, error) {
	var h, m int
	_, err := fmt.Sscanf(s, "%d:%d", &h, &m)
	if err != nil {
		return 0, err
	}
	return h*60 + m, nil
}

// CleanupStaleAttempts ردیف‌های قدیمی را پاک می‌کند.
func (g *Guard) CleanupStaleAttempts() error {
	return g.repo.CleanupStale(24 * time.Hour)
}

// Repository بازگرداندن repo.
func (g *Guard) Repository() Repository {
	return g.repo
}
