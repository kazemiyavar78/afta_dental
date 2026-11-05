package regulation

import (
	"context"
	"sort"
	"strconv"
	"strings"

	"github.com/tpdenta/afta-reception/internal/platform/apperror"
	"github.com/tpdenta/afta-reception/internal/platform/security/audit"
)

// regulationFields فیلدهای قابل مقایسه به‌روزرسانی ضابطه است.
var regulationFields = []string{"duration_days", "is_active", "photo_count", "description", "service_ids"}

// regulationSnap وضعیت قابل ممیزی ضابطه است.
type regulationSnap struct {
	DurationDays int    `json:"duration_days"`
	IsActive     bool   `json:"is_active"`
	PhotoCount   int    `json:"photo_count"`
	Description  string `json:"description"`
	ServiceIDs   string `json:"service_ids"`
}

// regulationTarget هدف رویداد ضابطه را می‌سازد.
// item ضابطه است. نام هدف توضیح است و اگر خالی باشد «ضابطه» می‌ماند.
func regulationTarget(item *Regulation) audit.Target {
	if item == nil {
		return audit.Target{Type: audit.TargetRegulation}
	}
	name := strings.TrimSpace(item.Description)
	if name == "" {
		name = "ضابطه"
	}
	return audit.Target{Type: audit.TargetRegulation, ID: strconv.FormatUint(uint64(item.ID), 10), Name: name}
}

// snapRegulation فیلدها و فهرست مرتب خدمات را برای Diff می‌سازد.
// item ضابطه و services پیوند خدمات است. خروجی ورودی audit.Diff است.
func snapRegulation(item *Regulation, services []RegulationService) regulationSnap {
	if item == nil {
		return regulationSnap{}
	}
	return regulationSnap{
		DurationDays: item.DurationDays,
		IsActive:     item.IsActive,
		PhotoCount:   item.PhotoCount,
		Description:  item.Description,
		ServiceIDs:   joinServiceIDs(services),
	}
}

// joinServiceIDs شناسه خدمات را به ترتیب ثابت به یک رشته تبدیل می‌کند.
// services پیوندها هستند. خروجی برای مقایسه فهرست است.
func joinServiceIDs(services []RegulationService) string {
	ids := make([]int, 0, len(services))
	for _, row := range services {
		ids = append(ids, int(row.ServiceID))
	}
	sort.Ints(ids)
	parts := make([]string, len(ids))
	for i, id := range ids {
		parts[i] = strconv.Itoa(id)
	}
	return strings.Join(parts, ",")
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

// recordCreated ایجاد ضابطه را ثبت می‌کند.
// ctx زمینه است. actor سازنده و item ضابطه جدید است. خروجی خطای ممیزی است.
func (s *Service) recordCreated(ctx context.Context, actor audit.Actor, item *Regulation) error {
	target := regulationTarget(item)
	return s.record(ctx, audit.Event{
		Type:        audit.EventEntityCreated,
		Result:      audit.ResultSuccess,
		Actor:       actor,
		Target:      target,
		Description: "ایجاد ضابطه «" + target.Name + "»",
	})
}

// recordUpdated اختلاف ضابطه را ثبت می‌کند.
// ctx زمینه است. before ضابطه قبلی، beforeServices خدمات قبلی و after ضابطه جدید است.
// اگر چیزی عوض نشده باشد رویدادی نوشته نمی‌شود.
func (s *Service) recordUpdated(ctx context.Context, actor audit.Actor, before *Regulation, beforeServices []RegulationService, after *Regulation) error {
	details := audit.Diff(snapRegulation(before, beforeServices), snapRegulation(after, after.Services), regulationFields)
	if len(details) == 0 {
		return nil
	}
	target := regulationTarget(after)
	return s.record(ctx, audit.Event{
		Type:        audit.EventEntityUpdated,
		Result:      audit.ResultSuccess,
		Actor:       actor,
		Target:      target,
		Details:     details,
		Description: "به‌روزرسانی ضابطه «" + target.Name + "»",
	})
}

// recordDeleted حذف ضابطه را ثبت می‌کند.
// ctx زمینه است. actor حذف‌کننده و item ضابطه حذف‌شده است. خروجی خطای ممیزی است.
func (s *Service) recordDeleted(ctx context.Context, actor audit.Actor, item *Regulation) error {
	target := regulationTarget(item)
	return s.record(ctx, audit.Event{
		Type:        audit.EventEntityDeleted,
		Result:      audit.ResultSuccess,
		Actor:       actor,
		Target:      target,
		Description: "حذف ضابطه «" + target.Name + "»",
	})
}
