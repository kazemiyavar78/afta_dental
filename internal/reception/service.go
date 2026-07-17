package reception

import (
	"errors"
	"fmt"
	"regexp"
	"time"
	"unicode/utf8"

	"github.com/tpdenta/afta-reception/internal/organization"
	"github.com/tpdenta/afta-reception/internal/patient"
	"github.com/tpdenta/afta-reception/internal/platform/apperror"
	"github.com/tpdenta/afta-reception/internal/platform/security/audit"
	"github.com/tpdenta/afta-reception/internal/services"
	"github.com/tpdenta/afta-reception/internal/tariff"
	"github.com/tpdenta/afta-reception/internal/user"
	"gorm.io/gorm"
)

var nationalCodePattern = regexp.MustCompile(`^\d{10}$`)

// Service لایه منطق کسب‌وکار پذیرش.
type Service struct {
	db           *gorm.DB
	repo         Repository
	audit        *audit.Manager
	patientSvc   *patient.Service
	orgSvc       *organization.Service
	serviceSvc   *services.Service
	tariffSvc    *tariff.Service
	userSvc      *user.Service
}

// NewService نمونه Service پذیرش را با وابستگی‌ها می‌سازد.
func NewService(
	db *gorm.DB,
	auditMgr *audit.Manager,
	patientSvc *patient.Service,
	orgSvc *organization.Service,
	serviceSvc *services.Service,
	tariffSvc *tariff.Service,
	userSvc *user.Service,
) *Service {
	return &Service{
		db:         db,
		repo:       NewRepository(db),
		audit:      auditMgr,
		patientSvc: patientSvc,
		orgSvc:     orgSvc,
		serviceSvc: serviceSvc,
		tariffSvc:  tariffSvc,
		userSvc:    userSvc,
	}
}

// CalculateServices خدمات را بر اساس بیمه‌ها محاسبه می‌کند (بدون ذخیره).
func (s *Service) CalculateServices(req CalculateRequest) (*CalculateResponse, error) {
	if req.InsuranceID == nil && req.AdditionalInsuranceID == nil {
		return nil, apperror.New("E-006", "لطفاً قبل از انتخاب خدمت، یک سازمان (پایه یا تکمیلی) انتخاب کنید.", "no organization selected", 400)
	}
	if len(req.Services) == 0 {
		return &CalculateResponse{Services: []CalculatedServiceLine{}}, nil
	}

	fmt.Println("req.Services", req.Services)
	lines, err := s.calculateLines(req.InsuranceID, req.AdditionalInsuranceID, req.AdditionalInsurancePercentage, req.AdditionalInsuranceCoverage, req.Services)
	if err != nil {
		return nil, err
	}
	return &CalculateResponse{Services: lines}, nil
}

// CreateReception پذیرش جدید ایجاد می‌کند؛ در صورت Save=true وضعیت saved می‌شود.
func (s *Service) CreateReception(req UpsertReceptionRequest, actorUserID int, ip string) (*ReceptionResponse, error) {
	if err := s.validateUpsert(req, req.Save); err != nil {
		return nil, err
	}

	patientID, err := s.resolvePatient(req.Patient, actorUserID, ip, true)
	if err != nil {
		return nil, err
	}

	calcReq := CalculateRequest{
		InsuranceID:                   req.InsuranceID,
		AdditionalInsuranceID:         req.AdditionalInsuranceID,
		AdditionalInsuranceCoverage:   req.AdditionalInsuranceCoverage,
		AdditionalInsurancePercentage: req.AdditionalInsurancePercentage,
		Services:                      req.Services,
	}
	var calcLines []CalculatedServiceLine
	if len(req.Services) > 0 {
		calcResp, calcErr := s.CalculateServices(calcReq)
		if calcErr != nil {
			return nil, calcErr
		}
		calcLines = calcResp.Services
	}
	if err := validateNonNegativeCash(calcLines); err != nil {
		return nil, err
	}

	receptionDate, err := parseDate(req.ReceptionDate)
	if err != nil {
		return nil, err
	}
	var bookingDate *time.Time
	if req.BookingDate != nil && *req.BookingDate != "" {
		bd, bErr := parseDate(*req.BookingDate)
		if bErr != nil {
			return nil, bErr
		}
		bookingDate = &bd
	}

	status := string(ReceptionStatusDraft)
	if req.Save {
		status = string(ReceptionStatusSaved)
	}
	regID := uint(actorUserID)

	rec := &Reception{
		PatientID:                       patientID,
		InsuranceID:                     req.InsuranceID,
		AdditionalInsuranceID:           req.AdditionalInsuranceID,
		DoctorID:                        req.DoctorID,
		AssistantID:                     req.AssistantID,
		BookingDate:                     bookingDate,
		ReceptionDate:                   receptionDate,
		Status:                          status,
		Description:                     req.Description,
		Discount:                        req.Discount,
		ReferralCode:                    req.ReferralCode,
		AdditionalInsuranceCoverage:     req.AdditionalInsuranceCoverage,
		AdditionalInsurancePercentage:   req.AdditionalInsurancePercentage,
		RegisteredByID:                  &regID,
		Services:                        toReceptionServices(0, calcLines),
	}

	if err := s.repo.Create(rec); err != nil {
		return nil, apperror.New("DB_ERROR", "خطا در ایجاد پذیرش.", err.Error(), 500)
	}

	_ = s.audit.LogEvent(&actorUserID, ip, audit.EventUserDataChange,
		fmt.Sprintf("ایجاد پذیرش %d", rec.ID))

	return s.toResponse(rec)
}

// UpdateReception پذیرش موجود را ویرایش می‌کند (فقط در صورت عدم حذف).
func (s *Service) UpdateReception(id uint, req UpsertReceptionRequest, actorUserID int, ip string) (*ReceptionResponse, error) {
	rec, err := s.repo.FindByIDUnscoped(id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperror.ErrNotFound
	}
	if err != nil {
		return nil, apperror.New("DB_ERROR", "خطا در خواندن پذیرش.", err.Error(), 500)
	}
	if rec.DeletedAt.Valid {
		return nil, apperror.New("E-012", "پذیرش حذف شده و قابل ویرایش نیست.", "deleted reception", 400)
	}

	if err := s.validateUpsert(req, req.Save || rec.Status == string(ReceptionStatusSaved)); err != nil {
		return nil, err
	}

	patientID, err := s.resolvePatient(req.Patient, actorUserID, ip, false)
	if err != nil {
		return nil, err
	}
	if patientID != rec.PatientID {
		return nil, apperror.New("E-011", "تغییر بیمار پذیرش مجاز نیست.", "patient change not allowed", 400)
	}

	var calcLines []CalculatedServiceLine
	if len(req.Services) > 0 {
		calcResp, calcErr := s.CalculateServices(CalculateRequest{
			InsuranceID:                   req.InsuranceID,
			AdditionalInsuranceID:         req.AdditionalInsuranceID,
			AdditionalInsuranceCoverage:   req.AdditionalInsuranceCoverage,
			AdditionalInsurancePercentage: req.AdditionalInsurancePercentage,
			Services:                      req.Services,
		})
		if calcErr != nil {
			return nil, calcErr
		}
		calcLines = calcResp.Services
	}
	if err := validateNonNegativeCash(calcLines); err != nil {
		return nil, err
	}

	receptionDate, err := parseDate(req.ReceptionDate)
	if err != nil {
		return nil, err
	}
	var bookingDate *time.Time
	if req.BookingDate != nil && *req.BookingDate != "" {
		bd, bErr := parseDate(*req.BookingDate)
		if bErr != nil {
			return nil, bErr
		}
		bookingDate = &bd
	}

	rec.InsuranceID = req.InsuranceID
	rec.AdditionalInsuranceID = req.AdditionalInsuranceID
	rec.DoctorID = req.DoctorID
	rec.AssistantID = req.AssistantID
	rec.BookingDate = bookingDate
	rec.ReceptionDate = receptionDate
	rec.Description = req.Description
	rec.Discount = req.Discount
	rec.ReferralCode = req.ReferralCode
	rec.AdditionalInsuranceCoverage = req.AdditionalInsuranceCoverage
	rec.AdditionalInsurancePercentage = req.AdditionalInsurancePercentage
	if req.Save {
		rec.Status = string(ReceptionStatusSaved)
	}

	if err := s.repo.Update(rec); err != nil {
		return nil, apperror.New("DB_ERROR", "خطا در بروزرسانی پذیرش.", err.Error(), 500)
	}
	if err := s.repo.ReplaceServices(rec.ID, toReceptionServices(rec.ID, calcLines)); err != nil {
		return nil, apperror.New("DB_ERROR", "خطا در بروزرسانی خدمات پذیرش.", err.Error(), 500)
	}

	_ = s.audit.LogEvent(&actorUserID, ip, audit.EventUserDataChange,
		fmt.Sprintf("ویرایش پذیرش %d", rec.ID))

	updated, err := s.repo.FindByID(rec.ID)
	if err != nil {
		return nil, apperror.New("DB_ERROR", "خطا در خواندن پذیرش.", err.Error(), 500)
	}
	return s.toResponse(updated)
}

// GetReception پذیرش را با شناسه برمی‌گرداند.
func (s *Service) GetReception(id uint) (*ReceptionResponse, error) {
	rec, err := s.repo.FindByIDUnscoped(id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperror.ErrNotFound
	}
	if err != nil {
		return nil, apperror.New("DB_ERROR", "خطا در خواندن پذیرش.", err.Error(), 500)
	}
	return s.toResponse(rec)
}

// ListReceptions لیست پذیرش‌های غیرحذف‌شده را برمی‌گرداند.
func (s *Service) ListReceptions() ([]ReceptionResponse, error) {
	list, err := s.repo.FindAll()
	if err != nil {
		return nil, apperror.New("DB_ERROR", "خطا در خواندن پذیرش‌ها.", err.Error(), 500)
	}
	result := make([]ReceptionResponse, 0, len(list))
	for i := range list {
		resp, err := s.toResponse(&list[i])
		if err != nil {
			return nil, err
		}
		result = append(result, *resp)
	}
	return result, nil
}

// Navigate به اولین/قبلی/بعدی/آخرین پذیرش می‌رود.
// اگر هیچ پذیرشی وجود نداشته باشد، پاسخ empty=true با وضعیت ۲۰۰ برمی‌گرداند (نه 404).
func (s *Service) Navigate(cursor *uint, dir string) (*ReceptionResponse, error) {
	var (
		rec *Reception
		err error
	)
	switch dir {
	case "first":
		rec, err = s.repo.FindFirst()
	case "last":
		rec, err = s.repo.FindLast()
	case "prev":
		if cursor == nil {
			return nil, apperror.New("BAD_REQUEST", "cursor برای ناوبری قبلی الزامی است.", "cursor required", 400)
		}
		rec, err = s.repo.FindPrev(*cursor)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			rec, err = s.repo.FindByID(*cursor)
		}
	case "next":
		if cursor == nil {
			return nil, apperror.New("BAD_REQUEST", "cursor برای ناوبری بعدی الزامی است.", "cursor required", 400)
		}
		rec, err = s.repo.FindNext(*cursor)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			rec, err = s.repo.FindByID(*cursor)
		}
	default:
		return nil, apperror.New("BAD_REQUEST", "جهت ناوبری نامعتبر است.", "invalid nav dir", 400)
	}
	if errors.Is(err, gorm.ErrRecordNotFound) || rec == nil {
		// هیچ پذیرشی ثبت نشده — فرم خالی برای کاربر
		if dir == "first" || dir == "last" {
			return &ReceptionResponse{Empty: true, Status: string(ReceptionStatusDraft), Services: []ReceptionServiceResponse{}}, nil
		}
		return nil, apperror.ErrNotFound
	}
	if err != nil {
		return nil, apperror.New("DB_ERROR", "خطا در ناوبری پذیرش.", err.Error(), 500)
	}
	return s.toResponse(rec)
}

// SoftDelete پذیرش را به‌صورت نرم حذف می‌کند.
func (s *Service) SoftDelete(id uint, actorUserID int, ip string) error {
	rec, err := s.repo.FindByID(id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return apperror.ErrNotFound
	}
	if err != nil {
		return apperror.New("DB_ERROR", "خطا در خواندن پذیرش.", err.Error(), 500)
	}
	if err := s.repo.Delete(rec.ID); err != nil {
		return apperror.New("DB_ERROR", "خطا در حذف پذیرش.", err.Error(), 500)
	}
	_ = s.audit.LogEvent(&actorUserID, ip, audit.EventUserDataChange,
		fmt.Sprintf("حذف نرم پذیرش %d", id))
	return nil
}

// Restore پذیرش حذف‌شده را بازیابی می‌کند.
func (s *Service) Restore(id uint, actorUserID int, ip string) (*ReceptionResponse, error) {
	rec, err := s.repo.FindByIDUnscoped(id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperror.ErrNotFound
	}
	if err != nil {
		return nil, apperror.New("DB_ERROR", "خطا در خواندن پذیرش.", err.Error(), 500)
	}
	if !rec.DeletedAt.Valid {
		return s.toResponse(rec)
	}
	if err := s.repo.Restore(id); err != nil {
		return nil, apperror.New("DB_ERROR", "خطا در بازیابی پذیرش.", err.Error(), 500)
	}
	_ = s.audit.LogEvent(&actorUserID, ip, audit.EventUserDataChange,
		fmt.Sprintf("بازیابی پذیرش %d", id))
	updated, err := s.repo.FindByID(id)
	if err != nil {
		return nil, apperror.New("DB_ERROR", "خطا در خواندن پذیرش.", err.Error(), 500)
	}
	return s.toResponse(updated)
}

// validateUpsert اعتبارسنجی درخواست ایجاد/ویرایش را انجام می‌دهد.
func (s *Service) validateUpsert(req UpsertReceptionRequest, requireComplete bool) error {
	if req.Patient.NationalCode != "" && !nationalCodePattern.MatchString(req.Patient.NationalCode) {
		return apperror.New("E-002", "کد ملی نامعتبر است.", "invalid national code", 400)
	}
	if requireComplete {
		if req.InsuranceID == nil && req.AdditionalInsuranceID == nil {
			return apperror.New("E-006", "لطفاً قبل از انتخاب خدمت، یک سازمان (پایه یا تکمیلی) انتخاب کنید.", "no organization", 400)
		}
		if req.DoctorID == nil {
			return apperror.New("E-007", "انتخاب پزشک الزامی است.", "doctor required", 400)
		}
		if len(req.Services) == 0 {
			return apperror.New("E-008", "حداقل یک خدمت باید ثبت شود.", "services required", 400)
		}
		if utf8.RuneCountInString(req.Patient.FirstName) == 0 || utf8.RuneCountInString(req.Patient.LastName) == 0 {
			return apperror.New("BAD_REQUEST", "نام و نام خانوادگی بیمار الزامی است.", "patient name required", 400)
		}
	}
	if req.DoctorID != nil {
		if err := s.validateDoctor(*req.DoctorID); err != nil {
			return err
		}
	}
	if req.AssistantID != nil {
		if err := s.validateAssistant(*req.AssistantID); err != nil {
			return err
		}
	}
	if req.InsuranceID != nil {
		if _, err := s.orgSvc.Get(*req.InsuranceID); err != nil {
			return apperror.New("E-004", "بیمه پایه یافت نشد.", err.Error(), 404)
		}
	}
	if req.AdditionalInsuranceID != nil {
		if _, err := s.orgSvc.Get(*req.AdditionalInsuranceID); err != nil {
			return apperror.New("E-004", "بیمه تکمیلی یافت نشد.", err.Error(), 404)
		}
	}
	return nil
}

// validateDoctor فعال بودن و نوع کاربر پزشک را بررسی می‌کند.
func (s *Service) validateDoctor(doctorID uint) error {
	u, err := s.userSvc.GetUserByID(int(doctorID), int(doctorID), "Admin")
	if err != nil {
		return apperror.New("E-003", "پزشک یافت نشد.", err.Error(), 404)
	}
	if u.UserType != user.UserTypeDoctor {
		return apperror.New("E-003", "کاربر انتخاب‌شده پزشک نیست.", "not a doctor", 400)
	}
	if !u.IsActive {
		return apperror.New("E-003", "پزشک غیرفعال است.", "doctor inactive", 400)
	}
	return nil
}

// validateAssistant نوع کاربر دستیار را بررسی می‌کند.
func (s *Service) validateAssistant(assistantID uint) error {
	u, err := s.userSvc.GetUserByID(int(assistantID), int(assistantID), "Admin")
	if err != nil {
		return apperror.New("BAD_REQUEST", "دستیار یافت نشد.", err.Error(), 404)
	}
	if u.UserType != user.UserTypeAssistant {
		return apperror.New("BAD_REQUEST", "کاربر انتخاب‌شده دستیار نیست.", "not an assistant", 400)
	}
	return nil
}

// resolvePatient بیمار موجود را پیدا می‌کند یا در صورت نیاز ایجاد می‌کند.
func (s *Service) resolvePatient(input PatientInput, actorID int, ip string, allowCreate bool) (uint, error) {
	if input.ID != nil && *input.ID > 0 {
		p, err := s.patientSvc.Get(*input.ID)
		if err != nil {
			return 0, apperror.New("E-001", "بیمار یافت نشد.", err.Error(), 404)
		}
		return p.ID, nil
	}
	if input.NationalCode != "" {
		if p, err := s.patientSvc.GetByNationalCode(input.NationalCode); err == nil {
			return p.ID, nil
		} else if !isNotFound(err) {
			return 0, err
		}
	}
	if input.FileNumber != "" {
		if p, err := s.patientSvc.GetByFileNumber(input.FileNumber); err == nil {
			return p.ID, nil
		} else if !isNotFound(err) {
			return 0, err
		}
	}
	if !allowCreate {
		return 0, apperror.New("E-001", "بیمار یافت نشد.", "patient not found", 404)
	}
	if input.NationalCode == "" || input.FileNumber == "" || input.FirstName == "" || input.LastName == "" || input.BirthDate == "" {
		return 0, apperror.New("BAD_REQUEST", "برای ایجاد بیمار جدید فیلدهای الزامی ناقص است.", "incomplete patient", 400)
	}
	created, err := s.patientSvc.Create(patient.CreateRequest{
		FirstName:         input.FirstName,
		LastName:          input.LastName,
		NationalCode:      input.NationalCode,
		BirthDate:         input.BirthDate,
		Address:           input.Address,
		HomePhoneNumber:   input.HomePhoneNumber,
		MobilePhoneNumber: input.MobilePhoneNumber,
		FileNumber:        input.FileNumber,
		Sex:               input.Sex,
	}, actorID, ip)
	if err != nil {
		return 0, err
	}
	return created.ID, nil
}

// calculateLines مبالغ هر سطر خدمت را بر اساس سه حالت بیمه محاسبه می‌کند.
func (s *Service) calculateLines(
	baseOrgID, suppOrgID *uint,
	suppPct *uint8,
	suppCoverage *int64,
	inputs []ServiceLineInput,
) ([]CalculatedServiceLine, error) {
	pct := 0.0
	if suppPct != nil {
		pct = float64(*suppPct) / 100.0
	}

	var remainingCoverage *int64
	if suppCoverage != nil {
		v := *suppCoverage
		remainingCoverage = &v
	}

	result := make([]CalculatedServiceLine, 0, len(inputs))
	for _, in := range inputs {
		item, err := s.resolveServiceItem(in)
		if err != nil {
			return nil, err
		}
		qty := in.Quantity
		if qty <= 0 {
			qty = item.DefaultCount
		}
		if qty <= 0 {
			qty = 1
		}

		var amount, tariffAmount, orgShare, suppShare, subsidy int64

		hasBase := baseOrgID != nil
		hasSupp := suppOrgID != nil

		switch {
		case hasBase && !hasSupp:
			t, tErr := s.tariffSvc.GetByOrganizationAndService(*baseOrgID, item.ID)
			if tErr != nil {
				return nil, tErr
			}
			amount = t.Amount * int64(qty)
			tariffAmount = t.TariffAmount * int64(qty)
			orgShare = t.OrganizationShare * int64(qty)
			subsidy = t.SubsidyShare * int64(qty)
			suppShare = 0

		case !hasBase && hasSupp:
			t, tErr := s.tariffSvc.GetByOrganizationAndService(*suppOrgID, item.ID)
			if tErr != nil {
				return nil, tErr
			}
			amount = t.Amount * int64(qty)
			tariffAmount = t.TariffAmount * int64(qty)
			orgShare = 0
			subsidy = t.SubsidyShare * int64(qty)
			suppShare = int64(float64(amount) * pct)
			if remainingCoverage != nil {
				suppShare = clampToCoverage(suppShare, remainingCoverage)
			}

		case hasBase && hasSupp:
			baseTariff, tErr := s.tariffSvc.GetByOrganizationAndService(*baseOrgID, item.ID)
			if tErr != nil {
				return nil, tErr
			}
			amount = baseTariff.Amount * int64(qty)
			tariffAmount = baseTariff.TariffAmount * int64(qty)
			orgShare = baseTariff.OrganizationShare * int64(qty)
			subsidy = baseTariff.SubsidyShare * int64(qty)
			suppShare = computeSupplementaryShare(amount, orgShare, pct)
			if remainingCoverage != nil {
				suppShare = clampToCoverage(suppShare, remainingCoverage)
			}

		default:
			return nil, apperror.New("E-006", "لطفاً قبل از انتخاب خدمت، یک سازمان (پایه یا تکمیلی) انتخاب کنید.", "no organization", 400)
		}

		result = append(result, CalculatedServiceLine{
			ServiceID:                          item.ID,
			ServiceCode:                        item.ServiceCode,
			ServiceName:                        item.Name,
			Quantity:                           qty,
			ServiceAmount:                      amount,
			ServiceTariff:                      tariffAmount,
			ServiceOrganizationShare:           orgShare,
			ServiceSupplementaryInsuranceShare: suppShare,
			ServiceSubsidyShare:                subsidy,
			ServiceDescription:                 in.ServiceDescription,
			TeethNumber:                        in.TeethNumber,
			TeethDirection:                     in.TeethDirection,
			HasDentalDirection:                 item.HasDentalDirection,
			HasTooth:                           item.HasTooth,
		})
	}
	return result, nil
}

// computeSupplementaryShare سهم بیمه تکمیلی را بر اساس درصد و سهم سازمان محاسبه می‌کند.
func computeSupplementaryShare(rate, orgShare int64, pct float64) int64 {
	suppRaw := int64(float64(rate) * pct)
	floor := orgShare
	if suppRaw < floor {
		return rate - floor
	}
	return suppRaw
}

// validateNonNegativeCash اگر سهم صندوق هر خدمت منفی باشد خطا برمی‌گرداند.
func validateNonNegativeCash(lines []CalculatedServiceLine) error {
	for _, line := range lines {
		cash := line.ServiceAmount - line.ServiceOrganizationShare - line.ServiceSupplementaryInsuranceShare - line.ServiceSubsidyShare
		if cash < 0 {
			return apperror.New(
				"E-013",
				"سهم صندوق یکی از خدمات منفی است؛ امکان ثبت وجود ندارد.",
				"negative cash share",
				400,
			)
		}
	}
	return nil
}

// clampToCoverage سهم تکمیلی را به سقف باقیمانده محدود می‌کند و سقف را کم می‌کند.
func clampToCoverage(share int64, remaining *int64) int64 {
	if remaining == nil {
		return share
	}
	if share > *remaining {
		share = *remaining
	}
	*remaining -= share
	if *remaining < 0 {
		*remaining = 0
	}
	return share
}

// resolveServiceItem خدمت را از شناسه یا کد پیدا می‌کند.
func (s *Service) resolveServiceItem(in ServiceLineInput) (*services.ServiceItem, error) {
	if in.ServiceID > 0 {
		return s.serviceSvc.FindItemByID(in.ServiceID)
	}
	if in.ServiceCode != "" {
		return s.serviceSvc.FindItemByCode(in.ServiceCode)
	}
	return nil, apperror.New("BAD_REQUEST", "کد یا شناسه خدمت الزامی است.", "service required", 400)
}

// toReceptionServices خطوط محاسبه‌شده را به مدل دامنه تبدیل می‌کند.
func toReceptionServices(receptionID uint, lines []CalculatedServiceLine) []ReceptionService {
	out := make([]ReceptionService, 0, len(lines))
	for _, l := range lines {
		out = append(out, ReceptionService{
			ReceptionID:                        receptionID,
			ServiceID:                          l.ServiceID,
			ServiceName:                        l.ServiceName,
			Quantity:                           l.Quantity,
			ServiceAmount:                      l.ServiceAmount,
			ServiceTariff:                      l.ServiceTariff,
			ServiceOrganizationShare:           l.ServiceOrganizationShare,
			ServiceSupplementaryInsuranceShare: l.ServiceSupplementaryInsuranceShare,
			ServiceSubsidyShare:                l.ServiceSubsidyShare,
			ServiceDescription:                 l.ServiceDescription,
			TeethNumber:                        l.TeethNumber,
			TeethDirection:                     l.TeethDirection,
		})
	}
	return out
}

// toResponse مدل دامنه را به DTO پاسخ تبدیل می‌کند.
func (s *Service) toResponse(rec *Reception) (*ReceptionResponse, error) {
	var booking *string
	if rec.BookingDate != nil {
		v := rec.BookingDate.Format("2006-01-02")
		booking = &v
	}

	svcResp := make([]ReceptionServiceResponse, 0, len(rec.Services))
	for _, svc := range rec.Services {
		hasDir, hasTooth := false, false
		if item, err := s.serviceSvc.FindItemByID(svc.ServiceID); err == nil && item != nil {
			hasDir = item.HasDentalDirection
			hasTooth = item.HasTooth
		}
		svcResp = append(svcResp, ReceptionServiceResponse{
			ID:                                 svc.ID,
			ServiceID:                          svc.ServiceID,
			ServiceName:                        svc.ServiceName,
			Quantity:                           svc.Quantity,
			ServiceAmount:                      svc.ServiceAmount,
			ServiceTariff:                      svc.ServiceTariff,
			ServiceOrganizationShare:           svc.ServiceOrganizationShare,
			ServiceSupplementaryInsuranceShare: svc.ServiceSupplementaryInsuranceShare,
			ServiceSubsidyShare:                svc.ServiceSubsidyShare,
			ServiceDescription:                 svc.ServiceDescription,
			TeethNumber:                        svc.TeethNumber,
			TeethDirection:                     svc.TeethDirection,
			HasDentalDirection:                 hasDir,
			HasTooth:                           hasTooth,
		})
	}

	resp := &ReceptionResponse{
		ID:                            rec.ID,
		PatientID:                     rec.PatientID,
		InsuranceID:                   rec.InsuranceID,
		AdditionalInsuranceID:         rec.AdditionalInsuranceID,
		DoctorID:                      rec.DoctorID,
		AssistantID:                   rec.AssistantID,
		BookingDate:                   booking,
		ReceptionDate:                 rec.ReceptionDate.Format("2006-01-02"),
		Status:                        rec.Status,
		Description:                   rec.Description,
		Discount:                      rec.Discount,
		ReferralCode:                  rec.ReferralCode,
		AdditionalInsuranceCoverage:   rec.AdditionalInsuranceCoverage,
		AdditionalInsurancePercentage: rec.AdditionalInsurancePercentage,
		RegisteredByID:                rec.RegisteredByID,
		Services:                      svcResp,
		Deleted:                       rec.DeletedAt.Valid,
	}

	if p, err := s.patientSvc.Get(rec.PatientID); err == nil {
		resp.Patient = p
	}
	if rec.DoctorID != nil {
		if d, err := s.userSvc.GetUserByID(int(*rec.DoctorID), int(*rec.DoctorID), "Admin"); err == nil {
			resp.DoctorName = d.Name + " " + d.Family
			resp.DoctorMedicalCode = d.MedicalCode
		}
	}
	if rec.AssistantID != nil {
		if a, err := s.userSvc.GetUserByID(int(*rec.AssistantID), int(*rec.AssistantID), "Admin"); err == nil {
			resp.AssistantName = a.Name + " " + a.Family
		}
	}
	return resp, nil
}

// parseDate تاریخ YYYY-MM-DD را پارس می‌کند.
func parseDate(value string) (time.Time, error) {
	t, err := time.Parse("2006-01-02", value)
	if err != nil {
		return time.Time{}, apperror.New("INVALID_DATE", "فرمت تاریخ نامعتبر است.", err.Error(), 400)
	}
	return t, nil
}

// isNotFound بررسی می‌کند خطا از نوع یافت‌نشدن باشد.
func isNotFound(err error) bool {
	var appErr *apperror.AppError
	if errors.As(err, &appErr) {
		return appErr.Code == "NOT_FOUND"
	}
	return errors.Is(err, gorm.ErrRecordNotFound)
}
