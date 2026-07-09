package audit

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
)

// Manager مدیریت رویدادهای امنیتی با Hash-Chain.
type Manager struct {
	repo     Repository
	auditKey []byte
}

// NewManager نمونه Manager می‌سازد.
func NewManager(db *gorm.DB, auditKey string) *Manager {
	return &Manager{
		repo:     NewRepository(db),
		auditKey: []byte(auditKey),
	}
}

// computeRowHash هش زنجیره‌ای ردیف را محاسبه می‌کند.
func (m *Manager) computeRowHash(prevHash string, userID *int, ip, eventType, description string, createdAt time.Time) string {
	uid := ""
	if userID != nil {
		uid = strconv.Itoa(*userID)
	}
	payload := strings.Join([]string{
		prevHash,
		uid,
		ip,
		eventType,
		description,
		createdAt.UTC().Format(time.RFC3339Nano),
	}, "|")

	mac := hmac.New(sha256.New, m.auditKey)
	mac.Write([]byte(payload))
	return hex.EncodeToString(mac.Sum(nil))
}

// LogEvent رویداد امنیتی جدید با Hash-Chain ثبت می‌کند.
func (m *Manager) LogEvent(userID *int, ip, eventType, description string) error {
	prevHash, err := m.repo.GetLastRowHash()
	if err != nil {
		return fmt.Errorf("خواندن آخرین هش: %w", err)
	}

	now := time.Now().UTC()
	rowHash := m.computeRowHash(prevHash, userID, ip, eventType, description, now)

	event := &SecurityEvent{
		UserID:      userID,
		Ip:          ip,
		EventType:   eventType,
		Description: description,
		CreatedAt:   now,
		PrevHash:    prevHash,
		RowHash:     rowHash,
	}

	return m.repo.Create(event)
}

// VerifyChain زنجیره هش را از یک checkpoint به بعد بررسی می‌کند.
func (m *Manager) VerifyChain(fromID int64) (bool, error) {
	events, err := m.repo.FindFromID(fromID)
	if err != nil {
		return false, err
	}

	var prevHash string
	if fromID > 1 {
		earlier, err := m.repo.FindFromID(fromID - 1)
		if err != nil {
			return false, err
		}
		if len(earlier) > 0 {
			prevHash = earlier[len(earlier)-1].RowHash
		}
	}

	for _, e := range events {
		if e.PrevHash != prevHash {
			return false, nil
		}
		expected := m.computeRowHash(e.PrevHash, e.UserID, e.Ip, e.EventType, e.Description, e.CreatedAt)
		if expected != e.RowHash {
			return false, nil
		}
		prevHash = e.RowHash
	}

	return true, nil
}

// GetLastEventID آخرین شناسه رویداد را برمی‌گرداند.
func (m *Manager) GetLastEventID() (int64, error) {
	return m.repo.GetLastEventID()
}

// Repository بازگرداندن repo برای worker.
func (m *Manager) Repository() Repository {
	return m.repo
}
