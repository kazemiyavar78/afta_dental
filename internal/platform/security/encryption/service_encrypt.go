package encryption

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

// ServiceSensitiveData فیلدهای حساس خدمت برای محاسبه هش یکپارچگی.
type ServiceSensitiveData struct {
	ServiceCode             string
	Name                    string
	TechnicalCoefficient    float64
	ProfessionalCoefficient float64
	ConsumptionCoefficient  float64
	ServiceRate             int
	ServiceTariff           int
	InternationalCode       string
	DefaultCount            int
	MaximumCount            int
	ServiceFeatures         string
	IsActive                bool
}

// ServiceEncryptionService سرویس رمزنگاری و IntegrityHash خدمت.
type ServiceEncryptionService struct {
	encryptor *Encryptor
}

// NewServiceEncryptionService نمونه ServiceEncryptionService می‌سازد.
func NewServiceEncryptionService(encryptor *Encryptor) *ServiceEncryptionService {
	return &ServiceEncryptionService{encryptor: encryptor}
}

func (s *ServiceEncryptionService) sensitiveFields(data ServiceSensitiveData) string {
	return fmt.Sprintf("%s|%s|%g|%g|%g|%d|%d|%s|%d|%d|%s|%t",
		data.ServiceCode,
		data.Name,
		data.TechnicalCoefficient,
		data.ProfessionalCoefficient,
		data.ConsumptionCoefficient,
		data.ServiceRate,
		data.ServiceTariff,
		data.InternationalCode,
		data.DefaultCount,
		data.MaximumCount,
		data.ServiceFeatures,
		data.IsActive,
	)
}

// CreateSecurityCode هش امنیتی خدمت را تولید و رمزنگاری می‌کند.
func (s *ServiceEncryptionService) CreateSecurityCode(data ServiceSensitiveData) (string, error) {
	hash := sha256.Sum256([]byte(s.sensitiveFields(data)))
	plaintext := hex.EncodeToString(hash[:])
	return s.encryptor.Encrypt(plaintext)
}

// CheckSecurityCode یکپارچگی IntegrityHash خدمت را بررسی می‌کند.
func (s *ServiceEncryptionService) CheckSecurityCode(data ServiceSensitiveData, securityCode string) bool {
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
