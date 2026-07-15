package encryption

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

// PatientSensitiveData فیلدهای حساس بیمار برای محاسبه هش یکپارچگی.
type PatientSensitiveData struct {
	FirstName         string
	LastName          string
	NationalCode      string
	BirthDate         string
	Address           string
	HomePhoneNumber   string
	MobilePhoneNumber string
	FileNumber        string
	Sex               bool
}

// PatientEncryptionService سرویس رمزنگاری و IntegrityHash بیمار.
type PatientEncryptionService struct {
	encryptor *Encryptor
}

// NewPatientEncryptionService نمونه PatientEncryptionService می‌سازد.
func NewPatientEncryptionService(encryptor *Encryptor) *PatientEncryptionService {
	return &PatientEncryptionService{encryptor: encryptor}
}

// sensitiveFields رشته یکپارچه‌سازی‌شده فیلدهای حساس بیمار را برمی‌گرداند.
func (s *PatientEncryptionService) sensitiveFields(data PatientSensitiveData) string {
	return fmt.Sprintf("%s|%s|%s|%s|%s|%s|%s|%s|%t",
		data.FirstName,
		data.LastName,
		data.NationalCode,
		data.BirthDate,
		data.Address,
		data.HomePhoneNumber,
		data.MobilePhoneNumber,
		data.FileNumber,
		data.Sex,
	)
}

// CreateSecurityCode هش امنیتی بیمار را تولید و رمزنگاری می‌کند.
func (s *PatientEncryptionService) CreateSecurityCode(data PatientSensitiveData) (string, error) {
	hash := sha256.Sum256([]byte(s.sensitiveFields(data)))
	plaintext := hex.EncodeToString(hash[:])
	return s.encryptor.Encrypt(plaintext)
}

// CheckSecurityCode یکپارچگی IntegrityHash بیمار را بررسی می‌کند.
func (s *PatientEncryptionService) CheckSecurityCode(data PatientSensitiveData, securityCode string) bool {
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
