package patient

import (
	"context"
	"strconv"
	"strings"

	"github.com/tpdenta/afta-reception/internal/platform/apperror"
	"github.com/tpdenta/afta-reception/internal/platform/security/audit"
)

// patientFields فیلدهایی هستند که به‌روزرسانی بیمار فقط اختلاف همین‌ها را ثبت می‌کند.
var patientFields = []string{
	"first_name", "last_name", "national_code", "birth_date", "address",
	"home_phone", "mobile_phone", "file_number", "sex", "is_foreign_national",
}

// patientSnap وضعیت قابل ممیزی بیمار است. کد ملی و تلفن ماسک شده‌اند.
type patientSnap struct {
	FirstName         string `json:"first_name"`
	LastName          string `json:"last_name"`
	NationalCode      string `json:"national_code"`
	BirthDate         string `json:"birth_date"`
	Address           string `json:"address"`
	HomePhone         string `json:"home_phone"`
	MobilePhone       string `json:"mobile_phone"`
	FileNumber        string `json:"file_number"`
	Sex               bool   `json:"sex"`
	IsForeignNational bool   `json:"is_foreign_national"`
}

// DisplayName نام خوانای بیمار برای هدف ممیزی است.
// first نام و last نام خانوادگی است. خروجی «نام + حرف اول نام خانوادگی» است و کد ملی در آن نیست.
func DisplayName(first, last string) string {
	first = strings.TrimSpace(first)
	last = strings.TrimSpace(last)
	if last == "" {
		return first
	}
	letter := string([]rune(last)[0])
	if first == "" {
		return letter
	}
	return first + " " + letter
}

// patientTarget هدف رویداد بیمار را می‌سازد.
// p بیمار است. نام هدف DisplayName است، نه کد ملی.
func patientTarget(p *Patient) audit.Target {
	if p == nil {
		return audit.Target{Type: audit.TargetPatient}
	}
	return audit.Target{
		Type: audit.TargetPatient,
		ID:   strconv.FormatUint(uint64(p.ID), 10),
		Name: DisplayName(p.FirstName, p.LastName),
	}
}

// snapFromPatient فیلدهای قابل مقایسه را با کد ملی و تلفن ماسک‌شده می‌سازد.
// p بیمار است. خروجی ورودی audit.Diff است.
func snapFromPatient(p *Patient) patientSnap {
	if p == nil {
		return patientSnap{}
	}
	return patientSnap{
		FirstName:         p.FirstName,
		LastName:          p.LastName,
		NationalCode:      audit.MaskTail(p.NationalCode, 4),
		BirthDate:         p.BirthDate.Format(birthDateLayout),
		Address:           derefString(p.Address),
		HomePhone:         audit.MaskTail(derefString(p.HomePhoneNumber), 4),
		MobilePhone:       audit.MaskTail(derefString(p.MobilePhoneNumber), 4),
		FileNumber:        p.FileNumber,
		Sex:               p.Sex,
		IsForeignNational: p.IsForeignNational,
	}
}

// record رویداد را می‌نویسد و شکست ثبت را به خطای عمومی عملیات تبدیل می‌کند.
// ctx زمینه است. e رویداد آماده است. خروجی nil یعنی ردیف نوشته شد.
// FAU_STG.4 / D-13: صدازننده نباید این خطا را نادیده بگیرد.
func (s *Service) record(ctx context.Context, e audit.Event) error {
	if s.audit == nil {
		return audit.Unavailable(apperror.ErrInternal)
	}
	if ctx == nil {
		ctx = context.Background()
	}
	return audit.Unavailable(s.audit.Record(ctx, e))
}

// recordIntegrity نقض یکپارچگی بیمار را ثبت می‌کند.
// ctx زمینه است. actor فاعل و p بیمار دستکاری‌شده است. خروجی خطای ممیزی است.
func (s *Service) recordIntegrity(ctx context.Context, actor audit.Actor, p *Patient) error {
	return s.record(ctx, audit.Event{
		Type:        audit.EventDataTampering,
		Result:      audit.ResultFailure,
		Reason:      audit.ReasonIntegrityViolation,
		Actor:       actor,
		Target:      patientTarget(p),
		Description: "نقض یکپارچگی بیمار «" + DisplayName(p.FirstName, p.LastName) + "»",
	})
}

// recordCreated ایجاد بیمار را ثبت می‌کند.
// ctx زمینه است. actor سازنده و p بیمار ذخیره‌شده است. کد ملی در هدف نیست.
func (s *Service) recordCreated(ctx context.Context, actor audit.Actor, p *Patient) error {
	return s.record(ctx, audit.Event{
		Type:        audit.EventEntityCreated,
		Result:      audit.ResultSuccess,
		Actor:       actor,
		Target:      patientTarget(p),
		Description: "ایجاد بیمار «" + patientTarget(p).Name + "»",
	})
}

// recordUpdated فقط فیلدهای عوض‌شده را ثبت می‌کند.
// ctx زمینه است. actor ویرایشگر است. before وضعیت قبلی و after وضعیت جدید است.
// اگر فیلدی عوض نشده باشد رویدادی نوشته نمی‌شود. کد ملی و تلفن ماسک شده‌اند.
func (s *Service) recordUpdated(ctx context.Context, actor audit.Actor, before, after *Patient) error {
	details := audit.Diff(snapFromPatient(before), snapFromPatient(after), patientFields)
	if len(details) == 0 {
		return nil
	}
	return s.record(ctx, audit.Event{
		Type:        audit.EventEntityUpdated,
		Result:      audit.ResultSuccess,
		Actor:       actor,
		Target:      patientTarget(after),
		Details:     details,
		Description: "به‌روزرسانی بیمار «" + patientTarget(after).Name + "»",
	})
}

// recordDeleted حذف بیمار را ثبت می‌کند.
// ctx زمینه است. actor حذف‌کننده و p بیمار حذف‌شده است. خروجی خطای ممیزی است.
func (s *Service) recordDeleted(ctx context.Context, actor audit.Actor, p *Patient) error {
	return s.record(ctx, audit.Event{
		Type:        audit.EventEntityDeleted,
		Result:      audit.ResultSuccess,
		Actor:       actor,
		Target:      patientTarget(p),
		Description: "حذف بیمار «" + patientTarget(p).Name + "»",
	})
}
