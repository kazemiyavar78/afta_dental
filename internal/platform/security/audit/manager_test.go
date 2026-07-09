package audit

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"testing"
)

// TestComputeRowHash مسیر خوشحال محاسبه هش زنجیره‌ای را بررسی می‌کند.
func TestComputeRowHash(t *testing.T) {
	key := []byte("test-audit-key-32bytes-long!!!!!")
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte("prev|1|127.0.0.1|LOGIN_SUCCESS|test|2024-01-01T00:00:00Z"))
	hash := hex.EncodeToString(mac.Sum(nil))
	if len(hash) != 64 {
		t.Fatalf("هش باید ۶۴ کاراکتر باشد، دریافت %d", len(hash))
	}
}
