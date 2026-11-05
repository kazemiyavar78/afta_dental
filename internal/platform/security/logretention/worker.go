package logretention

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/tpdenta/afta-reception/internal/platform/security/audit"
	"github.com/tpdenta/afta-reception/internal/platform/security/integrity"
	"github.com/tpdenta/afta-reception/internal/platform/security/loginguard"
	"github.com/tpdenta/afta-reception/internal/platform/security/settings"
	"gorm.io/gorm"
)

// LogCleanupSummary خلاصه پاکسازی لاگ.
type LogCleanupSummary struct {
	ID           int        `gorm:"column:ID;primaryKey;autoIncrement"`
	TargetTable  string     `gorm:"column:TableName;size:100;not null"`
	DeletedCount int        `gorm:"column:DeletedCount;not null"`
	FromDate     *time.Time `gorm:"column:FromDate"`
	ToDate       *time.Time `gorm:"column:ToDate"`
	CreatedAt    time.Time  `gorm:"column:CreatedAt;not null"`
}

func (LogCleanupSummary) TableName() string { return "LogCleanupSummaries" }

// Worker کارگر پس‌زمینه مدیریت حجم لاگ و بررسی یکپارچگی.
type Worker struct {
	db           *gorm.DB
	settings     *settings.Service
	audit        *audit.Manager
	signer       *integrity.Signer
	loginGuard   *loginguard.Guard
	ticker       time.Duration
	integritySvc IntegrityChecker
}

// IntegrityChecker اینترفیس بررسی یکپارچگی جداول.
type IntegrityChecker interface {
	VerifyAllUsers(ip string) (bool, error)
	VerifyAllRoles(ip string) (bool, error)
}

// NewWorker نمونه Worker می‌سازد.
func NewWorker(
	db *gorm.DB,
	settingsSvc *settings.Service,
	auditMgr *audit.Manager,
	signer *integrity.Signer,
	loginGuard *loginguard.Guard,
	integritySvc IntegrityChecker,
	ticker time.Duration,
) *Worker {
	return &Worker{
		db:           db,
		settings:     settingsSvc,
		audit:        auditMgr,
		signer:       signer,
		loginGuard:   loginGuard,
		ticker:       ticker,
		integritySvc: integritySvc,
	}
}

// Start کارگر پس‌زمینه را اجرا می‌کند.
func (w *Worker) Start(ctx context.Context) {
	t := time.NewTicker(w.ticker)
	defer t.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			w.runCycle()
		}
	}
}

// runCycle یک دور بررسی زنجیره، پاکسازی حجم، پاکسازی قفل ورود و بررسی HMAC جدول‌ها را اجرا می‌کند.
func (w *Worker) runCycle() {
	// بررسی قبل از پاک‌سازی است تا همان دور فقط ردیف‌هایی حذف شوند که همین حالا تأیید شده‌اند.
	w.verifyAuditChain()
	w.cleanupSecurityEvents()
	w.cleanupLoginAttempts()
	w.verifyIntegrity()
}

// cleanupSecurityEvents اگر حجم SecurityEvents از سقف گذشت، فقط پیشوند تأییدشده را حذف می‌کند.
// FAU_STG.1 / D-12: ردیف بررسی‌نشده حذف نمی‌شود و لنگر هش آخرین ردیف حذفی ذخیره می‌شود.
func (w *Worker) cleanupSecurityEvents() {
	if w.signer == nil {
		log.Printf("پاکسازی لاگ رد شد: امضای یکپارچگی نیست")
		return
	}
	maxSizeMB, err := w.settings.GetSettingInt(settings.MaximumSizeOfLogTables)
	if err != nil || maxSizeMB <= 0 {
		maxSizeMB = 100
	}
	currentSize, err := w.audit.Repository().GetTableSizeMB()
	if err != nil {
		log.Printf("خطا در محاسبه حجم SecurityEvents: %v", err)
		return
	}
	if currentSize <= float64(maxSizeMB) {
		return
	}
	deleteCount, ok := w.verifiedDeleteCount()
	if !ok {
		return
	}
	state, err := w.audit.LoadChainState()
	if err != nil {
		log.Printf("خطا در خواندن وضعیت زنجیره برای پاکسازی: %v", err)
		return
	}
	if !state.SealValid(w.signer.Verify) {
		log.Printf("پاکسازی لاگ رد شد: یکپارچگی وضعیت زنجیره نامعتبر است")
		return
	}
	deleted, from, to, err := w.audit.TrimVerifiedPrefix(&state, deleteCount, func(s *audit.AuditChainState) {
		s.Seal(w.signer.Sign)
	})
	if err != nil {
		log.Printf("خطا در پاکسازی SecurityEvents: %v", err)
		return
	}
	if deleted == 0 {
		return
	}
	w.writeCleanupSummary(deleted, from, to, currentSize)
}

// verifiedDeleteCount سقف حذف را ده درصد تعداد ردیف‌ها می‌گذارد و حداقل یک ردیف است.
// خروجی دوم false یعنی شمارش خطا داده یا جدولی خالی است.
func (w *Worker) verifiedDeleteCount() (int, bool) {
	count, err := w.audit.Repository().Count()
	if err != nil || count == 0 {
		if err != nil {
			log.Printf("خطا در شمارش SecurityEvents: %v", err)
		}
		return 0, false
	}
	deleteCount := int(count / 10)
	if deleteCount < 1 {
		deleteCount = 1
	}
	return deleteCount, true
}

// writeCleanupSummary خلاصه پاکسازی را با بازه واقعی تاریخ ردیف‌های حذف‌شده ذخیره می‌کند.
// خطای درج خلاصه و خطای رویداد ممیزی هر دو در لاگ سرور می‌مانند و با _ نادیده گرفته نمی‌شوند.
func (w *Worker) writeCleanupSummary(deleted int, from, to *time.Time, currentSize float64) {
	summary := LogCleanupSummary{
		TargetTable:  "SecurityEvents",
		DeletedCount: deleted,
		FromDate:     from,
		ToDate:       to,
		CreatedAt:    time.Now().UTC().Truncate(time.Microsecond),
	}
	if err := w.db.Create(&summary).Error; err != nil {
		log.Printf("خطا در ثبت خلاصه پاکسازی لاگ: %v", err)
		return
	}
	if err := w.audit.LogEvent(nil, "system", audit.EventLogCleanup,
		fmt.Sprintf("پاکسازی %d رکورد قدیمی SecurityEvents (حجم: %.2f MB)", deleted, currentSize)); err != nil {
		log.Printf("خطا در ثبت رویداد پاکسازی لاگ: %v", err)
	}
}

func (w *Worker) cleanupLoginAttempts() {
	if err := w.loginGuard.CleanupStaleAttempts(); err != nil {
		log.Printf("خطا در پاکسازی LoginAttempts: %v", err)
	}
}

func (w *Worker) verifyIntegrity() {
	if w.integritySvc == nil {
		return
	}

	ok, err := w.settings.VerifyAllIntegrity("system")
	if err != nil || !ok {
		log.Printf("شکست یکپارچگی SecuritySettings")
	}

	ok, err = w.integritySvc.VerifyAllUsers("system")
	if err != nil || !ok {
		log.Printf("شکست یکپارچگی Users")
	}

	ok, err = w.integritySvc.VerifyAllRoles("system")
	if err != nil || !ok {
		log.Printf("شکست یکپارچگی Roles")
	}
}

// verifyAuditChain زنجیره را از وضعیت ذخیره‌شده بررسی می‌کند و فقط خرابی تازه را لاگ می‌کند.
// بررسی موفق رویداد ممیزی نمی‌سازد تا جدول لاگ هر دور پر نشود.
func (w *Worker) verifyAuditChain() {
	if w.signer == nil {
		log.Printf("بررسی زنجیره ممیزی رد شد: امضای یکپارچگی نیست")
		return
	}
	state, err := w.audit.LoadChainState()
	if err != nil {
		log.Printf("خطا در خواندن وضعیت زنجیره ممیزی: %v", err)
		return
	}
	// ردیف وضعیت اگر HMAC نخورد قابل اعتماد نیست و بررسی از صفر انجام می‌شود.
	// FAU_STG.2 / R1-24: مدیر نمی‌تواند با عوض کردن یک تنظیم، checkpoint را جلو بیندازد.
	untrusted := !state.SealValid(w.signer.Verify)
	if untrusted {
		log.Printf("یکپارچگی وضعیت زنجیره ممیزی نامعتبر است؛ بررسی از ابتدا")
		state = audit.AuditChainState{}
	}
	lastGoodID, lastGoodHash, brokenAtID, err := w.audit.VerifyChain(state)
	if err != nil {
		log.Printf("خطا در بررسی زنجیره ممیزی: %v", err)
		return
	}
	now := time.Now().UTC().Truncate(time.Microsecond)
	report := audit.AdvanceChainState(&state, lastGoodID, lastGoodHash, brokenAtID, now)
	state.Seal(w.signer.Sign)
	if err := w.audit.SaveChainState(&state); err != nil {
		log.Printf("خطا در ذخیره وضعیت زنجیره ممیزی: %v", err)
		return
	}
	if !untrusted && !report {
		return
	}
	description := "یکپارچگی ردیف وضعیت زنجیره ممیزی نامعتبر بود"
	if brokenAtID != 0 {
		description = fmt.Sprintf("شکست زنجیره ممیزی در رویداد %d", brokenAtID)
	}
	if err := w.audit.LogEvent(nil, "system", audit.EventDataTampering, description); err != nil {
		log.Printf("خطا در ثبت دستکاری زنجیره ممیزی: %v", err)
	}
}
