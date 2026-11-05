package audit

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/tpdenta/afta-reception/internal/platform/security/integrity"
)

// newTestManager مدیر و مخزن حافظه‌ای را با کلید ثابت تست می‌سازد.
func newTestManager(t *testing.T) (*Manager, *memoryRepo) {
	t.Helper()
	repo := newMemoryRepo()
	m := &Manager{
		repo:     repo,
		auditKey: []byte("test-audit-key-32bytes-long!!!!!"),
	}
	return m, repo
}

// appendN به تعداد n رویداد ورود موفق پشت سر هم ثبت می‌کند.
func appendN(t *testing.T, m *Manager, n int) {
	t.Helper()
	for i := 0; i < n; i++ {
		if err := m.LogEvent(nil, "127.0.0.1", EventLoginSuccess, fmt.Sprintf("رویداد %d", i)); err != nil {
			t.Fatalf("ثبت رویداد %d: %v", i, err)
		}
	}
}

// tamperDescription توضیح ردیف id را عوض می‌کند و هش ذخیره‌شده را دست نمی‌زند.
func tamperDescription(t *testing.T, repo *memoryRepo, id int64) {
	t.Helper()
	for i := range repo.events {
		if repo.events[i].ID == id {
			repo.events[i].Description = "دستکاری‌شده"
			return
		}
	}
	t.Fatalf("ردیف %d پیدا نشد", id)
}

// countEventType تعداد ردیف‌های یک نوع رویداد را می‌شمارد.
func countEventType(repo *memoryRepo, eventType string) int {
	n := 0
	for _, event := range repo.events {
		if event.EventType == eventType {
			n++
		}
	}
	return n
}

// TestD11_HealthyChainAdvancesState زنجیره ۱۰۰ ردیفی را دسته‌ای تأیید می‌کند و state را تا ردیف آخر جلو می‌برد.
func TestD11_HealthyChainAdvancesState(t *testing.T) {
	prev := chainVerifyBatchSize
	chainVerifyBatchSize = 40
	t.Cleanup(func() { chainVerifyBatchSize = prev })

	m, repo := newTestManager(t)
	appendN(t, m, 100)
	lastID, hash, broken, err := m.VerifyChain(AuditChainState{})
	if err != nil {
		t.Fatalf("بررسی: %v", err)
	}
	if broken != 0 || lastID != 100 || hash == "" || hash != repo.events[99].RowHash {
		t.Fatalf("نتیجه سالم نیست: id=%d broken=%d", lastID, broken)
	}
	if repo.findAfterCalls != 3 {
		t.Fatalf("۱۰۰ ردیف با دسته ۴۰تایی باید ۳ بار خوانده شود، %d بار خوانده شد", repo.findAfterCalls)
	}
	state := AuditChainState{}
	if AdvanceChainState(&state, lastID, hash, broken, time.Now().UTC()) {
		t.Fatal("زنجیره سالم نباید رویداد دستکاری بسازد")
	}
	if state.LastVerifiedID != 100 || state.LastVerifiedHash != hash || state.LastResult != ChainResultOK || state.BrokenAtID != 0 {
		t.Fatalf("state جلو نرفت: %+v", state)
	}
}

// TestD11_TamperedDescriptionStopsAtPreviousRow تغییر توضیح ردیف ۴۰ را همان‌جا متوقف می‌کند و state را روی ۳۹ می‌گذارد.
func TestD11_TamperedDescriptionStopsAtPreviousRow(t *testing.T) {
	m, repo := newTestManager(t)
	appendN(t, m, 100)
	tamperDescription(t, repo, 40)

	lastID, hash, broken, err := m.VerifyChain(AuditChainState{})
	if err != nil {
		t.Fatalf("بررسی: %v", err)
	}
	if broken != 40 || lastID != 39 {
		t.Fatalf("انتظار توقف روی ۴۰ و ماندن روی ۳۹، id=%d broken=%d", lastID, broken)
	}
	state := AuditChainState{}
	AdvanceChainState(&state, lastID, hash, broken, time.Now().UTC())
	if state.LastVerifiedID != 39 || state.BrokenAtID != 40 || state.LastResult != ChainResultBroken {
		t.Fatalf("state باید روی ۳۹ بماند: %+v", state)
	}
}

// TestD11_RepeatCheckDoesNotRepeatTamperEvent بررسی دوباره همان خرابی را فقط یک بار ثبت می‌کند.
func TestD11_RepeatCheckDoesNotRepeatTamperEvent(t *testing.T) {
	m, repo := newTestManager(t)
	appendN(t, m, 100)
	tamperDescription(t, repo, 40)

	state := AuditChainState{}
	now := time.Now().UTC().Truncate(time.Microsecond)
	for round := 0; round < 2; round++ {
		lastID, hash, broken, err := m.VerifyChain(state)
		if err != nil {
			t.Fatalf("دور %d: %v", round, err)
		}
		if broken != 40 || lastID != 39 {
			t.Fatalf("دور %d: id=%d broken=%d", round, lastID, broken)
		}
		if AdvanceChainState(&state, lastID, hash, broken, now) {
			if err := m.LogEvent(nil, "system", EventDataTampering, "شکست زنجیره ممیزی در رویداد 40"); err != nil {
				t.Fatalf("ثبت هشدار: %v", err)
			}
		}
	}
	if got := countEventType(repo, EventDataTampering); got != 1 {
		t.Fatalf("رویداد دستکاری %d بار ثبت شد", got)
	}
}

// TestD12_CleanupPrefixThenVerifyRemaining سی ردیف تأییدشده اول را حذف می‌کند و ادامه زنجیره را سالم می‌بیند.
func TestD12_CleanupPrefixThenVerifyRemaining(t *testing.T) {
	signer := integrity.NewSigner("test-integrity-key-32bytes-long!!")
	m, repo := newTestManager(t)
	appendN(t, m, 30)

	state := AuditChainState{}
	now := time.Now().UTC().Truncate(time.Microsecond)
	lastID, hash, broken, err := m.VerifyChain(state)
	if err != nil || broken != 0 || lastID != 30 {
		t.Fatalf("تأیید پیشوند: id=%d broken=%d err=%v", lastID, broken, err)
	}
	if AdvanceChainState(&state, lastID, hash, broken, now) {
		t.Fatal("پیشوند سالم نباید هشدار بسازد")
	}
	appendN(t, m, 70)

	deleted, from, to, err := m.TrimVerifiedPrefix(&state, 30, func(s *AuditChainState) {
		s.Seal(signer.Sign)
	})
	if err != nil {
		t.Fatalf("پاکسازی: %v", err)
	}
	if deleted != 30 || from == nil || to == nil || to.Before(*from) {
		t.Fatalf("بازه حذف نادرست است: n=%d from=%v to=%v", deleted, from, to)
	}
	if state.AnchorID != 30 || state.AnchorHash == "" || state.LastVerifiedHash != state.AnchorHash {
		t.Fatalf("لنگر باید ردیف ۳۰ باشد: %+v", state)
	}
	if repo.hasID(1) || repo.hasID(30) || !repo.hasID(31) || !repo.hasID(100) {
		t.Fatal("فقط ردیف‌های تأییدشده باید حذف شوند")
	}
	lastID, _, broken, err = m.VerifyChain(state)
	if err != nil || broken != 0 || lastID != 100 {
		t.Fatalf("ادامه بررسی: id=%d broken=%d err=%v", lastID, broken, err)
	}
}

// TestD12_AppendAfterDeletingVerifiedTipUsesAnchor بعد از حذف همه ردیف‌های تأییدشده، رویداد جدید را به لنگر وصل می‌کند.
func TestD12_AppendAfterDeletingVerifiedTipUsesAnchor(t *testing.T) {
	signer := integrity.NewSigner("test-integrity-key-32bytes-long!!")
	m, repo := newTestManager(t)
	appendN(t, m, 5)
	state := AuditChainState{}
	lastID, hash, broken, err := m.VerifyChain(state)
	if err != nil || broken != 0 || lastID != 5 {
		t.Fatalf("تأیید: id=%d broken=%d err=%v", lastID, broken, err)
	}
	AdvanceChainState(&state, lastID, hash, broken, time.Now().UTC())
	deleted, _, _, err := m.TrimVerifiedPrefix(&state, 5, func(s *AuditChainState) {
		s.Seal(signer.Sign)
	})
	if err != nil || deleted != 5 {
		t.Fatalf("حذف نوک زنجیره: n=%d err=%v", deleted, err)
	}
	if err := m.LogEvent(nil, "127.0.0.1", EventLoginSuccess, "بعد از پاکسازی"); err != nil {
		t.Fatalf("ثبت بعد از پاکسازی: %v", err)
	}
	if len(repo.events) != 1 || repo.events[0].PrevHash != state.AnchorHash {
		t.Fatal("رویداد جدید باید PrevHash برابر لنگر داشته باشد")
	}
	lastID, _, broken, err = m.VerifyChain(state)
	if err != nil || broken != 0 || lastID != repo.events[0].ID {
		t.Fatalf("بررسی بعد از لنگر: id=%d broken=%d err=%v", lastID, broken, err)
	}
}

// TestD12_ZeroCheckpointUsesAnchorHash وقتی LastVerifiedID صفر است بررسی از AnchorHash شروع می‌شود.
func TestD12_ZeroCheckpointUsesAnchorHash(t *testing.T) {
	m, repo := newTestManager(t)
	repo.state.AnchorHash = "anchor-hash"
	appendN(t, m, 3)
	if repo.events[0].PrevHash != "anchor-hash" {
		t.Fatal("اولین رویداد بعد از جدول خالی باید به لنگر وصل شود")
	}
	lastID, _, broken, err := m.VerifyChain(AuditChainState{AnchorHash: "anchor-hash"})
	if err != nil || broken != 0 || lastID != 3 {
		t.Fatalf("بررسی از لنگر: id=%d broken=%d err=%v", lastID, broken, err)
	}
}

// TestD11_ConcurrentLogEventKeepsSingleChain بیست ثبت همزمان باید یک زنجیره سالم بسازند.
func TestD11_ConcurrentLogEventKeepsSingleChain(t *testing.T) {
	m, _ := newTestManager(t)
	const n = 20
	errCh := make(chan error, n)
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			errCh <- m.LogEvent(nil, "127.0.0.1", EventLoginSuccess, fmt.Sprintf("همزمان %d", i))
		}(i)
	}
	wg.Wait()
	close(errCh)
	for err := range errCh {
		if err != nil {
			t.Fatal(err)
		}
	}
	lastID, _, broken, err := m.VerifyChain(AuditChainState{})
	if err != nil || broken != 0 || lastID != n {
		t.Fatalf("زنجیره همزمان سالم نیست: id=%d broken=%d err=%v", lastID, broken, err)
	}
}

// TestCreatedAt_MicrosecondRoundTripKeepsHash هش زمان بریده‌شده را بعد از تجزیه دوباره همان می‌داند.
func TestCreatedAt_MicrosecondRoundTripKeepsHash(t *testing.T) {
	m, repo := newTestManager(t)
	raw := time.Date(2026, 11, 5, 19, 47, 1, 123456789, time.FixedZone("UTC+3:30", 3*3600+30*60))
	stored := raw.UTC().Truncate(time.Microsecond)
	roundTrip, err := time.Parse(time.RFC3339Nano, stored.Format(time.RFC3339Nano))
	if err != nil {
		t.Fatalf("تجزیه زمان: %v", err)
	}
	before := m.computeRowHash("prev", nil, "127.0.0.1", EventLoginSuccess, "ping", stored)
	after := m.computeRowHash("prev", nil, "127.0.0.1", EventLoginSuccess, "ping", roundTrip)
	if before != after {
		t.Fatalf("هش قبل و بعد از رفت‌وبرگشت زمان یکی نیست")
	}
	if err := m.LogEvent(nil, "127.0.0.1", EventLoginSuccess, "ping"); err != nil {
		t.Fatalf("ثبت: %v", err)
	}
	got := repo.events[0]
	if got.CreatedAt.Nanosecond()%1000 != 0 {
		t.Fatalf("CreatedAt باید مضرب میکروثانیه باشد: %d", got.CreatedAt.Nanosecond())
	}
	parsed, err := time.Parse(time.RFC3339Nano, got.CreatedAt.UTC().Format(time.RFC3339Nano))
	if err != nil {
		t.Fatalf("تجزیه زمان ذخیره‌شده: %v", err)
	}
	hashed := m.computeRowHash(got.PrevHash, got.UserID, got.Ip, got.EventType, got.Description, got.CreatedAt)
	rehashed := m.computeRowHash(got.PrevHash, got.UserID, got.Ip, got.EventType, got.Description, parsed)
	if hashed != rehashed || hashed != got.RowHash {
		t.Fatal("هش ذخیره‌شده با زمان رفت‌وبرگشت‌شده نمی‌خواند")
	}
}

// TestAuditChainState_SealRejectsTamper امضای سالم را می‌پذیرد و تغییر فیلد را رد می‌کند.
func TestAuditChainState_SealRejectsTamper(t *testing.T) {
	signer := integrity.NewSigner("test-integrity-key-32bytes-long!!")
	fresh := AuditChainState{}
	if !fresh.SealValid(signer.Verify) {
		t.Fatal("وضعیت خالی اولیه باید معتبر باشد")
	}
	state := AuditChainState{LastVerifiedID: 3, LastVerifiedHash: "abc", LastResult: ChainResultOK}
	state.Seal(signer.Sign)
	if !state.SealValid(signer.Verify) {
		t.Fatal("امضای تازه‌ساخته باید معتبر باشد")
	}
	state.LastVerifiedID = 99
	if state.SealValid(signer.Verify) {
		t.Fatal("تغییر شناسه تأییدشده باید امضا را باطل کند")
	}
}
