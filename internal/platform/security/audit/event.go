package audit

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/tpdenta/afta-reception/internal/platform/apperror"
	"gorm.io/gorm"
)

// ستون‌های nvarchar. برش قبل از درج جلوی خطای طول پایگاه را می‌گیرد.
const (
	maxActorUsername = 100
	maxTargetType    = 50
	maxTargetID      = 64
	maxTargetName    = 200
	maxResult        = 20
	maxReason        = 50
	maxUserAgent     = 512
	maxIP            = 45
	maxEventType     = 100
)

// Actor کسی است که رویداد را انجام داده. نام کاربری در همان لحظه ذخیره می‌شود.
// UserID تهی یعنی فاعل انسانی در context نبوده یا کار سیستمی است.
type Actor struct {
	UserID    *int
	Username  string
	IP        string
	UserAgent string
}

// Target موجودی است که رویداد روی آن انجام شده.
// Type یکی از مقدارهای user، role، session، setting، patient و مانند آن‌هاست.
type Target struct {
	Type string
	ID   string
	Name string
}

// Event ورودی ثبت مرکزی ممیزی است.
// Description اگر خالی باشد از نوع و هدف به فارسی ساخته می‌شود.
// Details قبل از ذخیره ماسک و به JSON تبدیل می‌شود.
type Event struct {
	Type        string
	Result      string
	Reason      string
	Actor       Actor
	Target      Target
	Details     any
	Description string
}

// SystemActor فاعل worker، راه‌اندازی و خاموشی است.
// خروجی شناسه کاربر ندارد. نام و IP هر دو system هستند تا ردیف بدون join مبدأ را نشان دهد.
func SystemActor() Actor {
	return Actor{Username: "system", IP: "system"}
}

// Unavailable خطای ثبت ممیزی را به خطای عمومی عملیات تبدیل می‌کند.
// err نتیجه Record یا RecordTx است. nil همان nil می‌ماند.
// هر خطای دیگر AUDIT_UNAVAILABLE با وضعیت ۵۰۳ است تا جزئیات پایگاه به کاربر نرسد.
// FAU_STG.4 / D-13: سرویس عملیات تغییردهنده بعد از شکست ثبت باید این را برگرداند.
func Unavailable(err error) error {
	if err == nil {
		return nil
	}
	return apperror.ErrAuditUnavailable
}

// Record رویداد را با هش نسخه ۲ به انتهای زنجیره اضافه می‌کند.
// ctx لغو درخواست است و قبل از درج بررسی می‌شود. e فاعل، هدف، نتیجه و دلیل است.
// خطا یعنی ردیف نوشته نشده است. قفل فرایند و UPDLOCK جلوی PrevHash تکراری را می‌گیرند.
// FAU_GEN.2: نام کاربری فاعل در خود ردیف می‌ماند، حتی اگر کاربر بعداً حذف یا عوض شود.
func (m *Manager) Record(ctx context.Context, e Event) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	row, err := m.prepareRow(e)
	if err != nil {
		return err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.repo.AppendChainedCtx(ctx, func(prevHash string) (*SecurityEvent, error) {
		sealed := m.seal(row, prevHash)
		return &sealed, nil
	})
}

// RecordTx رویداد را داخل تراکنش عملیات درج می‌کند و تراکنش تازه باز نمی‌کند.
// tx تراکنش باز سرویس است. e همان ورودی Record است. خطا باید تراکنش را برگرداند.
// FAU_STG.4: ممیزی و تغییر داده با هم commit می‌شوند. صدازننده نباید پیش از commit دوباره Record بزند.
func (m *Manager) RecordTx(tx *gorm.DB, e Event) error {
	row, err := m.prepareRow(e)
	if err != nil {
		return err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.repo.AppendInTx(tx, func(prevHash string) (*SecurityEvent, error) {
		sealed := m.seal(row, prevHash)
		return &sealed, nil
	})
}

// prepareRow فیلدهای قابل ذخیره را می‌سازد و هنوز هش نمی‌زند.
// e ورودی سرویس است. خروجی ردیف بدون زمان و هش است، یا خطای JSON جزئیات.
func (m *Manager) prepareRow(e Event) (SecurityEvent, error) {
	details, err := marshalDetails(e.Details)
	if err != nil {
		return SecurityEvent{}, err
	}
	description := strings.TrimSpace(e.Description)
	if description == "" {
		description = describeEvent(e)
	}
	return SecurityEvent{
		UserID:        copyUserID(e.Actor.UserID),
		ActorUsername: clip(e.Actor.Username, maxActorUsername),
		Ip:            clip(e.Actor.IP, maxIP),
		UserAgent:     clip(e.Actor.UserAgent, maxUserAgent),
		EventType:     clip(e.Type, maxEventType),
		Result:        clip(e.Result, maxResult),
		Reason:        clip(e.Reason, maxReason),
		TargetType:    clip(e.Target.Type, maxTargetType),
		TargetID:      clip(e.Target.ID, maxTargetID),
		TargetName:    clip(e.Target.Name, maxTargetName),
		Description:   description,
		Details:       details,
		HashVersion:   HashVersionV2,
	}, nil
}

// seal زمان، هش قبلی و هش ردیف را روی نسخه آماده می‌نویسد.
// row خروجی prepareRow است. prevHash هش ردیف قبلی یا لنگر است. خروجی آماده درج است.
func (m *Manager) seal(row SecurityEvent, prevHash string) SecurityEvent {
	row.CreatedAt = time.Now().UTC().Truncate(time.Microsecond)
	row.PrevHash = prevHash
	row.RowHash = m.computeRowHash(row)
	return row
}

// marshalDetails جزئیات را ماسک می‌کند و JSON رشته‌ای می‌سازد.
// details تهی یعنی ستون خالی. خطا یعنی مقدار قابل ذخیره نیست و رویداد نباید درج شود.
func marshalDetails(details any) (string, error) {
	if details == nil {
		return "", nil
	}
	raw, err := json.Marshal(MaskSensitive(details))
	if err != nil {
		return "", err
	}
	return string(raw), nil
}

// describeEvent توضیح فارسی را از برچسب نوع و هدف می‌سازد.
// e رویداد بدون توضیح است. خروجی یک جمله کوتاه برای ستون Description است.
func describeEvent(e Event) string {
	label := EventLabel(e.Type)
	target := describeTarget(e.Target)
	if target == "" {
		return label
	}
	return label + "؛ " + target
}

// describeTarget عبارت فارسی هدف را می‌سازد.
// t نوع، شناسه و نام است. خروجی خالی یعنی هدفی داده نشده.
func describeTarget(t Target) string {
	if t.Type == "" && t.ID == "" && t.Name == "" {
		return ""
	}
	kind := t.Type
	if label, ok := targetLabels[t.Type]; ok {
		kind = label
	}
	switch {
	case t.Name != "" && t.ID != "":
		return kind + " «" + t.Name + "» (" + t.ID + ")"
	case t.Name != "":
		return kind + " «" + t.Name + "»"
	case t.ID != "":
		if kind == "" {
			return t.ID
		}
		return kind + " " + t.ID
	default:
		return kind
	}
}

// copyUserID یک کپی از شناسه می‌سازد تا تغییر بعدی صدازننده ردیف را عوض نکند.
// userID می‌تواند nil باشد. خروجی nil همان فاعل بدون شناسه است.
func copyUserID(userID *int) *int {
	if userID == nil {
		return nil
	}
	id := *userID
	return &id
}

// clip رشته را به حداکثر تعداد نویسه یونیکد محدود می‌کند.
// s متن ورودی است و max سقف ستون است. برش با بایت انجام نمی‌شود تا نویسه فارسی نصف نشود.
func clip(s string, max int) string {
	if max <= 0 || s == "" {
		return s
	}
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	return string(runes[:max])
}
