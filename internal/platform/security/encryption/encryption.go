package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
)

// Encryptor رمزنگاری AES-256-GCM برای داده‌های حساس.
type Encryptor struct {
	gcm cipher.AEAD
}

// NewEncryptor نمونه Encryptor با کلید ۳۲ بایتی می‌سازد.
func NewEncryptor(key string) (*Encryptor, error) {
	keyBytes := []byte(key)
	if len(keyBytes) != 32 {
		return nil, fmt.Errorf("کلید AES باید دقیقاً ۳۲ بایت باشد، طول فعلی: %d", len(keyBytes))
	}

	block, err := aes.NewCipher(keyBytes)
	if err != nil {
		return nil, fmt.Errorf("ساخت cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("ساخت GCM: %w", err)
	}

	return &Encryptor{gcm: gcm}, nil
}

// Encrypt متن را با AES-256-GCM رمزنگاری می‌کند.
func (e *Encryptor) Encrypt(plaintext string) (string, error) {
	nonce := make([]byte, e.gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("تولید nonce: %w", err)
	}

	ciphertext := e.gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt متن رمزنگاری‌شده را باز می‌گرداند.
func (e *Encryptor) Decrypt(ciphertext string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", fmt.Errorf("decode base64: %w", err)
	}

	nonceSize := e.gcm.NonceSize()
	if len(data) < nonceSize {
		return "", fmt.Errorf("داده رمزنگاری‌شده کوتاه است")
	}

	nonce, cipherData := data[:nonceSize], data[nonceSize:]
	plaintext, err := e.gcm.Open(nil, nonce, cipherData, nil)
	if err != nil {
		return "", fmt.Errorf("رمزگشایی: %w", err)
	}

	return string(plaintext), nil
}
