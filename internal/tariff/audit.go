package tariff

import (
	"context"
	"strconv"
	"strings"

	"github.com/tpdenta/afta-reception/internal/platform/apperror"
	"github.com/tpdenta/afta-reception/internal/platform/security/audit"
	"gorm.io/gorm"
)

// tariffFields فیلدهای مبلغی قابل مقایسه بازمحاسبه تعرفه است.
var tariffFields = []string{"amount", "tariff_amount", "organization_share", "supplementary_share", "subsidy_share"}

// tariffSnap مبالغ قابل ممیزی یک ردیف تعرفه است.
type tariffSnap struct {
	Amount             int64 `json:"amount"`
	TariffAmount       int64 `json:"tariff_amount"`
	OrganizationShare  int64 `json:"organization_share"`
	SupplementaryShare int64 `json:"supplementary_share"`
	SubsidyShare       int64 `json:"subsidy_share"`
}

// snapTariff مبالغ تعرفه را برای Diff می‌سازد.
// row ردیف تعرفه است. خروجی ورودی audit.Diff است.
func snapTariff(row *Tariff) tariffSnap {
	if row == nil {
		return tariffSnap{}
	}
	return tariffSnap{
		Amount: row.Amount, TariffAmount: row.TariffAmount, OrganizationShare: row.OrganizationShare,
		SupplementaryShare: row.SupplementaryShare, SubsidyShare: row.SubsidyShare,
	}
}

// record رویداد را بیرون از تراکنش می‌نویسد.
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

// recordTx رویداد را داخل تراکنش ذخیره تعرفه می‌نویسد.
// tx تراکنش باز است. e رویداد است. خطا باید تراکنش را برگرداند.
func (s *Service) recordTx(tx *gorm.DB, e audit.Event) error {
	if s.audit == nil {
		return audit.Unavailable(apperror.ErrInternal)
	}
	return audit.Unavailable(s.audit.RecordTx(tx, e))
}

// recordTariffSave ذخیره دسته‌ای تعرفه را داخل همان تراکنش ثبت می‌کند.
// tx تراکنش ذخیره است. actor فاعل، orgName نام سازمان و organizationID شناسه آن است.
// createdCount و updatedCount تعداد ردیف‌ها و codes کد خدمت‌ها هستند.
func (s *Service) recordTariffSave(tx *gorm.DB, actor audit.Actor, orgName string, organizationID uint, createdCount, updatedCount int, codes []string) error {
	eventType := audit.EventEntityUpdated
	if createdCount > 0 && updatedCount == 0 {
		eventType = audit.EventEntityCreated
	}
	name := strings.TrimSpace(orgName)
	if len(codes) == 1 && codes[0] != "" {
		name = codes[0]
	}
	return s.recordTx(tx, audit.Event{
		Type:   eventType,
		Result: audit.ResultSuccess,
		Actor:  actor,
		Target: audit.Target{
			Type: audit.TargetTariff,
			ID:   strconv.FormatUint(uint64(organizationID), 10),
			Name: name,
		},
		Details: map[string]any{
			"created":       createdCount,
			"updated":       updatedCount,
			"service_codes": codes,
		},
		Description: "ذخیره تعرفه «" + name + "»",
	})
}

// recordTariffUpdated اختلاف مبالغ یک تعرفه را ثبت می‌کند.
// ctx زمینه است. actor ویرایشگر، before ردیف قبلی، after ردیف جدید و serviceCode کد خدمت است.
// اگر مبلغی عوض نشده باشد رویدادی نوشته نمی‌شود.
func (s *Service) recordTariffUpdated(ctx context.Context, actor audit.Actor, before, after *Tariff, serviceCode string) error {
	details := audit.Diff(snapTariff(before), snapTariff(after), tariffFields)
	if len(details) == 0 {
		return nil
	}
	return s.record(ctx, audit.Event{
		Type:   audit.EventEntityUpdated,
		Result: audit.ResultSuccess,
		Actor:  actor,
		Target: audit.Target{
			Type: audit.TargetTariff,
			ID:   strconv.FormatUint(uint64(after.ID), 10),
			Name: serviceCode,
		},
		Details:     details,
		Description: "بازمحاسبه تعرفه «" + serviceCode + "»",
	})
}
