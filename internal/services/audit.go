package services

import (
	"context"
	"strconv"
	"strings"

	"github.com/tpdenta/afta-reception/internal/platform/apperror"
	"github.com/tpdenta/afta-reception/internal/platform/security/audit"
)

// serviceFields فیلدهای قابل مقایسه به‌روزرسانی خدمت است.
var serviceFields = []string{
	"service_code", "name", "technical_coefficient", "professional_coefficient", "consumption_coefficient",
	"service_rate", "service_tariff", "specialist_rate", "specialist_tariff", "international_code",
	"default_count", "maximum_count", "service_features", "is_active", "has_dental_direction", "has_tooth", "allow_multiple_use",
}

// serviceSnap وضعیت قابل ممیزی خدمت است.
type serviceSnap struct {
	ServiceCode             string  `json:"service_code"`
	Name                    string  `json:"name"`
	TechnicalCoefficient    float64 `json:"technical_coefficient"`
	ProfessionalCoefficient float64 `json:"professional_coefficient"`
	ConsumptionCoefficient  float64 `json:"consumption_coefficient"`
	ServiceRate             int     `json:"service_rate"`
	ServiceTariff           int     `json:"service_tariff"`
	SpecialistRate          int     `json:"specialist_rate"`
	SpecialistTariff        int     `json:"specialist_tariff"`
	InternationalCode       string  `json:"international_code"`
	DefaultCount            int     `json:"default_count"`
	MaximumCount            int     `json:"maximum_count"`
	ServiceFeatures         string  `json:"service_features"`
	IsActive                bool    `json:"is_active"`
	HasDentalDirection      bool    `json:"has_dental_direction"`
	HasTooth                bool    `json:"has_tooth"`
	AllowMultipleUse        bool    `json:"allow_multiple_use"`
}

// serviceTarget هدف رویداد خدمت را می‌سازد. نام هدف کد خدمت است.
// item خدمت است. خروجی Target با نوع service است.
func serviceTarget(item *ServiceItem) audit.Target {
	if item == nil {
		return audit.Target{Type: audit.TargetService}
	}
	return audit.Target{
		Type: audit.TargetService,
		ID:   strconv.FormatUint(uint64(item.ID), 10),
		Name: item.ServiceCode,
	}
}

// excludedTarget هدف خدمت خارج‌شده را می‌سازد.
// organizationID سازمان، serviceID خدمت و code کد خدمت است. نام هدف کد است اگر باشد.
func excludedTarget(organizationID, serviceID uint, code string) audit.Target {
	if code == "" {
		code = strconv.FormatUint(uint64(serviceID), 10)
	}
	return audit.Target{
		Type: audit.TargetExcludedService,
		ID:   strconv.FormatUint(uint64(organizationID), 10) + ":" + strconv.FormatUint(uint64(serviceID), 10),
		Name: code,
	}
}

// snapService فیلدهای قابل مقایسه خدمت را می‌سازد.
// item خدمت است. خروجی ورودی audit.Diff است.
func snapService(item *ServiceItem) serviceSnap {
	if item == nil {
		return serviceSnap{}
	}
	return serviceSnap{
		ServiceCode: item.ServiceCode, Name: item.Name,
		TechnicalCoefficient: item.TechnicalCoefficient, ProfessionalCoefficient: item.ProfessionalCoefficient,
		ConsumptionCoefficient: item.ConsumptionCoefficient, ServiceRate: item.ServiceRate, ServiceTariff: item.ServiceTariff,
		SpecialistRate: item.SpecialistRate, SpecialistTariff: item.SpecialistTariff, InternationalCode: item.InternationalCode,
		DefaultCount: item.DefaultCount, MaximumCount: item.MaximumCount, ServiceFeatures: item.ServiceFeatures,
		IsActive: item.IsActive, HasDentalDirection: item.HasDentalDirection, HasTooth: item.HasTooth, AllowMultipleUse: item.AllowMultipleUse,
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

// recordIntegrity نقض یکپارچگی را با هدف داده‌شده ثبت می‌کند.
// ctx زمینه است. actor فاعل، target موجود و description توضیح فارسی است. خروجی خطای ممیزی است.
func (s *Service) recordIntegrity(ctx context.Context, actor audit.Actor, target audit.Target, description string) error {
	return s.record(ctx, audit.Event{
		Type: audit.EventDataTampering, Result: audit.ResultFailure, Reason: audit.ReasonIntegrityViolation,
		Actor: actor, Target: target, Description: description,
	})
}

// recordCreated ایجاد خدمت را ثبت می‌کند.
// ctx زمینه است. actor سازنده و item خدمت جدید است. خروجی خطای ممیزی است.
func (s *Service) recordCreated(ctx context.Context, actor audit.Actor, item *ServiceItem) error {
	target := serviceTarget(item)
	return s.record(ctx, audit.Event{
		Type: audit.EventEntityCreated, Result: audit.ResultSuccess,
		Actor: actor, Target: target, Description: "ایجاد خدمت «" + target.Name + "»",
	})
}

// recordUpdated اختلاف خدمت را ثبت می‌کند.
// ctx زمینه است. before خدمت قبلی و after خدمت جدید است. بدون تغییر رویدادی نوشته نمی‌شود.
func (s *Service) recordUpdated(ctx context.Context, actor audit.Actor, before, after *ServiceItem) error {
	details := audit.Diff(snapService(before), snapService(after), serviceFields)
	if len(details) == 0 {
		return nil
	}
	target := serviceTarget(after)
	return s.record(ctx, audit.Event{
		Type: audit.EventEntityUpdated, Result: audit.ResultSuccess,
		Actor: actor, Target: target, Details: details, Description: "به‌روزرسانی خدمت «" + target.Name + "»",
	})
}

// recordDeleted حذف خدمت را ثبت می‌کند.
// ctx زمینه است. actor حذف‌کننده و item خدمت حذف‌شده است. خروجی خطای ممیزی است.
func (s *Service) recordDeleted(ctx context.Context, actor audit.Actor, item *ServiceItem) error {
	target := serviceTarget(item)
	return s.record(ctx, audit.Event{
		Type: audit.EventEntityDeleted, Result: audit.ResultSuccess,
		Actor: actor, Target: target, Description: "حذف خدمت «" + target.Name + "»",
	})
}

// recordExcluded افزودن یا حذف خدمات خارج‌شده را با کد خدمت ثبت می‌کند.
// ctx زمینه است. eventType نوع رویداد، organizationID سازمان و serviceIDs شناسه‌ها هستند.
// description شروع توضیح فارسی است. نام هدف کدهای خدمت است.
func (s *Service) recordExcluded(ctx context.Context, actor audit.Actor, eventType string, organizationID uint, serviceIDs []uint, description string) error {
	codes := make([]string, 0, len(serviceIDs))
	for _, id := range serviceIDs {
		code := strconv.FormatUint(uint64(id), 10)
		if item, err := s.repo.FindByID(id); err == nil && item != nil && item.ServiceCode != "" {
			code = item.ServiceCode
		}
		codes = append(codes, code)
	}
	name := strings.Join(codes, "، ")
	return s.record(ctx, audit.Event{
		Type: eventType, Result: audit.ResultSuccess, Actor: actor,
		Target: audit.Target{
			Type: audit.TargetExcludedService,
			ID:   strconv.FormatUint(uint64(organizationID), 10),
			Name: name,
		},
		Description: description + " «" + name + "»",
	})
}
