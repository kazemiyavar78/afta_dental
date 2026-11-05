package organization

import (
	"context"
	"strconv"

	"github.com/tpdenta/afta-reception/internal/platform/apperror"
	"github.com/tpdenta/afta-reception/internal/platform/security/audit"
)

// organizationFields فیلدهای قابل مقایسه به‌روزرسانی سازمان است.
var organizationFields = []string{"name", "is_takmili", "is_active", "is_free", "package_id", "center_package_id"}

// organizationSnap وضعیت قابل ممیزی سازمان است.
type organizationSnap struct {
	Name            string `json:"name"`
	IsTakmili       bool   `json:"is_takmili"`
	IsActive        bool   `json:"is_active"`
	IsFree          bool   `json:"is_free"`
	PackageID       uint   `json:"package_id"`
	CenterPackageID uint   `json:"center_package_id"`
}

// organizationTarget هدف رویداد سازمان را می‌سازد.
// o سازمان است. نام هدف نام سازمان است.
func organizationTarget(o *Organization) audit.Target {
	if o == nil {
		return audit.Target{Type: audit.TargetOrganization}
	}
	return audit.Target{
		Type: audit.TargetOrganization,
		ID:   strconv.FormatUint(uint64(o.ID), 10),
		Name: o.Name,
	}
}

// snapOrganization فیلدهای قابل مقایسه را می‌سازد.
// o سازمان است. خروجی ورودی audit.Diff است.
func snapOrganization(o *Organization) organizationSnap {
	if o == nil {
		return organizationSnap{}
	}
	return organizationSnap{
		Name: o.Name, IsTakmili: o.IsTakmili, IsActive: o.IsActive, IsFree: o.IsFree,
		PackageID: o.PackageID, CenterPackageID: o.CenterPackageID,
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

// recordIntegrity نقض یکپارچگی سازمان را ثبت می‌کند.
// ctx زمینه است. actor فاعل و o سازمان دستکاری‌شده است. خروجی خطای ممیزی است.
func (s *Service) recordIntegrity(ctx context.Context, actor audit.Actor, o *Organization) error {
	target := organizationTarget(o)
	return s.record(ctx, audit.Event{
		Type: audit.EventDataTampering, Result: audit.ResultFailure, Reason: audit.ReasonIntegrityViolation,
		Actor: actor, Target: target, Description: "نقض یکپارچگی سازمان «" + target.Name + "»",
	})
}

// recordCreated ایجاد سازمان را ثبت می‌کند.
// ctx زمینه است. actor سازنده و o سازمان جدید است. خروجی خطای ممیزی است.
func (s *Service) recordCreated(ctx context.Context, actor audit.Actor, o *Organization) error {
	target := organizationTarget(o)
	return s.record(ctx, audit.Event{
		Type: audit.EventEntityCreated, Result: audit.ResultSuccess,
		Actor: actor, Target: target, Description: "ایجاد سازمان «" + target.Name + "»",
	})
}

// recordUpdated اختلاف سازمان را ثبت می‌کند.
// ctx زمینه است. before سازمان قبلی و after سازمان جدید است. بدون تغییر رویدادی نوشته نمی‌شود.
func (s *Service) recordUpdated(ctx context.Context, actor audit.Actor, before, after *Organization) error {
	details := audit.Diff(snapOrganization(before), snapOrganization(after), organizationFields)
	if len(details) == 0 {
		return nil
	}
	target := organizationTarget(after)
	return s.record(ctx, audit.Event{
		Type: audit.EventEntityUpdated, Result: audit.ResultSuccess,
		Actor: actor, Target: target, Details: details, Description: "به‌روزرسانی سازمان «" + target.Name + "»",
	})
}

// recordDeleted حذف سازمان را ثبت می‌کند.
// ctx زمینه است. actor حذف‌کننده و o سازمان حذف‌شده است. خروجی خطای ممیزی است.
func (s *Service) recordDeleted(ctx context.Context, actor audit.Actor, o *Organization) error {
	target := organizationTarget(o)
	return s.record(ctx, audit.Event{
		Type: audit.EventEntityDeleted, Result: audit.ResultSuccess,
		Actor: actor, Target: target, Description: "حذف سازمان «" + target.Name + "»",
	})
}
