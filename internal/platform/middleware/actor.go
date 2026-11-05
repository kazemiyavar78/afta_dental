package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/tpdenta/afta-reception/internal/platform/security/audit"
)

// ActorFromContext فاعل ممیزی را از context درخواست می‌سازد.
// c درخواست جاری است. خروجی شناسه، نام کاربری، IP و User-Agent است.
// اگر کاربر در context نباشد UserID تهی است و اشاره‌گر به صفر ساخته نمی‌شود.
// FAU_GEN.2 / D-14: ردیف ممیزی نباید کاربر ناشناس را با شناسه صفر ذخیره کند.
func ActorFromContext(c *gin.Context) audit.Actor {
	actor := audit.Actor{}
	if c == nil {
		return actor
	}
	if c.Request != nil {
		actor.IP = c.ClientIP()
		actor.UserAgent = c.Request.UserAgent()
	}
	raw, exists := c.Get(ContextKeyUserID)
	if exists {
		if uid, ok := raw.(int); ok {
			id := uid
			actor.UserID = &id
		}
	}
	name, exists := c.Get(ContextKeyUsername)
	if exists {
		if username, ok := name.(string); ok {
			actor.Username = username
		}
	}
	return actor
}
