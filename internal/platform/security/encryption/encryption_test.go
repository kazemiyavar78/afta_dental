package encryption

import (
	"testing"
)

// TestEncryptor_EncryptDecrypt مسیر خوشحال رمزنگاری AES-256-GCM را بررسی می‌کند.
func TestEncryptor_EncryptDecrypt(t *testing.T) {
	e, err := NewEncryptor("01234567890123456789012345678901")
	if err != nil {
		t.Fatalf("ساخت encryptor: %v", err)
	}

	encrypted, err := e.Encrypt("داده حساس")
	if err != nil {
		t.Fatalf("رمزنگاری: %v", err)
	}

	decrypted, err := e.Decrypt(encrypted)
	if err != nil {
		t.Fatalf("رمزگشایی: %v", err)
	}

	if decrypted != "داده حساس" {
		t.Fatalf("انتظار 'داده حساس'، دریافت '%s'", decrypted)
	}
}
