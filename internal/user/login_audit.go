package user

import (
	"context"
	"strconv"
	"time"

	"github.com/tpdenta/afta-reception/internal/platform/apperror"
	"github.com/tpdenta/afta-reception/internal/platform/security/audit"
	"github.com/tpdenta/afta-reception/internal/platform/security/session"
)

// rejectLockedIP اگر IP قفل باشد LOGIN_BLOCKED را حداکثر یک بار در دقیقه ثبت می‌کند.
// ctx زمینه است. actor مبدأ درخواست و username نام تایپ‌شده است.
// خروجی ErrIPLocked یا خطای ممیزی است. nil یعنی IP قفل نیست.
// R1-18: دلیل ip_locked است و شمارنده با میان‌افزار ورود مشترک است.
func (s *Service) rejectLockedIP(ctx context.Context, actor audit.Actor, username string) error {
	err := s.loginGuard.CheckIPLock(actor.IP)
	if err == nil {
		return nil
	}
	if app, ok := err.(*apperror.AppError); !ok || app.Code != apperror.ErrIPLocked.Code {
		return err
	}
	if s.loginGuard.AllowBlockedLoginLog(actor.IP) {
		if recErr := s.record(ctx, loginBlockedEvent(actor, username)); recErr != nil {
			return recErr
		}
	}
	return err
}

// failLogin ورود ناموفق را با دلیل ثبت می‌کند و در صورت نیاز شمارنده IP را بالا می‌برد.
// ctx زمینه است. actor فاعل، username نام هدف و reason کد دلیل است.
// count true یعنی این شکست باید در آستانه قفل IP حساب شود.
// خروجی خطای مناسب کاربر است، یا خطای ممیزی اگر ثبت نشود.
func (s *Service) failLogin(ctx context.Context, actor audit.Actor, username, reason string, count bool) error {
	if err := s.record(ctx, loginFailedEvent(actor, username, reason)); err != nil {
		return err
	}
	if count {
		if err := s.noteIPFailure(ctx, actor); err != nil {
			return err
		}
	}
	return loginFailureError(reason)
}

// noteIPFailure بعد از تلاش ناموفق، اگر IP تازه قفل شده باشد IP_LOCKED می‌نویسد.
// ctx زمینه است. actor مبدأ است. مدت قفل در جزئیات می‌آید.
// خروجی خطای گارد یا ممیزی است. nil یعنی قفلی رخ نداده یا رویداد نوشته شده است.
func (s *Service) noteIPFailure(ctx context.Context, actor audit.Actor) error {
	locked, minutes, err := s.loginGuard.RecordFailedLogin(actor.IP)
	if err != nil {
		return err
	}
	if !locked {
		return nil
	}
	return s.record(ctx, audit.Event{
		Type:        audit.EventIPLocked,
		Result:      audit.ResultSuccess,
		Actor:       actor,
		Target:      audit.Target{Name: actor.IP},
		Details:     map[string]any{"lock_minutes": minutes},
		Description: "قفل نشانی IP",
	})
}

// finishLogin پس از احراز موفق نشست می‌سازد و SESSION_CREATED را با IP و مرورگر ثبت می‌کند.
// ctx زمینه است. actor فاعل است. user کاربر واردشده است.
// خروجی کاربر، نشست و خطا است. شناسه نشست در توضیح رویداد نیست.
func (s *Service) finishLogin(ctx context.Context, actor audit.Actor, user *User) (*User, *session.Session, error) {
	if err := s.loginGuard.ResetLoginAttempts(actor.IP); err != nil {
		return nil, nil, apperror.New("DB_ERROR", "خطا در ورود.", err.Error(), 500)
	}
	sess, err := s.session.CreateSession(ctx, user.ID, actor.IP, actor.UserAgent)
	if err != nil {
		return nil, nil, err
	}
	now := time.Now().UTC()
	user.LastLoginAt = &now
	_ = s.repo.Update(user)
	if err := s.record(ctx, sessionCreatedEvent(actorWithUser(actor, user), user)); err != nil {
		return nil, nil, err
	}
	return user, sess, nil
}

// rejectOutsideWorkHours ورود خارج از ساعت کاری را با دلیل outside_work_hours ثبت می‌کند.
// ctx زمینه است. actor فاعل و user کاربر پیدا‌شده است.
// خروجی nil یعنی ساعت مجاز است. خطای پایگاه بدون رویداد ورود ناموفق برمی‌گردد.
func (s *Service) rejectOutsideWorkHours(ctx context.Context, actor audit.Actor, user *User) error {
	perms, err := s.repo.FindPermissionsByRoleID(user.RoleID)
	if err != nil {
		return apperror.New("DB_ERROR", "خطا در خواندن مجوزها.", err.Error(), 500)
	}
	names := make([]string, 0, len(perms))
	for _, p := range perms {
		names = append(names, p.Name)
	}
	err = s.loginGuard.CheckWorkHours(HasAllPermissions(names))
	if err == nil {
		return nil
	}
	if app, ok := err.(*apperror.AppError); ok && app.Code == apperror.ErrWorkHours.Code {
		return s.failLogin(ctx, actor, user.Username, audit.ReasonOutsideWorkHours, false)
	}
	return err
}

// roleIntegrity یکپارچگی نقش را بدون رویداد بررسی می‌کند.
// role نقش کاربر است. خروجی ErrIntegrity یا خطای خواندن مجوز است.
func (s *Service) roleIntegrity(role *Role) error {
	permIDs, err := s.repo.FindPermissionIDsByRoleID(role.ID)
	if err != nil {
		return apperror.New("DB_ERROR", "خطا در خواندن مجوزهای نقش.", err.Error(), 500)
	}
	if !VerifyRoleIntegrity(s.signer, role, permIDs) {
		return apperror.ErrIntegrity
	}
	return nil
}

// loginFailedEvent رویداد LOGIN_FAILED را با نام کاربری واردشده می‌سازد.
// actor مبدأ است. username متن فرم و reason کد دلیل است. خروجی هنوز ثبت نشده است.
func loginFailedEvent(actor audit.Actor, username, reason string) audit.Event {
	target := audit.Target{Type: audit.TargetUser, Name: username}
	if actor.UserID != nil {
		target.ID = strconv.Itoa(*actor.UserID)
	}
	return audit.Event{
		Type:        audit.EventLoginFailed,
		Result:      audit.ResultFailure,
		Reason:      reason,
		Actor:       actor,
		Target:      target,
		Description: "ورود ناموفق؛ " + audit.ReasonLabel(reason),
	}
}

// loginBlockedEvent تلاش در زمان قفل IP را می‌سازد.
// actor مبدأ و username نام تایپ‌شده است. دلیل همیشه ip_locked است.
func loginBlockedEvent(actor audit.Actor, username string) audit.Event {
	target := audit.Target{}
	if username != "" {
		target = audit.Target{Type: audit.TargetUser, Name: username}
	}
	return audit.Event{
		Type:        audit.EventLoginBlocked,
		Result:      audit.ResultFailure,
		Reason:      audit.ReasonIPLocked,
		Actor:       actor,
		Target:      target,
		Description: "رد ورود؛ نشانی IP قفل است",
	}
}

// sessionCreatedEvent ورود موفق را به‌صورت ایجاد نشست ثبت می‌کند.
// actor کاربر واردشده است و user همان حساب است. IP و مرورگر فقط در جزئیات‌اند.
func sessionCreatedEvent(actor audit.Actor, user *User) audit.Event {
	return audit.Event{
		Type:   audit.EventSessionCreated,
		Result: audit.ResultSuccess,
		Actor:  actor,
		Target: userTarget(user.ID, user.Username),
		Details: map[string]any{
			"ip":      actor.IP,
			"browser": actor.UserAgent,
		},
		Description: "ایجاد نشست؛ ورود موفق «" + user.Username + "»",
	}
}

// actorWithUser شناسه و نام کاربر پیدا‌شده را روی فاعل کپی می‌کند.
// base فاعل درخواست است و user حساب است. شناسه کپی می‌شود تا به فیلد کاربر وصل نماند.
func actorWithUser(base audit.Actor, user *User) audit.Actor {
	if user == nil {
		return base
	}
	id := user.ID
	base.UserID = &id
	base.Username = user.Username
	return base
}

// loginFailureError کد دلیل را به خطای قابل نمایش برای کاربر تبدیل می‌کند.
// reason کد ستون Reason است. پیام کاربر عمومی می‌ماند و دلیل دقیق فقط در ممیزی است.
func loginFailureError(reason string) error {
	switch reason {
	case audit.ReasonAccountLocked, audit.ReasonAccountInactive:
		return apperror.ErrAccountLocked
	case audit.ReasonOutsideWorkHours:
		return apperror.ErrWorkHours
	case audit.ReasonIntegrityViolation:
		return apperror.ErrIntegrity
	default:
		return apperror.ErrInvalidCreds
	}
}
