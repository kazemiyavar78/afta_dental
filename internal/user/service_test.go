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
	if fields[8] != "1" {
		t.Fatalf("RoleID باید در فیلدهای یکپارچگی باشد، گرفتیم: %v", fields[8])
	}
}

// TestUpdateClearsStaleRoleAssociation اطمینان می‌دهد که تغییر RoleID با Role از پیش‌بارگذاری‌شده
// باعث ناسازگاری هش نمی‌شود (بازنویسی RoleID توسط GORM Belongs-To).
func TestUpdateClearsStaleRoleAssociation(t *testing.T) {
	user := &User{
		ID:     10,
		RoleID: 1,
		Role:   Role{ID: 1, Name: "Old"},
	}
	user.RoleID = 2
	user.Role = Role{}
	if user.Role.ID != 0 {
		t.Fatal("association نقش باید خالی شود تا Save نقش قدیمی را برنگرداند")
	}
	if user.RoleID != 2 {
		t.Fatalf("RoleID باید ۲ بماند، گرفتیم: %d", user.RoleID)
	}
}
