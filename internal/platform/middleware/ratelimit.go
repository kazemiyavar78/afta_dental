package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tpdenta/afta-reception/internal/platform/apperror"
	"github.com/tpdenta/afta-reception/internal/platform/security/loginguard"
)

// rateLimitEntry ورودی شمارنده نرخ درخواست.
type rateLimitEntry struct {
	count     int
	windowEnd time.Time
}

// RateLimiterMiddleware محدودیت نرخ سراسری بر اساس IP.
func RateLimiterMiddleware(maxPerMinute int) gin.HandlerFunc {
	var mu sync.Mutex
	entries := make(map[string]*rateLimitEntry)

	return func(c *gin.Context) {
		ip := c.ClientIP()
		now := time.Now()

		mu.Lock()
		entry, ok := entries[ip]
		if !ok || now.After(entry.windowEnd) {
			entries[ip] = &rateLimitEntry{count: 1, windowEnd: now.Add(time.Minute)}
			mu.Unlock()
			c.Next()
			return
		}

		entry.count++
		if entry.count > maxPerMinute {
			mu.Unlock()
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "تعداد درخواست‌ها بیش از حد مجاز است.",
				"code":  "RATE_LIMIT",
			})
			return
		}
		mu.Unlock()
		c.Next()
	}
}

// LoginRateLimiterMiddleware محدودیت سخت‌گیرانه‌تر برای /api/login با استفاده از LoginAttempts.
func LoginRateLimiterMiddleware(guard *loginguard.Guard) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		if err := guard.CheckIPLock(ip); err != nil {
			WriteError(c, err)
			c.Abort()
			return
		}
		c.Next()
	}
}

// SkipRateLimitForLogin علامت‌گذاری مسیر لاگین (برای استفاده در router).
func IsLoginPath(c *gin.Context) bool {
	return c.Request.URL.Path == "/api/login"
}

// HandleRateLimitError خطای rate limit را مدیریت می‌کند.
func HandleRateLimitError(c *gin.Context) {
	c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
		"error": apperror.ErrIPLocked.UserMsg,
		"code":  "RATE_LIMIT",
	})
}
