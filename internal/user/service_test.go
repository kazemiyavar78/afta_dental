package user

import (
	"testing"

	"golang.org/x/crypto/bcrypt"
)

// TestBcryptHash مسیر خوشحال هش bcrypt را بررسی می‌کند.
func TestBcryptHash(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("Test@1234"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("خطا در هش: %v", err)
	}
	if err := bcrypt.CompareHashAndPassword(hash, []byte("Test@1234")); err != nil {
		t.Fatal("رمز باید با هش مطابقت داشته باشد")
	}
}

// TestBuildUserIntegrityFields فیلدهای یکپارچگی را بررسی می‌کند.
func TestBuildUserIntegrityFields(t *testing.T) {
	u := &User{
		ID:           1,
		Username:     "test",
		PasswordHash: "hash",
		Name:         "نام",
		Family:       "خانوادگی",
		RoleID:       1,
		IsActive:     true,
	}
	fields := BuildUserIntegrityFields(u)
	if len(fields) < 5 {
		t.Fatal("فیلدهای یکپارچگی ناکافی")
	}
}
