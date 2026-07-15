package encryption

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

// OrganizationSensitiveData فیلدهای حساس سازمان برای محاسبه هش یکپارچگی.
type OrganizationSensitiveData struct {
	Name      string
	IsTakmili bool
	IsActive  bool
	PackageID uint
}

// OrganizationEncryptionService سرویس رمزنگاری و IntegrityHash سازمان.
type OrganizationEncryptionService struct {
	encryptor *Encryptor
}

// NewOrganizationEncryptionService نمونه OrganizationEncryptionService می‌سازد.
func NewOrganizationEncryptionService(encryptor *Encryptor) *OrganizationEncryptionService {
	return &OrganizationEncryptionService{encryptor: encryptor}
}

func (s *OrganizationEncryptionService) sensitiveFields(data OrganizationSensitiveData) string {
	return fmt.Sprintf("%s|%t|%t|%d",
		data.Name, data.IsTakmili, data.IsActive, data.PackageID)
}

// CreateSecurityCode هش امنیتی سازمان را تولید و رمزنگاری می‌کند.
func (s *OrganizationEncryptionService) CreateSecurityCode(data OrganizationSensitiveData) (string, error) {
	hash := sha256.Sum256([]byte(s.sensitiveFields(data)))
	plaintext := hex.EncodeToString(hash[:])
	return s.encryptor.Encrypt(plaintext)
}

// CheckUserSecurityCode یکپارچگی IntegrityHash سازمان را بررسی می‌کند.
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
