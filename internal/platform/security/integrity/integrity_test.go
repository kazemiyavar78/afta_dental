package integrity

import (
	"testing"
)

// TestSigner_SignVerify مسیر خوشحال امضا و تأیید HMAC را بررسی می‌کند.
func TestSigner_SignVerify(t *testing.T) {
	s := NewSigner("test-integrity-key-32bytes-long!!")
	hash := s.Sign("field1", "field2", "field3")
	if !s.Verify(hash, "field1", "field2", "field3") {
		t.Fatal("امضا باید معتبر باشد")
	}
	if s.Verify(hash, "field1", "field2", "tampered") {
		t.Fatal("امضای دستکاری‌شده نباید معتبر باشد")
	}
}
