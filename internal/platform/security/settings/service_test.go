package settings

import (
	"testing"

	"github.com/tpdenta/afta-reception/internal/platform/security/integrity"
)

// TestSigner_SignSettings مسیر خوشحال امضای تنظیمات را بررسی می‌کند.
func TestSigner_SignSettings(t *testing.T) {
	s := integrity.NewSigner("test-integrity-key-32bytes-long!!")
	hash := s.Sign(MaximumNumberOfFailedLogin, "5")
	if !s.Verify(hash, MaximumNumberOfFailedLogin, "5") {
		t.Fatal("امضای تنظیمات باید معتبر باشد")
	}
}
