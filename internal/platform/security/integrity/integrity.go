package integrity

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

// Signer امضای HMAC-SHA256 برای یکپارچگی داده.
type Signer struct {
	key []byte
}

// NewSigner نمونه Signer با کلید جدا از AES می‌سازد.
func NewSigner(key string) *Signer {
	return &Signer{key: []byte(key)}
}

// Sign فیلدهای داده را با HMAC-SHA256 امضا می‌کند.
func (s *Signer) Sign(fields ...string) string {
	mac := hmac.New(sha256.New, s.key)
	mac.Write([]byte(strings.Join(fields, "|")))
	return hex.EncodeToString(mac.Sum(nil))
}

// Verify امضا را با فیلدهای داده مقایسه می‌کند.
func (s *Signer) Verify(hash string, fields ...string) bool {
	expected := s.Sign(fields...)
	return hmac.Equal([]byte(expected), []byte(hash))
}
