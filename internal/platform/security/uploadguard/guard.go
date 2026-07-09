package uploadguard

import (
	"time"

	"github.com/tpdenta/afta-reception/internal/platform/apperror"
	"github.com/tpdenta/afta-reception/internal/platform/security/settings"
	"gorm.io/gorm"
)

// UploadCounter شمارنده آپلود اتمیک.
type UploadCounter struct {
	UserID      int       `gorm:"column:UserID;primaryKey"`
	WindowStart time.Time `gorm:"column:WindowStart;primaryKey"`
	UploadCount int       `gorm:"column:UploadCount;not null;default:0"`
}

func (UploadCounter) TableName() string { return "UploadCounters" }

// Guard محدودیت تعداد آپلود بر اساس دیتابیس.
type Guard struct {
	db       *gorm.DB
	settings *settings.Service
}

// NewGuard نمونه Guard می‌سازد.
func NewGuard(db *gorm.DB, settingsSvc *settings.Service) *Guard {
	return &Guard{db: db, settings: settingsSvc}
}

// CheckAndIncrement بررسی و افزایش اتمیک شمارنده آپلود.
func (g *Guard) CheckAndIncrement(userID int) error {
	maxUploads, _ := g.settings.GetSettingInt(settings.MaximumUploadsPerWindow)
	if maxUploads <= 0 {
		maxUploads = 10
	}
	windowMinutes, _ := g.settings.GetSettingInt(settings.UploadWindowMinutes)
	if windowMinutes <= 0 {
		windowMinutes = 60
	}

	now := time.Now().UTC()
	windowStart := now.Truncate(time.Duration(windowMinutes) * time.Minute)

	var count int
	err := g.db.Transaction(func(tx *gorm.DB) error {
		result := tx.Exec(`
			MERGE UploadCounters WITH (HOLDLOCK) AS target
			USING (SELECT @p1 AS UserID, @p2 AS WindowStart) AS source
			ON target.UserID = source.UserID AND target.WindowStart = source.WindowStart
			WHEN MATCHED THEN
				UPDATE SET UploadCount = target.UploadCount + 1
			WHEN NOT MATCHED THEN
				INSERT (UserID, WindowStart, UploadCount) VALUES (@p1, @p2, 1);
		`, userID, windowStart)

		if result.Error != nil {
			return result.Error
		}

		var counter UploadCounter
		if err := tx.Where("UserID = ? AND WindowStart = ?", userID, windowStart).First(&counter).Error; err != nil {
			return err
		}
		count = counter.UploadCount
		return nil
	})

	if err != nil {
		return apperror.New("DB_ERROR", "خطا در بررسی محدودیت آپلود.", err.Error(), 500)
	}

	if count > maxUploads {
		return apperror.New("UPLOAD_LIMIT", "تعداد آپلود در این بازه بیش از حد مجاز است.", "upload limit exceeded", 429)
	}

	return nil
}
