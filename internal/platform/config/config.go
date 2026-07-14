package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

// Config تنظیمات برنامه که از متغیرهای محیطی خوانده می‌شود.
type Config struct {
	DevMode             bool
	ServerHost          string
	ServerPort          string
	DatabaseURL         string
	AESKey              string
	IntegrityHMACKey    string
	AuditHMACKey        string
	CSRFHMACKey         string
	LogRetentionTicker  time.Duration
	IntegrityTicker     time.Duration
	SecureCookies       bool
	AllowedOrigins      string
	RateLimitPerMinute  int
	LoginRateLimitPerMin int
}

// Load تنظیمات را از متغیرهای محیطی بارگذاری می‌کند.
func Load() (*Config, error) {
	cfg := &Config{
		DevMode:              isDevMode(),
		ServerHost:           getEnv("SERVER_HOST", "0.0.0.0"),
		ServerPort:           getEnv("SERVER_PORT", "8080"),
		DatabaseURL:          getEnv("DATABASE_URL", ""),
		AESKey:               getEnv("AES_KEY", ""),
		IntegrityHMACKey:     getEnv("INTEGRITY_HMAC_KEY", ""),
		AuditHMACKey:         getEnv("AUDIT_HMAC_KEY", ""),
		CSRFHMACKey:          getEnv("CSRF_HMAC_KEY", ""),
		LogRetentionTicker:   getDurationEnv("LOG_RETENTION_TICKER", 15*time.Minute),
		IntegrityTicker:      getDurationEnv("INTEGRITY_TICKER", 15*time.Minute),
		SecureCookies:        getBoolEnv("SECURE_COOKIES", true),
		AllowedOrigins:       getEnv("ALLOWED_ORIGINS", "http://192.168.1.60:5173,http://localhost:5173"),
		RateLimitPerMinute:   getIntEnv("RATE_LIMIT_PER_MINUTE", 120),
		LoginRateLimitPerMin: getIntEnv("LOGIN_RATE_LIMIT_PER_MINUTE", 10),
	}

	if cfg.DatabaseURL == "" {
		return nil, &ConfigError{Field: "DATABASE_URL", Msg: "متغیر DATABASE_URL الزامی است"}
	}
	if cfg.AESKey == "" || cfg.IntegrityHMACKey == "" || cfg.AuditHMACKey == "" || cfg.CSRFHMACKey == "" {
		return nil, &ConfigError{Field: "SECURITY_KEYS", Msg: "کلیدهای امنیتی (AES_KEY, INTEGRITY_HMAC_KEY, AUDIT_HMAC_KEY, CSRF_HMAC_KEY) الزامی هستند"}
	}

	return cfg, nil
}

// ConfigError خطای بارگذاری تنظیمات.
type ConfigError struct {
	Field string
	Msg   string
}

func (e *ConfigError) Error() string {
	return e.Msg
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getIntEnv(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return fallback
}

func getBoolEnv(key string, fallback bool) bool {
	if v := os.Getenv(key); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}
	return fallback
}

func getDurationEnv(key string, fallback time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return fallback
}

// isDevMode بررسی می‌کند برنامه در حالت توسعه اجرا می‌شود یا خیر.
// APP_ENV=development یا DEV_MODE=true حالت توسعه را فعال می‌کند.
func isDevMode() bool {
	if getBoolEnv("DEV_MODE", false) {
		return true
	}
	return strings.EqualFold(getEnv("APP_ENV", "production"), "development")
}
