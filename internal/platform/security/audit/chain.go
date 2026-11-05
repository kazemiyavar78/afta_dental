package audit

import (
	"strconv"
	"time"
)

const (
	// ChainResultOK یعنی آخرین بررسی زنجیره سالم بوده است.
	ChainResultOK = "ok"
	// ChainResultBroken یعنی بررسی روی یک ردیف متوقف شده است.
	ChainResultBroken = "broken"
	// chainStateRowID تنها ردیف جدول AuditChainStates است.
	chainStateRowID = 1
)

// chainVerifyBatchSize اندازه هر خواندن VerifyChain است تا کل جدول در حافظه نیاید.
var chainVerifyBatchSize = 500

// AuditChainState وضعیت بررسی زنجیره ممیزی است؛ یک ردیف در AuditChainStates.
// LastVerifiedID و LastVerifiedHash آخرین ردیف سالم‌اند.
// AnchorID و AnchorHash هش آخرین ردیفی هستند که پاک‌سازی حذف کرده است.
// IntegrityHash با INTEGRITY_HMAC_KEY (جدا از کلید زنجیره رویداد) ساخته می‌شود.
// FAU_STG.2 / R1-24: این ردیف تنظیم قابل ویرایش مدیر نیست.
type AuditChainState struct {
	ID               int       `gorm:"column:ID;primaryKey;autoIncrement:false"`
	LastVerifiedID   int64     `gorm:"column:LastVerifiedID;not null"`
	LastVerifiedHash string    `gorm:"column:LastVerifiedHash;size:128;not null"`
	AnchorID         int64     `gorm:"column:AnchorID;not null"`
	AnchorHash       string    `gorm:"column:AnchorHash;size:128;not null"`
	LastCheckAt      time.Time `gorm:"column:LastCheckAt;not null"`
	LastResult       string    `gorm:"column:LastResult;size:16;not null"`
	BrokenAtID       int64     `gorm:"column:BrokenAtID;not null"`
	IntegrityHash    string    `gorm:"column:IntegrityHash;size:128;not null"`
}

// TableName نام جدول وضعیت زنجیره را برمی‌گرداند.
func (AuditChainState) TableName() string {
	return "AuditChainStates"
}

// IntegrityFields ورودی HMAC وضعیت زنجیره را به ترتیب ثابت برمی‌گرداند.
// state همان ردیف است. خروجی رشته‌هایی است که Signer با INTEGRITY_HMAC_KEY امضا می‌کند.
// زمان به میکروثانیهٔ یونیکس می‌رود تا قالب RFC3339 و منطقه زمانی بعد از خواندن از SQL Server هش را عوض نکند.
func (s AuditChainState) IntegrityFields() []string {
	at := s.LastCheckAt.UTC().Truncate(time.Microsecond)
	return []string{
		strconv.FormatInt(s.LastVerifiedID, 10),
		s.LastVerifiedHash,
		strconv.FormatInt(s.AnchorID, 10),
		s.AnchorHash,
		strconv.FormatInt(at.UnixMicro(), 10),
		s.LastResult,
		strconv.FormatInt(s.BrokenAtID, 10),
	}
}

// Seal هش یکپارچگی را با تابع امضا می‌نویسد.
// sign همان integrity.Signer.Sign است. زمان بررسی به میکروثانیه بریده می‌شود تا پس از خواندن از SQL Server عوض نشود.
func (s *AuditChainState) Seal(sign func(fields ...string) string) {
	if s == nil || sign == nil {
		return
	}
	s.normalize()
	s.IntegrityHash = sign(s.IntegrityFields()...)
}

// SealValid درست است اگر ردیف هنوز ذخیره نشده باشد یا HMAC فعلی با فیلدها بخواند.
// verify همان integrity.Signer.Verify است. ردیف خالی اولیه معتبر است تا بررسی اول از صفر شروع شود.
func (s AuditChainState) SealValid(verify func(hash string, fields ...string) bool) bool {
	if s.untouched() {
		return true
	}
	if verify == nil || s.IntegrityHash == "" {
		return false
	}
	return verify(s.IntegrityHash, s.IntegrityFields()...)
}

// untouched یعنی هنوز بررسی یا لنگر ذخیره نشده است.
func (s AuditChainState) untouched() bool {
	return s.IntegrityHash == "" && s.LastVerifiedID == 0 && s.LastVerifiedHash == "" &&
		s.AnchorID == 0 && s.AnchorHash == "" && s.LastResult == "" && s.BrokenAtID == 0
}

// normalize زمان بررسی را به UTC و میکروثانیه محدود می‌کند.
func (s *AuditChainState) normalize() {
	if s == nil {
		return
	}
	s.ID = chainStateRowID
	s.LastCheckAt = s.LastCheckAt.UTC().Truncate(time.Microsecond)
}

// chainStartHash هش قبلیِ اولین ردیف بررسی‌نشده را برمی‌گرداند.
// اگر هنوز ردیف سالمی ثبت نشده و لنگر پاک‌سازی هست، لنگر؛ وگرنه LastVerifiedHash (برای شروع زنجیره، خالی).
func chainStartHash(state AuditChainState) string {
	if state.LastVerifiedID == 0 && state.AnchorHash != "" {
		return state.AnchorHash
	}
	return state.LastVerifiedHash
}

// VerifyChain رویدادهای با ID بزرگ‌تر از LastVerifiedID را دسته‌ای و به ترتیب بررسی می‌کند.
// state نقطه شروع است. خروجی آخرین ردیف سالم، هش آن، شناسه اولین ردیف خراب (صفر اگر سالم باشد) و خطای خواندن است.
// وضعیت فقط تا آخرین ردیف سالم جلو می‌رود؛ ردیف خراب و بعد از آن تأیید نمی‌شوند.
// FAU_STG.1 / D-11: هش شروع، هش همان checkpoint است نه هش آخرین ردیف جدول.
func (m *Manager) VerifyChain(state AuditChainState) (lastGoodID int64, lastGoodHash string, brokenAtID int64, err error) {
	prevHash := chainStartHash(state)
	lastGoodID = state.LastVerifiedID
	lastGoodHash = prevHash
	afterID := state.LastVerifiedID
	limit := chainVerifyBatchSize
	if limit <= 0 {
		limit = 500
	}
	for {
		batch, readErr := m.repo.FindAfterID(afterID, limit)
		if readErr != nil {
			return lastGoodID, lastGoodHash, 0, readErr
		}
		if len(batch) == 0 {
			return lastGoodID, lastGoodHash, 0, nil
		}
		for _, event := range batch {
			if !m.rowMatches(event, prevHash) {
				return lastGoodID, lastGoodHash, event.ID, nil
			}
			prevHash = event.RowHash
			lastGoodID = event.ID
			lastGoodHash = event.RowHash
		}
		afterID = batch[len(batch)-1].ID
		if len(batch) < limit {
			return lastGoodID, lastGoodHash, 0, nil
		}
	}
}

// rowMatches یعنی PrevHash ردیف با هش قبلی یکی است و RowHash از همان فیلدها دوباره ساخته می‌شود.
func (m *Manager) rowMatches(event SecurityEvent, prevHash string) bool {
	if event.PrevHash != prevHash {
		return false
	}
	expected := m.computeRowHash(event.PrevHash, event.UserID, event.Ip, event.EventType, event.Description, event.CreatedAt)
	return expected == event.RowHash
}

// AdvanceChainState نتیجه VerifyChain را در state می‌نویسد.
// lastGoodID و lastGoodHash آخرین ردیف سالم‌اند، brokenAtID شناسه خرابی یا صفر، now زمان بررسی است.
// خروجی true یعنی BrokenAtID تازه است و باید یک بار DATA_TAMPERING ثبت شود.
// FAU_STG.2: خرابی تکراری رویداد تازه نمی‌سازد.
func AdvanceChainState(state *AuditChainState, lastGoodID int64, lastGoodHash string, brokenAtID int64, now time.Time) bool {
	if state == nil {
		return false
	}
	prevBroken := state.BrokenAtID
	state.LastCheckAt = now.UTC().Truncate(time.Microsecond)
	state.LastVerifiedID = lastGoodID
	state.LastVerifiedHash = lastGoodHash
	if brokenAtID != 0 {
		state.LastResult = ChainResultBroken
		state.BrokenAtID = brokenAtID
		return brokenAtID != prevBroken
	}
	state.LastResult = ChainResultOK
	state.BrokenAtID = 0
	return false
}
