package user

import (
	"context"
	"sort"
	"strconv"
	"strings"

	"github.com/tpdenta/afta-reception/internal/platform/apperror"
	"github.com/tpdenta/afta-reception/internal/platform/security/audit"
)

// fieldChange مقدار قبل و بعد یک فیلد تغییریافته است.
type fieldChange struct {
	Before any `json:"before"`
	After  any `json:"after"`
}

// record رویداد را می‌نویسد و شکست ثبت را به خطای عمومی عملیات تبدیل می‌کند.
// ctx لغو درخواست است. e رویداد آماده است. خروجی nil یعنی ردیف نوشته شد.
// FAU_STG.4: صدازننده نباید این خطا را نادیده بگیرد.
func (s *Service) record(ctx context.Context, e audit.Event) error {
	if s.audit == nil {
		return audit.Unavailable(apperror.ErrInternal)
	}
	return audit.Unavailable(s.audit.Record(ctx, e))
}

// recordIntegrity نقض یکپارچگی را با فاعل و هدف ثبت می‌کند.
// ctx زمینه است. actor فاعل، target موجود دستکاری‌شده و description توضیح فارسی است.
// خروجی خطای ممیزی است. اگر nil باشد صدازننده ErrIntegrity را برمی‌گرداند.
func (s *Service) recordIntegrity(ctx context.Context, actor audit.Actor, target audit.Target, description string) error {
	return s.record(ctx, audit.Event{
		Type:        audit.EventDataTampering,
		Result:      audit.ResultFailure,
		Reason:      audit.ReasonIntegrityViolation,
		Actor:       actor,
		Target:      target,
		Description: description,
	})
}

// userTarget هدف رویداد کاربر را از روی شناسه و نام کاربری می‌سازد.
// id شناسه کاربر و username نام اوست. خروجی Target با نوع user است.
func userTarget(id int, username string) audit.Target {
	return audit.Target{Type: audit.TargetUser, ID: strconv.Itoa(id), Name: username}
}

// roleTarget هدف رویداد نقش را می‌سازد.
// role نقش ذخیره‌شده است. نام هدف همان نام نقش است.
func roleTarget(role *Role) audit.Target {
	if role == nil {
		return audit.Target{Type: audit.TargetRole}
	}
	return audit.Target{Type: audit.TargetRole, ID: strconv.Itoa(role.ID), Name: role.Name}
}

// recordRoleCreated ایجاد نقش را با نام نقش به‌عنوان هدف ثبت می‌کند.
// ctx زمینه است. actor سازنده و role نقش جدید است. permNames نام مجوزهای اولیه است.
// خروجی خطای ممیزی است. FAU_GEN.2 / R1-02: نام کاربری فاعل در خود رویداد می‌ماند.
func (s *Service) recordRoleCreated(ctx context.Context, actor audit.Actor, role *Role, permNames []string) error {
	details := map[string]any{}
	if len(permNames) > 0 {
		details["permissions"] = append([]string(nil), permNames...)
	}
	return s.record(ctx, audit.Event{
		Type:        audit.EventRoleCreated,
		Result:      audit.ResultSuccess,
		Actor:       actor,
		Target:      roleTarget(role),
		Details:     details,
		Description: "ایجاد نقش «" + role.Name + "»",
	})
}

// recordRoleDeleted حذف نقش را ثبت می‌کند.
// ctx زمینه است. actor حذف‌کننده و role نقش حذف‌شده است. خروجی خطای ممیزی است.
func (s *Service) recordRoleDeleted(ctx context.Context, actor audit.Actor, role *Role) error {
	return s.record(ctx, audit.Event{
		Type:        audit.EventRoleDeleted,
		Result:      audit.ResultSuccess,
		Actor:       actor,
		Target:      roleTarget(role),
		Description: "حذف نقش «" + role.Name + "»",
	})
}

// recordRoleUpdated فقط فیلدهای عوض‌شده نقش را ثبت می‌کند.
// ctx زمینه است. actor ویرایشگر است. before نقش قبلی و after نقش جدید است.
// beforePerms و afterPerms مجوزهای قبل و بعد هستند. اگر هیچ فرقی نباشد رویدادی نوشته نمی‌شود.
// R1-04: توضیح، مجوزهای اضافه‌شده و حذف‌شده را به فارسی می‌گوید.
func (s *Service) recordRoleUpdated(ctx context.Context, actor audit.Actor, before, after Role, beforePerms, afterPerms []Permission) error {
	details, description, changed := roleUpdateDetails(before, after, beforePerms, afterPerms)
	if !changed {
		return nil
	}
	return s.record(ctx, audit.Event{
		Type:        audit.EventRoleUpdated,
		Result:      audit.ResultSuccess,
		Actor:       actor,
		Target:      roleTarget(&after),
		Details:     details,
		Description: description,
	})
}

// roleUpdateDetails جزئیات و توضیح فارسی به‌روزرسانی نقش را می‌سازد.
// before و after نقش، و فهرست مجوزها ورودی‌اند.
// خروجی نقشه فقط فیلدهای عوض‌شده، جمله فارسی، و اینکه تغییری بوده یا نه است.
func roleUpdateDetails(before, after Role, beforePerms, afterPerms []Permission) (map[string]any, string, bool) {
	details := map[string]any{}
	var parts []string
	if before.Name != after.Name {
		details["name"] = fieldChange{Before: before.Name, After: after.Name}
		parts = append(parts, "نام: «"+before.Name+"» به «"+after.Name+"»")
	}
	if before.Description != after.Description {
		details["description"] = fieldChange{Before: before.Description, After: after.Description}
		parts = append(parts, "توضیحات: «"+before.Description+"» به «"+after.Description+"»")
	}
	added, removed := diffNames(permissionNames(beforePerms), permissionNames(afterPerms))
	if len(added) > 0 || len(removed) > 0 {
		details["added_permissions"] = added
		details["removed_permissions"] = removed
		parts = append(parts, permissionChangeText(added, removed))
	}
	if len(details) == 0 {
		return nil, "", false
	}
	return details, strings.Join(parts, "؛ "), true
}

// permissionChangeText جمله مجوزهای اضافه‌شده و حذف‌شده را فقط برای طرفِ غیرخالی می‌سازد.
// added و removed نام مجوزها هستند. خروجی با «؛ » به بقیه توضیح وصل می‌شود.
func permissionChangeText(added, removed []string) string {
	var parts []string
	if len(added) > 0 {
		parts = append(parts, "مجوزهای اضافه‌شده: "+strings.Join(added, "، "))
	}
	if len(removed) > 0 {
		parts = append(parts, "مجوزهای حذف‌شده: "+strings.Join(removed, "، "))
	}
	return strings.Join(parts, "؛ ")
}

// permissionNames نام مجوزها را به ترتیب ثابت برمی‌گرداند.
// perms فهرست مجوز است. خروجی برای diff و جزئیات است.
func permissionNames(perms []Permission) []string {
	names := make([]string, 0, len(perms))
	for _, p := range perms {
		names = append(names, p.Name)
	}
	sort.Strings(names)
	return names
}

// diffNames نام‌های اضافه‌شده و حذف‌شده را برمی‌گرداند.
// before و after مجموعه‌های نام هستند. هر دو خروجی مرتب‌اند و اگر طرفی خالی باشد اسلایس خالی است نه nil.
func diffNames(before, after []string) (added, removed []string) {
	oldSet := map[string]struct{}{}
	for _, name := range before {
		oldSet[name] = struct{}{}
	}
	newSet := map[string]struct{}{}
	for _, name := range after {
		newSet[name] = struct{}{}
	}
	added = []string{}
	removed = []string{}
	for _, name := range after {
		if _, ok := oldSet[name]; !ok {
			added = append(added, name)
		}
	}
	for _, name := range before {
		if _, ok := newSet[name]; !ok {
			removed = append(removed, name)
		}
	}
	sort.Strings(added)
	sort.Strings(removed)
	return added, removed
}

// recordUserCreated ایجاد کاربر را ثبت می‌کند.
// ctx زمینه است. actor سازنده و user کاربر جدید است. رمز و کد امنیتی در رویداد نیستند.
// خروجی خطای ممیزی است. R1-21.
func (s *Service) recordUserCreated(ctx context.Context, actor audit.Actor, user *User) error {
	return s.record(ctx, audit.Event{
		Type:        audit.EventUserCreated,
		Result:      audit.ResultSuccess,
		Actor:       actor,
		Target:      userTarget(user.ID, user.Username),
		Description: "ایجاد کاربر «" + user.Username + "»",
	})
}

// recordUserUpdate اختلاف فیلدهای کاربر را ثبت می‌کند.
// ctx زمینه است. actor ویرایشگر است. before کپی قبل از تغییر و after کاربر بعد از تغییر است.
// oldRole و newRole نام نقش‌اند، نه شناسه. اگر فیلد کسب‌وکار عوض نشده باشد خروجی nil و بدون رویداد است.
// PasswordHash و SecurityCode هرگز در جزئیات نمی‌آیند. R1-21.
func (s *Service) recordUserUpdate(ctx context.Context, actor audit.Actor, before, after User, oldRole, newRole string) error {
	details := userFieldChanges(before, after, oldRole, newRole)
	if len(details) == 0 {
		return nil
	}
	return s.record(ctx, audit.Event{
		Type:        audit.EventUserUpdated,
		Result:      audit.ResultSuccess,
		Actor:       actor,
		Target:      userTarget(after.ID, after.Username),
		Details:     details,
		Description: userUpdateDescription(details),
	})
}

// recordPasswordChanged فقط رویداد تغییر رمز را می‌نویسد و خود رمز را ذخیره نمی‌کند.
// ctx زمینه است. actor کاربر و user همان حساب است. خروجی خطای ممیزی است.
func (s *Service) recordPasswordChanged(ctx context.Context, actor audit.Actor, user *User) error {
	return s.record(ctx, audit.Event{
		Type:        audit.EventPasswordChanged,
		Result:      audit.ResultSuccess,
		Actor:       actor,
		Target:      userTarget(user.ID, user.Username),
		Description: "تغییر رمز عبور",
	})
}

// userFieldChanges فیلدهای قابل نمایش را که عوض شده‌اند برمی‌گرداند.
// before و after دو وضعیت کاربرند. oldRole و newRole نام نقش‌اند.
// خروجی nil است اگر چیزی عوض نشده باشد. رمز و کد امنیتی مقایسه نمی‌شوند.
func userFieldChanges(before, after User, oldRole, newRole string) map[string]any {
	details := map[string]any{}
	putChange(details, "name", before.Name, after.Name)
	putChange(details, "family", before.Family, after.Family)
	if before.PhoneNumber != after.PhoneNumber {
		details["phone"] = fieldChange{Before: maskPhone(before.PhoneNumber), After: maskPhone(after.PhoneNumber)}
	}
	putChange(details, "address", before.Address, after.Address)
	putChange(details, "medical_code", medicalText(before.MedicalCode), medicalText(after.MedicalCode))
	if before.RoleID != after.RoleID || oldRole != newRole {
		details["role"] = fieldChange{Before: oldRole, After: newRole}
	}
	putChange(details, "user_type", string(before.UserType), string(after.UserType))
	if before.IsActive != after.IsActive {
		details["is_active"] = fieldChange{Before: before.IsActive, After: after.IsActive}
	}
	if before.IsLocked != after.IsLocked {
		details["is_locked"] = fieldChange{Before: before.IsLocked, After: after.IsLocked}
	}
	if len(details) == 0 {
		return nil
	}
	return details
}

// putChange اگر دو رشته فرق داشته باشند یک before/after در جزئیات می‌گذارد.
// details نقشه خروجی است. key نام فیلد و before/after مقدارها هستند.
func putChange(details map[string]any, key, before, after string) {
	if before == after {
		return
	}
	details[key] = fieldChange{Before: before, After: after}
}

// userUpdateDescription برچسب فارسی فیلدهای داخل جزئیات را به یک جمله تبدیل می‌کند.
// details خروجی userFieldChanges است. ترتیب برچسب‌ها ثابت است.
func userUpdateDescription(details map[string]any) string {
	labels := []struct{ key, label string }{
		{"name", "نام"},
		{"family", "نام خانوادگی"},
		{"phone", "تلفن"},
		{"address", "آدرس"},
		{"medical_code", "کد نظام"},
		{"role", "نقش"},
		{"user_type", "نوع کاربر"},
		{"is_active", "فعال"},
		{"is_locked", "قفل"},
	}
	var names []string
	for _, item := range labels {
		if _, ok := details[item.key]; ok {
			names = append(names, item.label)
		}
	}
	return "فیلدهای تغییریافته: " + strings.Join(names, "، ")
}

// maskPhone همه نویسه‌ها جز چهار رقم آخر را ستاره می‌کند.
// phone شماره خام است. اگر کوتاه‌تر از چهار نویسه باشد همان متن برمی‌گردد.
func maskPhone(phone string) string {
	runes := []rune(phone)
	if len(runes) <= 4 {
		return phone
	}
	return strings.Repeat("*", len(runes)-4) + string(runes[len(runes)-4:])
}

// medicalText مقدار کد نظام را از اشاره‌گر می‌خواند.
// value می‌تواند nil باشد. خروجی رشته خالی به‌جای nil است تا diff ساده بماند.
func medicalText(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

// actorID شناسه عددی فاعل را برمی‌گرداند.
// actor فاعل درخواست است. اگر شناسه نباشد صفر برمی‌گردد.
func actorID(actor audit.Actor) int {
	if actor.UserID == nil {
		return 0
	}
	return *actor.UserID
}
