package encryption

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

// OrganizationPackageSensitiveData فیلدهای حساس بسته تعرفه برای هش یکپارچگی.
type OrganizationPackageSensitiveData struct {
	PackageName                      string
	PackageDescription               string
	TechnicalCoefficient             int
	TechnicalProfessionalCoefficient int
	ConsumptionCoefficient           int
	SubsidyPercentage                int
	SupplementaryPercentage          int
	OrganizationPercentage           int
}

// OrganizationPackageEncryptionService سرویس رمزنگاری و IntegrityHash بسته تعرفه.
type OrganizationPackageEncryptionService struct {
	encryptor *Encryptor
}

// NewOrganizationPackageEncryptionService نمونه OrganizationPackageEncryptionService می‌سازد.
func NewOrganizationPackageEncryptionService(encryptor *Encryptor) *OrganizationPackageEncryptionService {
	return &OrganizationPackageEncryptionService{encryptor: encryptor}
}

func (s *OrganizationPackageEncryptionService) sensitiveFields(data OrganizationPackageSensitiveData) string {
	return fmt.Sprintf("%s|%s|%d|%d|%d|%d|%d|%d",
		data.PackageName,
		data.PackageDescription,
		data.TechnicalCoefficient,
		data.TechnicalProfessionalCoefficient,
		data.ConsumptionCoefficient,
		data.SubsidyPercentage,
		data.SupplementaryPercentage,
		data.OrganizationPercentage,
	)
}

// CreateSecurityCode هش امنیتی بسته تعرفه را تولید و رمزنگاری می‌کند.
func (s *OrganizationPackageEncryptionService) CreateSecurityCode(data OrganizationPackageSensitiveData) (string, error) {
	hash := sha256.Sum256([]byte(s.sensitiveFields(data)))
	plaintext := hex.EncodeToString(hash[:])
	return s.encryptor.Encrypt(plaintext)
}

// CheckSecurityCode یکپارچگی IntegrityHash بسته تعرفه را بررسی می‌کند.
func (s *OrganizationPackageEncryptionService) CheckSecurityCode(data OrganizationPackageSensitiveData, securityCode string) bool {
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
