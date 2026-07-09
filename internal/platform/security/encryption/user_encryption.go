package encryption

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

// SensitiveUserData داده‌های حساس کاربر برای SecurityCode (بدون وابستگی به ماژول user).
type SensitiveUserData struct {
	Name          string
	Family        string
	PhoneNumber   string
	PasswordHash  string
	Address       string
	RoleID        int
	MedicalCode   string
}

// UserEncryptionService سرویس رمزنگاری و SecurityCode کاربر.
type UserEncryptionService struct {
	encryptor *Encryptor
}

// NewUserEncryptionService نمونه UserEncryptionService می‌سازد.
func NewUserEncryptionService(encryptor *Encryptor) *UserEncryptionService {
	return &UserEncryptionService{encryptor: encryptor}
}

func (s *UserEncryptionService) sensitiveFields(data SensitiveUserData) string {
	return fmt.Sprintf("%s|%s|%s|%s|%s|%d|%s",
		data.Name, data.Family, data.PhoneNumber, data.PasswordHash, data.Address, data.RoleID, data.MedicalCode)
}

// CreateSecurityCode کد امنیتی کاربر را تولید و رمزنگاری می‌کند.
func (s *UserEncryptionService) CreateSecurityCode(data SensitiveUserData) (string, error) {
	hash := sha256.Sum256([]byte(s.sensitiveFields(data)))
	plaintext := hex.EncodeToString(hash[:])
	return s.encryptor.Encrypt(plaintext)
}

// CheckUserSecurityCode یکپارچگی SecurityCode را بررسی می‌کند.
func (s *UserEncryptionService) CheckUserSecurityCode(data SensitiveUserData, securityCode string) bool {
	if securityCode == "" {
		return false
	}

	decrypted, err := s.encryptor.Decrypt(securityCode)
	if err != nil {
		return false
	}

	hash := sha256.Sum256([]byte(s.sensitiveFields(data)))
	expected := hex.EncodeToString(hash[:])
	return decrypted == expected
}

// EncryptField یک فیلد حساس را رمزنگاری می‌کند.
func (s *UserEncryptionService) EncryptField(value string) (string, error) {
	if value == "" {
		return "", nil
	}
	return s.encryptor.Encrypt(value)
}

// DecryptField یک فیلد رمزنگاری‌شده را باز می‌گرداند.
func (s *UserEncryptionService) DecryptField(encrypted string) (string, error) {
	if encrypted == "" {
		return "", nil
	}
	return s.encryptor.Decrypt(encrypted)
}
