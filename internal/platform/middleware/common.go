package middleware

import (
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tpdenta/afta-reception/internal/platform/apperror"
)

var sensitivePattern = regexp.MustCompile(`(?i)(password|token|secret|authorization|csrf)`)

// RecoveryMiddleware بازیابی از panic بدون افشای stack trace.
func RecoveryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("panic recovered: %v", r)
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"error": apperror.ErrInternal.UserMsg,
					"code":  apperror.ErrInternal.Code,
				})
			}
		}()
		c.Next()
	}
}

// SecurityHeadersMiddleware هدرهای امنیتی HTTP را تنظیم می‌کند.
func SecurityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("Content-Security-Policy", "default-src 'self'")
		c.Header("Referrer-Policy", "no-referrer")
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		c.Next()
	}
}

// RequestLoggerMiddleware لاگ ساختاریافته بدون داده حساس.
func RequestLoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()

		if sensitivePattern.MatchString(path) {
			log.Printf("[REQUEST] %s %s %d %v [REDACTED]", c.Request.Method, path, status, latency)
		} else {
			log.Printf("[REQUEST] %s %s %d %v", c.Request.Method, path, status, latency)
		}
	}
}

// CORSMiddleware تنظیم CORS برای فرانت React.
func CORSMiddleware(allowedOrigins string) gin.HandlerFunc {
	origins := strings.Split(allowedOrigins, ",")
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		for _, o := range origins {
			if strings.TrimSpace(o) == origin {
				c.Header("Access-Control-Allow-Origin", origin)
				break
			}
		}
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Headers", "Content-Type, X-CSRF-Token, Authorization")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

// WriteError پاسخ خطای یکدست می‌نویسد.
func WriteError(c *gin.Context, err error) {
	if appErr, ok := err.(*apperror.AppError); ok {
		c.JSON(appErr.HTTPStatus, gin.H{
			"error": appErr.UserMsg,
			"code":  appErr.Code,
		})
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{
		"error": apperror.ErrInternal.UserMsg,
		"code":  apperror.ErrInternal.Code,
	})
}
