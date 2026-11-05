package reception

import (
	"context"
	"strconv"

	"github.com/tpdenta/afta-reception/internal/patient"
	"github.com/tpdenta/afta-reception/internal/platform/apperror"
	"github.com/tpdenta/afta-reception/internal/platform/security/audit"
	"gorm.io/gorm"
)

// receptionFields فیلدهای قابل مقایسه پذیرش است. خدمات در رویداد جداگانه صندوق ثبت می‌شوند.
var receptionFields = []string{
	"status", "description", "discount", "insurance_id", "additional_insurance_id",
	"special_code_id", "doctor_id", "assistant_id", "referral_code",
	"additional_insurance_coverage", "additional_insurance_percentage", "reception_ended",
}

// receptionSnap فیلدهای قابل ممیزی پذیرش است.
type receptionSnap struct {
	Status                        string `json:"status"`
	Description                   string `json:"description"`
	Discount                      int64  `json:"discount"`
	InsuranceID                   *uint  `json:"insurance_id"`
	AdditionalInsuranceID         *uint  `json:"additional_insurance_id"`
	SpecialCodeID                 *uint  `json:"special_code_id"`
	DoctorID                      *uint  `json:"doctor_id"`
	AssistantID                   *uint  `json:"assistant_id"`
	ReferralCode                  *int64 `json:"referral_code"`
	AdditionalInsuranceCoverage   *int64 `json:"additional_insurance_coverage"`
	AdditionalInsurancePercentage *uint8 `json:"additional_insurance_percentage"`
	ReceptionEnded                bool   `json:"reception_ended"`
}

// record رویداد پذیرش را بیرون از تراکنش می‌نویسد.
// ctx زمینه است. e رویداد است. خطای ثبت‌کننده به AUDIT_UNAVAILABLE تبدیل می‌شود.
func (s *Service) record(ctx context.Context, e audit.Event) error {
	if s.audit == nil {
		return audit.Unavailable(apperror.ErrInternal)
	}
	if ctx == nil {
		ctx = context.Background()
	}
	return audit.Unavailable(s.audit.Record(ctx, e))
}

// recordTx رویداد پذیرش را داخل تراکنش ذخیره می‌نویسد.
// tx تراکنش باز است. e رویداد است. خطا باید همان تراکنش را برگرداند.
func (s *Service) recordTx(tx *gorm.DB, e audit.Event) error {
	if s.audit == nil {
		return audit.Unavailable(apperror.ErrInternal)
	}
	return audit.Unavailable(s.audit.RecordTx(tx, e))
}

// receptionCreatedEvent رویداد ایجاد پذیرش را می‌سازد.
// actor ثبت‌کننده و rec پذیرش ذخیره‌شده است. نام هدف نام بیمار است.
func (s *Service) receptionCreatedEvent(actor audit.Actor, rec *Reception) audit.Event {
	return audit.Event{
		Type: audit.EventEntityCreated, Result: audit.ResultSuccess, Actor: actor,
		Target: s.receptionTarget(rec), Description: "ایجاد پذیرش",
	}
}

// receptionUpdatedEvent اختلاف پذیرش را می‌سازد.
// before نسخه قبل از تغییر و after نسخه جدید است. خروجی دوم false یعنی تغییری نیست.
func (s *Service) receptionUpdatedEvent(actor audit.Actor, before, after *Reception) (audit.Event, bool) {
	details := audit.Diff(snapReception(before), snapReception(after), receptionFields)
	if len(details) == 0 {
		return audit.Event{}, false
	}
	return audit.Event{
		Type: audit.EventEntityUpdated, Result: audit.ResultSuccess, Actor: actor,
		Target: s.receptionTarget(after), Details: details, Description: "به‌روزرسانی پذیرش",
	}, true
}

// receptionTarget هدف پذیرش را با نام نمایشی بیمار می‌سازد.
// rec پذیرش است. نام از پرونده بیمار خوانده می‌شود و کد ملی در آن نیست.
func (s *Service) receptionTarget(rec *Reception) audit.Target {
	if rec == nil {
		return audit.Target{Type: audit.TargetReception}
	}
	return audit.Target{
		Type: audit.TargetReception,
		ID:   strconv.FormatUint(uint64(rec.ID), 10),
		Name: s.receptionName(rec),
	}
}

// receptionName نام خوانای بیمار پذیرش را برمی‌گرداند.
// rec پذیرش است. خروجی «نام + حرف اول نام خانوادگی» است.
func (s *Service) receptionName(rec *Reception) string {
	if rec == nil || s.patientSvc == nil {
		return ""
	}
	p, err := s.patientSvc.Get(rec.PatientID)
	if err != nil || p == nil {
		return ""
	}
	return patient.DisplayName(p.FirstName, p.LastName)
}

// snapReception فیلدهای قابل مقایسه پذیرش را می‌سازد.
// rec پذیرش است. خروجی ورودی audit.Diff است.
func snapReception(rec *Reception) receptionSnap {
	if rec == nil {
		return receptionSnap{}
	}
	return receptionSnap{
		Status: rec.Status, Description: rec.Description, Discount: rec.Discount,
		InsuranceID: rec.InsuranceID, AdditionalInsuranceID: rec.AdditionalInsuranceID,
		SpecialCodeID: rec.SpecialCodeID, DoctorID: rec.DoctorID, AssistantID: rec.AssistantID,
		ReferralCode: rec.ReferralCode, AdditionalInsuranceCoverage: rec.AdditionalInsuranceCoverage,
		AdditionalInsurancePercentage: rec.AdditionalInsurancePercentage, ReceptionEnded: rec.ReceptionEnded,
	}
}
