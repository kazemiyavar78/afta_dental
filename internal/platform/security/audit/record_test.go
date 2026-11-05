package audit

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/tpdenta/afta-reception/internal/platform/apperror"
)

// TestRecord_FullActor نام، هدف، نتیجه، دلیل و جزئیات ماسک‌شده را با هش نسخه ۲ ذخیره می‌کند.
func TestRecord_FullActor(t *testing.T) {
	m, repo := newTestManager(t)
	userID := 7
	err := m.Record(context.Background(), Event{
		Type:   EventRoleCreated,
		Result: ResultSuccess,
		Actor: Actor{
			UserID:    &userID,
			Username:  "narges",
			IP:        "192.0.2.15",
			UserAgent: "Reception/1",
		},
		Target: Target{Type: TargetRole, ID: "3", Name: "پذیرش"},
		Details: map[string]any{
			"password": "hidden-secret",
			"name":     "پذیرش",
		},
	})
	if err != nil {
		t.Fatalf("Record: %v", err)
	}
	if len(repo.events) != 1 {
		t.Fatalf("باید یک ردیف ذخیره شود، %d ذخیره شد", len(repo.events))
	}
	got := repo.events[0]
	if got.UserID == nil || *got.UserID != 7 || got.ActorUsername != "narges" {
		t.Fatalf("فاعل نادرست است: id=%v name=%s", got.UserID, got.ActorUsername)
	}
	if got.Ip != "192.0.2.15" || got.UserAgent != "Reception/1" {
		t.Fatalf("مبدأ نادرست است: ip=%s ua=%s", got.Ip, got.UserAgent)
	}
	if got.EventType != EventRoleCreated || got.Result != ResultSuccess {
		t.Fatalf("نوع یا نتیجه نادرست است: %s %s", got.EventType, got.Result)
	}
	if got.TargetType != TargetRole || got.TargetID != "3" || got.TargetName != "پذیرش" {
		t.Fatalf("هدف نادرست است: %+v", got)
	}
	if got.HashVersion != HashVersionV2 {
		t.Fatalf("رویداد جدید باید هش نسخه ۲ باشد: %d", got.HashVersion)
	}
	if strings.Contains(got.Details, "hidden-secret") || !strings.Contains(got.Details, "***") {
		t.Fatalf("جزئیات ماسک نشده است: %s", got.Details)
	}
	if got.Description == "" || !strings.Contains(got.Description, "نقش") || !strings.Contains(got.Description, "پذیرش") {
		t.Fatalf("توضیح خودکار فارسی ساخته نشد: %s", got.Description)
	}
	if got.RowHash != m.computeRowHash(got) {
		t.Fatal("هش ذخیره‌شده با فیلدهای ردیف نمی‌خواند")
	}
	userID = 99
	if got.UserID == nil || *got.UserID != 7 {
		t.Fatal("شناسه فاعل نباید به متغیر صدازننده وصل بماند")
	}
	lastID, _, broken, verifyErr := m.VerifyChain(AuditChainState{})
	if verifyErr != nil || broken != 0 || lastID != got.ID {
		t.Fatalf("زنجیره نسخه ۲ سالم نیست: id=%d broken=%d err=%v", lastID, broken, verifyErr)
	}
}

// TestRecord_SystemActor فاعل سیستمی شناسه ندارد و نام و IP آن system است.
func TestRecord_SystemActor(t *testing.T) {
	m, repo := newTestManager(t)
	err := m.Record(context.Background(), Event{
		Type:        EventAuditStarted,
		Result:      ResultSuccess,
		Actor:       SystemActor(),
		Description: "آغاز توابع ممیزی پس از migrate",
	})
	if err != nil {
		t.Fatalf("Record: %v", err)
	}
	got := repo.events[0]
	if got.UserID != nil {
		t.Fatalf("فاعل سیستمی نباید شناسه داشته باشد: %d", *got.UserID)
	}
	if got.ActorUsername != "system" || got.Ip != "system" {
		t.Fatalf("فاعل سیستمی نادرست است: name=%s ip=%s", got.ActorUsername, got.Ip)
	}
	if got.EventType != EventAuditStarted || got.HashVersion != HashVersionV2 {
		t.Fatalf("رویداد آغاز ممیزی نادرست است: %+v", got)
	}
	if m.computeRowHash(got) != got.RowHash {
		t.Fatal("هش فاعل سیستمی با ردیف نمی‌خواند")
	}
}

// TestRecord_CancelledContextStoresNothing context لغوشده هیچ ردیفی نمی‌نویسد.
func TestRecord_CancelledContextStoresNothing(t *testing.T) {
	m, repo := newTestManager(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := m.Record(ctx, Event{Type: EventLogout, Result: ResultSuccess, Actor: SystemActor()})
	if err == nil {
		t.Fatal("context لغوشده باید خطا بدهد")
	}
	if len(repo.events) != 0 {
		t.Fatalf("نباید ردیفی ذخیره شود، %d ذخیره شد", len(repo.events))
	}
}

// TestRecordTx_AppendsWithoutNewTransaction رویداد را از همان مسیر زنجیره در حافظه اضافه می‌کند.
func TestRecordTx_AppendsWithoutNewTransaction(t *testing.T) {
	m, repo := newTestManager(t)
	err := m.RecordTx(nil, Event{
		Type:        EventSettingChanged,
		Result:      ResultSuccess,
		Reason:      ReasonSettingsChanged,
		Actor:       SystemActor(),
		Description: "تغییر تنظیم داخل تراکنش",
	})
	if err != nil {
		t.Fatalf("RecordTx: %v", err)
	}
	if len(repo.events) != 1 || repo.events[0].Reason != ReasonSettingsChanged {
		t.Fatalf("ردیف تراکنش ذخیره نشد: %+v", repo.events)
	}
	if repo.events[0].HashVersion != HashVersionV2 || repo.events[0].RowHash == "" {
		t.Fatal("RecordTx باید هش نسخه ۲ بسازد")
	}
}

// TestRecordTx_NilTransactionReturnsError مخزن پایگاه بدون تراکنش خطا می‌دهد.
func TestRecordTx_NilTransactionReturnsError(t *testing.T) {
	m := &Manager{repo: NewRepository(nil), auditKey: []byte("key")}
	err := m.RecordTx(nil, Event{Type: EventLogout, Actor: SystemActor()})
	if err == nil {
		t.Fatal("تراکنش تهی باید خطا بدهد")
	}
	if Unavailable(err) != apperror.ErrAuditUnavailable {
		t.Fatal("خطای تراکنش تهی باید به AUDIT_UNAVAILABLE تبدیل شود")
	}
}

// TestUnavailable_NilStaysNil خطای تهی را به خطای ۵۰۳ تبدیل نمی‌کند.
func TestUnavailable_NilStaysNil(t *testing.T) {
	if Unavailable(nil) != nil {
		t.Fatal("نبود خطا باید nil بماند")
	}
}

// TestHashVersions_MixedChainStillVerifies ردیف نسخه ۱ و نسخه ۲ پشت سر هم یک زنجیره سالم می‌سازند.
func TestHashVersions_MixedChainStillVerifies(t *testing.T) {
	m, repo := newTestManager(t)
	if err := m.LogEvent(nil, "127.0.0.1", EventLoginSuccess, "ردیف قدیمی"); err != nil {
		t.Fatalf("LogEvent: %v", err)
	}
	userID := 2
	if err := m.Record(context.Background(), Event{
		Type:   EventUserUpdated,
		Result: ResultSuccess,
		Actor:  Actor{UserID: &userID, Username: "reza", IP: "198.51.100.4", UserAgent: "UA"},
		Target: Target{Type: TargetUser, ID: "2", Name: "reza"},
	}); err != nil {
		t.Fatalf("Record: %v", err)
	}
	if repo.events[0].HashVersion != HashVersionV1 || repo.events[1].HashVersion != HashVersionV2 {
		t.Fatalf("نسخه هش‌ها باید ۱ سپس ۲ باشد: %d %d", repo.events[0].HashVersion, repo.events[1].HashVersion)
	}
	lastID, _, broken, err := m.VerifyChain(AuditChainState{})
	if err != nil || broken != 0 || lastID != 2 {
		t.Fatalf("زنجیره مخلوط سالم نیست: id=%d broken=%d err=%v", lastID, broken, err)
	}

	repo.events[1].ActorUsername = "دستکاری"
	_, _, broken, err = m.VerifyChain(AuditChainState{})
	if err != nil || broken != 2 {
		t.Fatalf("دستکاری نام در نسخه ۲ باید روی ردیف ۲ بایستد: broken=%d err=%v", broken, err)
	}
	repo.events[1].ActorUsername = "reza"
	repo.events[0].ActorUsername = "نباید نسخه ۱ را خراب کند"
	_, _, broken, err = m.VerifyChain(AuditChainState{})
	if err != nil || broken != 0 {
		t.Fatalf("فیلد جدید روی ردیف نسخه ۱ نباید زنجیره را بشکند: broken=%d err=%v", broken, err)
	}
}

// TestRecord_RejectsUnmarshalableDetails مقدار غیرقابل JSON هیچ ردیفی نمی‌نویسد.
func TestRecord_RejectsUnmarshalableDetails(t *testing.T) {
	m, repo := newTestManager(t)
	err := m.Record(context.Background(), Event{
		Type:    EventUpload,
		Result:  ResultFailure,
		Actor:   SystemActor(),
		Details: make(chan int),
	})
	if err == nil {
		t.Fatal("جزئیات غیرقابل ذخیره باید خطا بدهد")
	}
	if len(repo.events) != 0 {
		t.Fatal("ردیف ناقص نباید ذخیره شود")
	}
}

// TestDescribeEvent_UsesPersianLabel توضیح خالی را از نوع و نام هدف می‌سازد.
func TestDescribeEvent_UsesPersianLabel(t *testing.T) {
	got := describeEvent(Event{
		Type:   EventEntityDeleted,
		Target: Target{Type: TargetPatient, ID: "15", Name: "آزاده"},
	})
	if got != "حذف رکورد؛ بیمار «آزاده» (15)" {
		t.Fatalf("توضیح: %s", got)
	}
}

// TestRecord_DetailsJSONKeepsSafeFields فیلد غیرحساس در JSON جزئیات می‌ماند.
func TestRecord_DetailsJSONKeepsSafeFields(t *testing.T) {
	m, repo := newTestManager(t)
	if err := m.Record(context.Background(), Event{
		Type:        EventEntityUpdated,
		Result:      ResultSuccess,
		Actor:       SystemActor(),
		Description: "به‌روزرسانی",
		Details:     map[string]any{"file_number": "100", "password_hash": "bcrypt"},
	}); err != nil {
		t.Fatalf("Record: %v", err)
	}
	var details map[string]any
	if err := json.Unmarshal([]byte(repo.events[0].Details), &details); err != nil {
		t.Fatalf("JSON جزئیات: %v", err)
	}
	if details["file_number"] != "100" || details["password_hash"] != "***" {
		t.Fatalf("جزئیات: %#v", details)
	}
}
