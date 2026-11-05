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

// IntegrityChecker فهرست ردیف‌های کاربر و نقش است که HMAC نمی‌خورند.
// پیاده‌سازی رویداد نمی‌سازد. ثبت یافته در audit.ApplyIntegrityFindings است.
type IntegrityChecker interface {
	InvalidUsers() ([]audit.IntegritySubject, error)
	InvalidRoles() ([]audit.IntegritySubject, error)
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
// دور اول همان لحظه اجرا می‌شود تا AuditChainStates و وضعیت صحت از ابتدا پر باشند.
// دورهای بعد به فاصله ticker هستند. لغو ctx حلقه را بعد از دور جاری متوقف می‌کند.
func (w *Worker) Start(ctx context.Context) {
	ticker := time.NewTicker(w.ticker)
	defer ticker.Stop()
	runCycles(ctx, ticker.C, w.runCycle)
}

// runCycles یک دور را بلافاصله و بعد با هر تیک اجرا می‌کند.
// ctx لغو حلقه است. ticks کانال تیک است. cycle همان کار یک دور است و خروجی ندارد.
func runCycles(ctx context.Context, ticks <-chan time.Time, cycle func()) {
	if ctx.Err() != nil {
		return
	}
	cycle()
	for {
		select {
		case <-ctx.Done():
			return
		case _, ok := <-ticks:
			if !ok {
				return
			}
			cycle()
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
	if err := w.audit.Record(context.Background(), audit.Event{
		Type:        audit.EventLogCleanup,
		Result:      audit.ResultSuccess,
		Actor:       audit.SystemActor(),
		Description: fmt.Sprintf("پاکسازی %d رکورد قدیمی SecurityEvents (حجم: %.2f MB)", deleted, currentSize),
		Details:     map[string]any{"deleted": deleted, "size_mb": currentSize},
	}); err != nil {
		log.Printf("خطا در ثبت رویداد پاکسازی لاگ: %v", err)
	}
}

func (w *Worker) cleanupLoginAttempts() {
	if err := w.loginGuard.CleanupStaleAttempts(); err != nil {
		log.Printf("خطا در پاکسازی LoginAttempts: %v", err)
	}
}

// verifyIntegrity کاربر، نقش و تنظیم را تا آخر بررسی می‌کند و یافته تکراری نمی‌سازد.
// خطا فقط در لاگ سرور می‌ماند. شکست یک نوع، نوع‌های دیگر را حذف نمی‌کند.
// FDP_SDI.2: هشدار تازه فقط برای ردیفی است که قبلاً یافته باز نداشته.
func (w *Worker) verifyIntegrity() {
	if w.integritySvc == nil || w.db == nil || w.audit == nil || w.settings == nil {
		return
	}
	store := audit.NewFindingStore(w.db)
	now := time.Now().UTC().Truncate(time.Microsecond)
	w.applyScan(store, now, audit.TargetSetting, w.settings.InvalidSettings)
	w.applyScan(store, now, audit.TargetUser, w.integritySvc.InvalidUsers)
	w.applyScan(store, now, audit.TargetRole, w.integritySvc.InvalidRoles)
}

// applyScan یک نوع را تا آخر می‌خواند و یافته‌ها را با تابع مشترک هم‌تراز می‌کند.
// list فقط ردیف‌های نامعتبر را می‌دهد. خطای خواندن همان نوع را رد می‌کند تا یافته باز بسته نشود.
func (w *Worker) applyScan(store audit.FindingStore, now time.Time, entityType string, list func() ([]audit.IntegritySubject, error)) {
	rows, err := list()
	if err != nil {
		log.Printf("خطا در بررسی یکپارچگی %s: %v", entityType, err)
		return
	}
	if err := audit.ApplyIntegrityFindings(context.Background(), store, w.audit, entityType, rows, now); err != nil {
		log.Printf("خطا در ثبت یافته یکپارچگی %s: %v", entityType, err)
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
	if err := w.audit.Record(context.Background(), audit.Event{
		Type:        audit.EventDataTampering,
		Result:      audit.ResultFailure,
		Reason:      audit.ReasonIntegrityViolation,
		Actor:       audit.SystemActor(),
		Description: description,
	}); err != nil {
		log.Printf("خطا در ثبت دستکاری زنجیره ممیزی: %v", err)
	}
}
