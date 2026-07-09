package logretention

import (
	"context"
	"fmt"
	"log"
	"strconv"
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

func (w *Worker) runCycle() {
	w.cleanupSecurityEvents()
	w.cleanupLoginAttempts()
	w.verifyIntegrity()
	w.verifyAuditChain()
}

func (w *Worker) cleanupSecurityEvents() {
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

	count, err := w.audit.Repository().Count()
	if err != nil || count == 0 {
		return
	}

	deleteCount := int(count / 10)
	if deleteCount < 1 {
		deleteCount = 1
	}

	deleted, err := w.audit.Repository().DeleteOldest(deleteCount)
	if err != nil {
		log.Printf("خطا در پاکسازی SecurityEvents: %v", err)
		return
	}

	now := time.Now().UTC()
	summary := LogCleanupSummary{
		TargetTable:  "SecurityEvents",
		DeletedCount: int(deleted),
		CreatedAt:    now,
	}
	_ = w.db.Create(&summary).Error

	_ = w.audit.LogEvent(nil, "system", audit.EventLogCleanup,
		fmt.Sprintf("پاکسازی %d رکورد قدیمی SecurityEvents (حجم: %.2f MB)", deleted, currentSize))
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

func (w *Worker) verifyAuditChain() {
	checkpointStr, err := w.settings.GetSettingValue(settings.LastVerifiedEventID)
	if err != nil {
		return
	}

	checkpoint, _ := strconv.ParseInt(checkpointStr, 10, 64)
	ok, err := w.audit.VerifyChain(checkpoint)
	if err != nil || !ok {
		_ = w.audit.LogEvent(nil, "system", audit.EventDataTampering,
			fmt.Sprintf("شکست تأیید Hash-Chain از checkpoint %d", checkpoint))
		return
	}

	lastID, err := w.audit.GetLastEventID()
	if err == nil && lastID > 0 {
		_ = w.settings.UpdateSetting(settings.LastVerifiedEventID, strconv.FormatInt(lastID, 10), 0, "system")
	}
}
