package audit

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"strconv"
	"strings"
	"time"
)

// computeRowHash هش یک ردیف را با HMAC-SHA256 و کلید AUDIT_HMAC_KEY می‌سازد.
// event فیلدهای همان ردیف، از جمله PrevHash و HashVersion، است.
// خروجی هگز ۶۴ نویسه‌ای است. نسخه ۰ و ۱ فرمول قدیمی را نگه می‌دارند تا ردیف قبلی هنوز تأیید شود.
// FAU_STG.1: تغییر هر فیلد داخل payload همان نسخه، RowHash را عوض می‌کند.
func (m *Manager) computeRowHash(event SecurityEvent) string {
	version := event.HashVersion
	if version == 0 || version == HashVersionV1 {
		return hashPayload(m.auditKey, hashPayloadV1(event))
	}
	return hashPayload(m.auditKey, hashPayloadV2(event))
}

// hashPayloadV1 رشته نسخه ۱ است: هش قبلی، شناسه کاربر، IP، نوع، توضیح و زمان.
// event ردیف است. فیلدهای فاعل و هدف عمداً وارد نمی‌شوند.
func hashPayloadV1(event SecurityEvent) string {
	return strings.Join([]string{
		event.PrevHash,
		formatUserID(event.UserID),
		event.Ip,
		event.EventType,
		event.Description,
		formatHashTime(event.CreatedAt),
	}, "|")
}

// hashPayloadV2 رشته نسخه ۲ است و فیلدهای جدید را هم به ترتیب ثابت شامل می‌شود.
// event ردیف است. شماره نسخه داخل payload است تا نسخه‌های بعدی با همین فیلدها اشتباه گرفته نشوند.
func hashPayloadV2(event SecurityEvent) string {
	return strings.Join([]string{
		event.PrevHash,
		strconv.Itoa(int(event.HashVersion)),
		formatUserID(event.UserID),
		event.ActorUsername,
		event.Ip,
		event.UserAgent,
		event.EventType,
		event.Result,
		event.Reason,
		event.TargetType,
		event.TargetID,
		event.TargetName,
		event.Description,
		event.Details,
		formatHashTime(event.CreatedAt),
	}, "|")
}

// hashPayload HMAC-SHA256 رشته را با کلید زنجیره به هگز تبدیل می‌کند.
// key کلید AUDIT_HMAC_KEY است و payload متن کاننیکال ردیف است.
func hashPayload(key []byte, payload string) string {
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(payload))
	return hex.EncodeToString(mac.Sum(nil))
}

// formatUserID شناسه فاعل را برای هش رشته می‌کند.
// userID تهی یعنی بخش کاربر خالی بماند، از جمله فاعل سیستمی.
func formatUserID(userID *int) string {
	if userID == nil {
		return ""
	}
	return strconv.Itoa(*userID)
}

// formatHashTime زمان را به UTC و RFC3339Nano تبدیل می‌کند.
// createdAt همان لحظه‌ای است که در ردیف ذخیره شده. منطقه زمانی در رشته هش نمی‌ماند.
func formatHashTime(createdAt time.Time) string {
	return createdAt.UTC().Format(time.RFC3339Nano)
}
