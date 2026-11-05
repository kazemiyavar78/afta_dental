package audit

import (
	"sync"
	"time"

	"gorm.io/gorm"
)

// Manager رویداد امنیتی را با زنجیره هش ثبت و بررسی می‌کند.
type Manager struct {
	mu       sync.Mutex
	repo     Repository
	auditKey []byte
}

// NewManager مدیر را با مخزن GORM و کلید HMAC زنجیره می‌سازد.
// db اتصال پایگاه است و auditKey همان AUDIT_HMAC_KEY است، جدا از INTEGRITY_HMAC_KEY.
func NewManager(db *gorm.DB, auditKey string) *Manager {
	return &Manager{
		repo:     NewRepository(db),
		auditKey: []byte(auditKey),
	}
}

// LogEvent مسیر ثبت قدیمی است و تا مهاجرت فراخوانی‌ها در گام ۱-۳ و ۱-۴ می‌ماند.
// userID فاعل است و اگر nil باشد بخش کاربر هش خالی می‌ماند. ip، eventType و description فیلدهای ردیف‌اند.
// خروجی خطا یعنی درج انجام نشده. ردیف با HashVersion=1 نوشته می‌شود تا فرمول هش قبلی ثابت بماند.
// FAU_STG.1: دو درخواست همزمان نباید دو ردیف با یک هش قبلی بسازند.
func (m *Manager) LogEvent(userID *int, ip, eventType, description string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.repo.AppendChained(func(prevHash string) (*SecurityEvent, error) {
		// datetimeoffset در SQL Server تا ۱۰۰ نانوثانیه ذخیره می‌کند و ساعت لینوکس ریزتر است.
		// اگر نانوثانیهٔ کامل در هش برود، پس از خواندن ردیف هش نمی‌خواند.
		// برش به میکروثانیه هم در Go و هم در SQL Server پایدار است.
		row := SecurityEvent{
			UserID:      userID,
			Ip:          ip,
			EventType:   eventType,
			Description: description,
			CreatedAt:   time.Now().UTC().Truncate(time.Microsecond),
			PrevHash:    prevHash,
			HashVersion: HashVersionV1,
		}
		row.RowHash = m.computeRowHash(row)
		return &row, nil
	})
}

// TrimVerifiedPrefix فقط ردیف‌های تأییدشده با ID کمتر یا برابر LastVerifiedID را حذف می‌کند.
// state نقطه تأیید است و limit سقف تعداد است. seal بعد از به‌روزرسانی لنگر، HMAC را می‌نویسد و نباید LogEvent صدا بزند.
// خروجی تعداد حذف، بازه CreatedAt همان ردیف‌ها و خطا است. اگر چیزی حذف نشود هر سه تهی‌اند.
// FAU_STG.1 / D-12: هش آخرین ردیف حذفی لنگر ردیف بعدی می‌شود.
func (m *Manager) TrimVerifiedPrefix(state *AuditChainState, limit int, seal func(*AuditChainState)) (deleted int, from, to *time.Time, err error) {
	if state == nil || state.LastVerifiedID <= 0 || limit <= 0 {
		return 0, nil, nil, nil
	}
	m.mu.Lock()
	defer m.mu.Unlock()

	var start, end time.Time
	var saved AuditChainState
	rows, err := m.repo.DeleteVerifiedPrefix(state.LastVerifiedID, limit, func(found []SecurityEvent) (*AuditChainState, error) {
		if len(found) == 0 {
			return nil, nil
		}
		next := *state
		last := found[len(found)-1]
		next.AnchorID = last.ID
		next.AnchorHash = last.RowHash
		if next.LastVerifiedID <= next.AnchorID {
			next.LastVerifiedHash = next.AnchorHash
		}
		if seal != nil {
			seal(&next)
		}
		start = found[0].CreatedAt
		end = last.CreatedAt
		saved = next
		return &saved, nil
	})
	if err != nil || len(rows) == 0 {
		return 0, nil, nil, err
	}
	*state = saved
	return len(rows), &start, &end, nil
}

// LoadChainState ردیف وضعیت را می‌خواند. اگر جدول خالی باشد ردیف صفر با ID ثابت برمی‌گرداند.
func (m *Manager) LoadChainState() (AuditChainState, error) {
	return m.repo.LoadChainState()
}

// SaveChainState ردیف وضعیت را درج یا جایگزین می‌کند.
// state نباید nil باشد. خطا یعنی پایگاه به‌روز نشده است.
func (m *Manager) SaveChainState(state *AuditChainState) error {
	return m.repo.SaveChainState(state)
}

// Repository مخزن را برای خواندن لاگ و حجم جدول برمی‌گرداند.
func (m *Manager) Repository() Repository {
	return m.repo
}
