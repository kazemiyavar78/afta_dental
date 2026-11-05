package organizationpackage

import (
	"context"
	"strconv"

	"github.com/tpdenta/afta-reception/internal/platform/apperror"
	"github.com/tpdenta/afta-reception/internal/platform/security/audit"
)

// packageFields فیلدهای قابل مقایسه به‌روزرسانی بسته تعرفه است.
var packageFields = []string{
	"package_name", "package_description", "technical_coefficient",
	"technical_professional_coefficient", "consumption_coefficient",
	"subsidy_percentage", "supplementary_percentage", "organization_percentage",
}

// packageSnap وضعیت قابل ممیزی بسته تعرفه است.
type packageSnap struct {
	PackageName                      string `json:"package_name"`
	PackageDescription               string `json:"package_description"`
	TechnicalCoefficient             int    `json:"technical_coefficient"`
	TechnicalProfessionalCoefficient int    `json:"technical_professional_coefficient"`
	ConsumptionCoefficient           int    `json:"consumption_coefficient"`
	SubsidyPercentage                int    `json:"subsidy_percentage"`
	SupplementaryPercentage          int    `json:"supplementary_percentage"`
	OrganizationPercentage           int    `json:"organization_percentage"`
}

// packageTarget هدف رویداد بسته تعرفه را می‌سازد.
// o بسته است. نام هدف نام بسته است.
func packageTarget(o *OrganizationPackage) audit.Target {
	if o == nil {
		return audit.Target{Type: audit.TargetOrganizationPackage}
	}
	return audit.Target{
		Type: audit.TargetOrganizationPackage,
		ID:   strconv.FormatUint(uint64(o.ID), 10),
		Name: o.PackageName,
	}
}

// snapPackage فیلدهای قابل مقایسه را می‌سازد.
// o بسته است. خروجی ورودی audit.Diff است.
func snapPackage(o *OrganizationPackage) packageSnap {
	if o == nil {
		return packageSnap{}
	}
	return packageSnap{
		PackageName: o.PackageName, PackageDescription: o.PackageDescription,
		TechnicalCoefficient: o.TechnicalCoefficient, TechnicalProfessionalCoefficient: o.TechnicalProfessionalCoefficient,
		ConsumptionCoefficient: o.ConsumptionCoefficient, SubsidyPercentage: o.SubsidyPercentage,
		SupplementaryPercentage: o.SupplementaryPercentage, OrganizationPercentage: o.OrganizationPercentage,
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

// recordIntegrity نقض یکپارچگی بسته را ثبت می‌کند.
// ctx زمینه است. actor فاعل و o بسته دستکاری‌شده است. خروجی خطای ممیزی است.
func (s *Service) recordIntegrity(ctx context.Context, actor audit.Actor, o *OrganizationPackage) error {
	target := packageTarget(o)
	return s.record(ctx, audit.Event{
		Type: audit.EventDataTampering, Result: audit.ResultFailure, Reason: audit.ReasonIntegrityViolation,
		Actor: actor, Target: target, Description: "نقض یکپارچگی بسته تعرفه «" + target.Name + "»",
	})
}

// recordCreated ایجاد بسته را ثبت می‌کند.
// ctx زمینه است. actor سازنده و o بسته جدید است. خروجی خطای ممیزی است.
func (s *Service) recordCreated(ctx context.Context, actor audit.Actor, o *OrganizationPackage) error {
	target := packageTarget(o)
	return s.record(ctx, audit.Event{
		Type: audit.EventEntityCreated, Result: audit.ResultSuccess,
		Actor: actor, Target: target, Description: "ایجاد بسته تعرفه «" + target.Name + "»",
	})
}

// recordUpdated اختلاف بسته را ثبت می‌کند.
// ctx زمینه است. before بسته قبلی و after بسته جدید است. بدون تغییر رویدادی نوشته نمی‌شود.
func (s *Service) recordUpdated(ctx context.Context, actor audit.Actor, before, after *OrganizationPackage) error {
	details := audit.Diff(snapPackage(before), snapPackage(after), packageFields)
	if len(details) == 0 {
		return nil
	}
	target := packageTarget(after)
	return s.record(ctx, audit.Event{
		Type: audit.EventEntityUpdated, Result: audit.ResultSuccess,
		Actor: actor, Target: target, Details: details, Description: "به‌روزرسانی بسته تعرفه «" + target.Name + "»",
	})
}

// recordDeleted حذف بسته را ثبت می‌کند.
// ctx زمینه است. actor حذف‌کننده و o بسته حذف‌شده است. خروجی خطای ممیزی است.
func (s *Service) recordDeleted(ctx context.Context, actor audit.Actor, o *OrganizationPackage) error {
	target := packageTarget(o)
	return s.record(ctx, audit.Event{
		Type: audit.EventEntityDeleted, Result: audit.ResultSuccess,
		Actor: actor, Target: target, Description: "حذف بسته تعرفه «" + target.Name + "»",
	})
}
