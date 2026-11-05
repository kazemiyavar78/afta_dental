package audit

import (
	"testing"
	"time"
)

// TestComputeRowHash_V1MatchesManualHMAC هش نسخه ۱ را با HMAC جداگانه مقایسه می‌کند.
func TestComputeRowHash_V1MatchesManualHMAC(t *testing.T) {
	key := []byte("test-audit-key-32bytes-long!!!!!")
	m := &Manager{auditKey: key}
	tehran := time.FixedZone("UTC+3:30", 3*3600+30*60)
	createdAt := time.Date(2024, 1, 1, 3, 30, 0, 123456789, tehran)
	userID := 1

	cases := []struct {
		name   string
		userID *int
	}{
		{name: "with user", userID: &userID},
		{name: "without user", userID: nil},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			event := SecurityEvent{
				PrevHash:      "prev",
				UserID:        tc.userID,
				Ip:            "127.0.0.1",
				EventType:     "LOGIN_SUCCESS",
				Description:   "test",
				CreatedAt:     createdAt,
				HashVersion:   HashVersionV1,
				ActorUsername: "باید در نسخه ۱ بی‌اثر باشد",
				Details:       `{"password":"secret"}`,
			}
			got := m.computeRowHash(event)
			want := hashPayload(key, hashPayloadV1(event))
			if got != want {
				t.Fatalf("هش نسخه ۱ با محاسبه دستی یکی نیست: got %s want %s", got, want)
			}
			if len(got) != 64 {
				t.Fatalf("هش باید ۶۴ نویسه باشد، دریافت %d", len(got))
			}
			changed := event
			changed.ActorUsername = "دیگر"
			if m.computeRowHash(changed) != got {
				t.Fatal("هش نسخه ۱ نباید با نام کاربری عوض شود")
			}
		})
	}
}

// TestComputeRowHash_V2IncludesNewFields هش نسخه ۲ را با payload کامل مقایسه می‌کند.
func TestComputeRowHash_V2IncludesNewFields(t *testing.T) {
	key := []byte("test-audit-key-32bytes-long!!!!!")
	m := &Manager{auditKey: key}
	createdAt := time.Date(2026, 11, 5, 20, 0, 0, 0, time.UTC)
	userID := 4
	event := SecurityEvent{
		PrevHash:      "prev",
		UserID:        &userID,
		ActorUsername: "narges",
		Ip:            "192.0.2.10",
		UserAgent:     "UnitTest",
		EventType:     EventRoleCreated,
		Result:        ResultSuccess,
		Reason:        "",
		TargetType:    TargetRole,
		TargetID:      "8",
		TargetName:    "پذیرش",
		Description:   "ایجاد نقش",
		Details:       `{"name":"پذیرش"}`,
		CreatedAt:     createdAt,
		HashVersion:   HashVersionV2,
	}
	got := m.computeRowHash(event)
	want := hashPayload(key, hashPayloadV2(event))
	if got != want || len(got) != 64 {
		t.Fatalf("هش نسخه ۲ نادرست است: %s", got)
	}
	changed := event
	changed.ActorUsername = "other"
	if m.computeRowHash(changed) == got {
		t.Fatal("تغییر نام کاربری باید هش نسخه ۲ را عوض کند")
	}
	v1 := event
	v1.HashVersion = HashVersionV1
	if m.computeRowHash(v1) == got {
		t.Fatal("هش نسخه ۱ و ۲ یک ردیف نباید یکی باشد")
	}
}

// TestComputeRowHash_ZeroVersionUsesV1 ردیف بدون شماره نسخه مثل ردیف قدیمی نسخه ۱ تأیید می‌شود.
func TestComputeRowHash_ZeroVersionUsesV1(t *testing.T) {
	m := &Manager{auditKey: []byte("test-audit-key-32bytes-long!!!!!")}
	event := SecurityEvent{
		PrevHash:    "prev",
		Ip:          "127.0.0.1",
		EventType:   EventLoginSuccess,
		Description: "old",
		CreatedAt:   time.Date(2020, 1, 2, 3, 4, 5, 0, time.UTC),
	}
	asV1 := event
	asV1.HashVersion = HashVersionV1
	if m.computeRowHash(event) != m.computeRowHash(asV1) {
		t.Fatal("نسخه صفر باید همان هش نسخه ۱ را بسازد")
	}
}
