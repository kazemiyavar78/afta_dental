package session

import (
	"context"
	"strconv"

	"github.com/tpdenta/afta-reception/internal/platform/apperror"
	"github.com/tpdenta/afta-reception/internal/platform/security/audit"
)

// recordSessionEnd پایان نشست و خروج صاحب آن را با یک دلیل ثبت می‌کند.
// ctx لغو درخواست است. actor کسی است که نشست را بسته، یا فاعل سیستمی.
// sess نشستِ در حال حذف است. reason یکی از کدهای Reason است.
// خروجی خطای ممیزی یا خطای خواندن نام صاحب است. شناسه کامل نشست در توضیح نمی‌آید.
// FTA_SSL / R1-07: حذف نشست و خروج دو رویداد جدا هستند.
func (s *Service) recordSessionEnd(ctx context.Context, actor audit.Actor, sess *Session, reason string) error {
	if sess == nil {
		return nil
	}
	name, err := s.accountName(sess.PersonnelAccountID)
	if err != nil {
		return err
	}
	target := audit.Target{
		Type: audit.TargetUser,
		ID:   strconv.Itoa(sess.PersonnelAccountID),
		Name: name,
	}
	details := map[string]any{"session_prefix": sessionPrefix(sess)}
	term := audit.Event{
		Type:        audit.EventSessionTerminated,
		Result:      audit.ResultSuccess,
		Reason:      reason,
		Actor:       actor,
		Target:      target,
		Details:     details,
		Description: sessionEndText("پایان نشست", name, reason),
	}
	logout := audit.Event{
		Type:        audit.EventLogout,
		Result:      audit.ResultSuccess,
		Reason:      reason,
		Actor:       logoutActor(actor, sess, name),
		Target:      target,
		Details:     details,
		Description: sessionEndText("خروج", name, reason),
	}
	if err := s.record(ctx, term); err != nil {
		return err
	}
	return s.record(ctx, logout)
}

// accountName نام کاربری صاحب نشست را از تابع تزریق‌شده می‌خواند.
// userID شناسه صاحب است. اگر تابع تنظیم نشده باشد نام خالی و بدون خطا برمی‌گردد.
func (s *Service) accountName(userID int) (string, error) {
	if s.namer == nil {
		return "", nil
	}
	return s.namer(userID)
}

// record یک رویداد را می‌نویسد و خطای پایگاه را به خطای عمومی ممیزی تبدیل می‌کند.
// ctx زمینه درخواست است. e رویداد آماده است. خروجی nil یعنی ثبت شد.
func (s *Service) record(ctx context.Context, e audit.Event) error {
	if s.audit == nil {
		return audit.Unavailable(apperror.ErrInternal)
	}
	return audit.Unavailable(s.audit.Record(ctx, e))
}

// logoutActor فاعل رویداد خروج را صاحب نشست می‌گذارد.
// closer کسی است که نشست را بسته. sess نشست و ownerName نام صاحب است.
// اگر خود صاحب بسته باشد، IP و نام همان درخواست حفظ می‌شود.
func logoutActor(closer audit.Actor, sess *Session, ownerName string) audit.Actor {
	id := sess.PersonnelAccountID
	owner := audit.Actor{
		UserID:    &id,
		Username:  ownerName,
		IP:        sess.Ip,
		UserAgent: sess.Browser,
	}
	if closer.UserID != nil && *closer.UserID == id {
		if closer.Username != "" {
			owner.Username = closer.Username
		}
		if closer.IP != "" {
			owner.IP = closer.IP
		}
		if closer.UserAgent != "" {
			owner.UserAgent = closer.UserAgent
		}
	}
	return owner
}

// sessionEndText توضیح فارسی را بدون شناسه نشست می‌سازد.
// kind «پایان نشست» یا «خروج» است. name نام صاحب و reason کد دلیل است.
// خروجی یک جمله است. پیشوند شناسه نشست فقط در Details می‌ماند.
func sessionEndText(kind, name, reason string) string {
	text := kind
	if name != "" {
		text += " کاربر «" + name + "»"
	}
	if label := audit.ReasonLabel(reason); label != "" {
		text += "؛ " + label
	}
	return text
}

// sessionPrefix هشت نویسه اول شناسه نشست را برای جزئیات برمی‌گرداند.
// sess نشست است. خروجی کوتاه است تا UUID کامل در لاگ نماند.
func sessionPrefix(sess *Session) string {
	raw := sess.ID.String()
	if len(raw) <= 8 {
		return raw
	}
	return raw[:8]
}

// closeReason دلیل حذف از صفحه نشست‌ها را از روی فاعل و صاحب انتخاب می‌کند.
// actor درخواست‌کننده است. ownerID صاحب نشست است.
// اگر فاعل همان صاحب نباشد دلیل admin_terminated است، وگرنه خروج کاربر.
func closeReason(actor audit.Actor, ownerID int) string {
	if actor.UserID == nil || *actor.UserID != ownerID {
		return audit.ReasonAdminTerminated
	}
	return audit.ReasonUserLogout
}
