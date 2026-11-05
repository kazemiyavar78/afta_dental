package specialcode

import (
	"context"
	"strconv"

	"github.com/tpdenta/afta-reception/internal/platform/apperror"
	"github.com/tpdenta/afta-reception/internal/platform/security/audit"
)

// specialCodeFields فیلدهای قابل مقایسه به‌روزرسانی کد خاص است.
var specialCodeFields = []string{"code", "name", "description", "percentage", "is_active"}

// specialCodeSnap وضعیت قابل ممیزی کد خاص است.
type specialCodeSnap struct {
	Code        string `json:"code"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Percentage  uint8  `json:"percentage"`
	IsActive    bool   `json:"is_active"`
}

// specialTarget هدف رویداد را از روی کد می‌سازد.
// item کد خاص است. نام هدف همان کد است.
func specialTarget(item *SpecialCode) audit.Target {
	if item == nil {
		return audit.Target{Type: audit.TargetSpecialCode}
	}
	return audit.Target{
		Type: audit.TargetSpecialCode,
		ID:   strconv.FormatUint(uint64(item.ID), 10),
		Name: item.Code,
	}
}

// snapSpecialCode فیلدهای قابل مقایسه را می‌سازد.
// item کد خاص است. خروجی ورودی audit.Diff است.
func snapSpecialCode(item *SpecialCode) specialCodeSnap {
	if item == nil {
		return specialCodeSnap{}
	}
	return specialCodeSnap{
		Code: item.Code, Name: item.Name, Description: item.Description,
		Percentage: item.Percentage, IsActive: item.IsActive,
	}
}

// record رویداد را می‌نویسد و شکست را به AUDIT_UNAVAILABLE تبدیل می‌کند.
// ctx زمینه است. e رویداد است. nil یعنی ثبت شد.
func (s *Service) record(ctx context.Context, e audit.Event) error {
	if s.audit == nil {
		return audit.Unavailable(apperror.ErrInternal)
	}
	if ctx == nil {
		ctx = context.Background()
	}
	return audit.Unavailable(s.audit.Record(ctx, e))
}

// recordIntegrity نقض یکپارچگی کد خاص را ثبت می‌کند.
// ctx زمینه است. actor فاعل و item کد دستکاری‌شده است. خروجی خطای ممیزی است.
func (s *Service) recordIntegrity(ctx context.Context, actor audit.Actor, item *SpecialCode) error {
	target := specialTarget(item)
	return s.record(ctx, audit.Event{
		Type:        audit.EventDataTampering,
		Result:      audit.ResultFailure,
		Reason:      audit.ReasonIntegrityViolation,
		Actor:       actor,
		Target:      target,
		Description: "نقض یکپارچگی کد خاص «" + target.Name + "»",
	})
}

// recordCreated ایجاد کد خاص را ثبت می‌کند.
// ctx زمینه است. actor سازنده و item کد جدید است. خروجی خطای ممیزی است.
func (s *Service) recordCreated(ctx context.Context, actor audit.Actor, item *SpecialCode) error {
	target := specialTarget(item)
	return s.record(ctx, audit.Event{
		Type:        audit.EventEntityCreated,
		Result:      audit.ResultSuccess,
		Actor:       actor,
		Target:      target,
		Description: "ایجاد کد خاص «" + target.Name + "»",
	})
}

// recordUpdated اختلاف کد خاص را ثبت می‌کند.
// ctx زمینه است. before کد قبلی و after کد جدید است. بدون تغییر، رویدادی نوشته نمی‌شود.
func (s *Service) recordUpdated(ctx context.Context, actor audit.Actor, before, after *SpecialCode) error {
	details := audit.Diff(snapSpecialCode(before), snapSpecialCode(after), specialCodeFields)
	if len(details) == 0 {
		return nil
	}
	target := specialTarget(after)
	return s.record(ctx, audit.Event{
		Type:        audit.EventEntityUpdated,
		Result:      audit.ResultSuccess,
		Actor:       actor,
		Target:      target,
		Details:     details,
		Description: "به‌روزرسانی کد خاص «" + target.Name + "»",
	})
}

// recordDeleted حذف کد خاص را ثبت می‌کند.
// ctx زمینه است. actor حذف‌کننده و item کد حذف‌شده است. خروجی خطای ممیزی است.
func (s *Service) recordDeleted(ctx context.Context, actor audit.Actor, item *SpecialCode) error {
	target := specialTarget(item)
	return s.record(ctx, audit.Event{
		Type:        audit.EventEntityDeleted,
		Result:      audit.ResultSuccess,
		Actor:       actor,
		Target:      target,
		Description: "حذف کد خاص «" + target.Name + "»",
	})
}
