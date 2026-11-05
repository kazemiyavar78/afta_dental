package audit

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"strconv"
	"strings"
	"testing"
	"time"
)

// TestComputeRowHash خروجی Manager.computeRowHash را با HMAC دستی مقایسه می‌کند.
func TestComputeRowHash(t *testing.T) {
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
			got := m.computeRowHash("prev", tc.userID, "127.0.0.1", "LOGIN_SUCCESS", "test", createdAt)
			want := manualRowHash(key, "prev", tc.userID, "127.0.0.1", "LOGIN_SUCCESS", "test", createdAt)
			if got != want {
				t.Fatalf("هش زنجیره با محاسبه دستی یکی نیست: got %s want %s", got, want)
			}
			if len(got) != 64 {
				t.Fatalf("هش باید ۶۴ نویسه باشد، دریافت %d", len(got))
			}
		})
	}
}

// manualRowHash همان رشته و HMAC را جدا از Manager می‌سازد تا انتظار تست مستقل باشد.
func manualRowHash(key []byte, prevHash string, userID *int, ip, eventType, description string, createdAt time.Time) string {
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
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(payload))
	return hex.EncodeToString(mac.Sum(nil))
}
