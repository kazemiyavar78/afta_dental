package encryption


import (
	"fmt"
	"crypto/sha256"
	"encoding/hex"
)

type OrganizationSensitiveData struct {
	Name string
	IsTakmili bool
	IsActive bool
}

// UserEncryptionService سرویس رمزنگاری و SecurityCode کاربر.
type OrganizationEncryptionService struct {
	encryptor *Encryptor
}

// NewUserEncryptionService نمونه UserEncryptionService می‌سازد.
func NewOrganizationEncryptionService(encryptor *Encryptor) *OrganizationEncryptionService {
	return &OrganizationEncryptionService{encryptor: encryptor}
}

func (s *OrganizationEncryptionService) sensitiveFields(data OrganizationSensitiveData) string {
	return fmt.Sprintf("%s|%t|%t",
		data.Name, data.IsTakmili, data.IsActive)
}

// CreateSecurityCode کد امنیتی کاربر را تولید و رمزنگاری می‌کند.
func (s *OrganizationEncryptionService) CreateSecurityCode(data OrganizationSensitiveData) (string, error) {
	hash := sha256.Sum256([]byte(s.sensitiveFields(data)))
	plaintext := hex.EncodeToString(hash[:])
	return s.encryptor.Encrypt(plaintext)
}

// CheckUserSecurityCode یکپارچگی SecurityCode را بررسی می‌کند.
func (s *OrganizationEncryptionService) CheckUserSecurityCode(data OrganizationSensitiveData, securityCode string) bool {
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
func (s *OrganizationEncryptionService) EncryptField(value string) (string, error) {
	if value == "" {
		return "", nil
	}
	return s.encryptor.Encrypt(value)
}

// DecryptField یک فیلد رمزنگاری‌شده را باز می‌گرداند.
func (s *OrganizationEncryptionService) DecryptField(encrypted string) (string, error) {
	if encrypted == "" {
		return "", nil
	}
	return s.encryptor.Decrypt(encrypted)
}
