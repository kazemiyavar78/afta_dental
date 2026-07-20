package reception

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/tpdenta/afta-reception/internal/organization"
	"github.com/tpdenta/afta-reception/internal/patient"
	"github.com/tpdenta/afta-reception/internal/platform/apperror"
	"github.com/tpdenta/afta-reception/internal/platform/security/audit"
	"github.com/tpdenta/afta-reception/internal/regulation"
	"github.com/tpdenta/afta-reception/internal/services"
	"github.com/tpdenta/afta-reception/internal/specialcode"
	"github.com/tpdenta/afta-reception/internal/tariff"
	"github.com/tpdenta/afta-reception/internal/user"
	"github.com/tpdenta/afta-reception/internal/wallet"
	"gorm.io/gorm"
)

// Service لایه منطق کسب‌وکار پذیرش.
type Service struct {
	db             *gorm.DB
	repo           Repository
	audit          *audit.Manager
	patientSvc     *patient.Service
	orgSvc         *organization.Service
	serviceSvc     *services.Service
	tariffSvc      *tariff.Service
	userSvc        *user.Service
	walletSvc      *wallet.Service
	specialCodeSvc *specialcode.Service
	regulationSvc  *regulation.Service
	uploadDir      string
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
	walletSvc *wallet.Service,
	specialCodeSvc *specialcode.Service,
	regulationSvc *regulation.Service,
) *Service {
	return &Service{
		db:             db,
		repo:           NewRepository(db),
		audit:          auditMgr,
		patientSvc:     patientSvc,
		orgSvc:         orgSvc,
		serviceSvc:     serviceSvc,
		tariffSvc:      tariffSvc,
		userSvc:        userSvc,
		walletSvc:      walletSvc,
		specialCodeSvc: specialCodeSvc,
		regulationSvc:  regulationSvc,
		uploadDir:      filepath.Join("uploads", "reception"),
	}
}

// CalculateServices خدمات را بر اساس بیمه‌ها محاسبه می‌کند (بدون ذخیره).
func (s *Service) CalculateServices(req CalculateRequest) (*CalculateResponse, error) {
	return s.CalculateServicesWithActor(req, 0, "")
}

// CalculateServicesWithActor محاسبه خدمات با بررسی یکپارچگی کد خاص.
func (s *Service) CalculateServicesWithActor(req CalculateRequest, actorID int, ip string) (*CalculateResponse, error) {
	if req.InsuranceID == nil && req.AdditionalInsuranceID == nil {
		return nil, apperror.New("E-006", "لطفاً قبل از انتخاب خدمت، یک سازمان (پایه یا تکمیلی) انتخاب کنید.", "no organization selected", 400)
	}
	if len(req.Services) == 0 {
		return &CalculateResponse{Services: []CalculatedServiceLine{}}, nil
	}

	var specialPct uint8
	if req.SpecialCodeID != nil && *req.SpecialCodeID > 0 && s.specialCodeSvc != nil {
		sc, err := s.specialCodeSvc.GetActiveForCalc(*req.SpecialCodeID, actorID, ip)
		if err != nil {
			return nil, err
		}
		if sc != nil {
			specialPct = sc.Percentage
		}
	}

	lines, err := s.calculateLines(req.InsuranceID, req.AdditionalInsuranceID, req.AdditionalInsurancePercentage, req.AdditionalInsuranceCoverage, specialPct, req.Services)
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
		SpecialCodeID:                 req.SpecialCodeID,
		AdditionalInsuranceCoverage:   req.AdditionalInsuranceCoverage,
		AdditionalInsurancePercentage: req.AdditionalInsurancePercentage,
		Services:                      req.Services,
	}
	var calcLines []CalculatedServiceLine
	if len(req.Services) > 0 {
		calcResp, calcErr := s.CalculateServicesWithActor(calcReq, actorUserID, ip)
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
		SpecialCodeID:                   req.SpecialCodeID,
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

	// با Save=true پذیرش و تراکنش صندوق در یک تراکنش دیتابیس ذخیره می‌شوند؛
	// اگر ثبت صندوق شکست بخورد، پذیرش هم rollback می‌شود.
	if req.Save {
		if err := s.db.Transaction(func(tx *gorm.DB) error {
			if err := NewRepository(tx).Create(rec); err != nil {
				return apperror.New("DB_ERROR", "خطا در ایجاد پذیرش.", err.Error(), 500)
			}
			return s.recordCashFundAddedTx(tx, rec, calcLines, actorUserID, ip)
		}); err != nil {
			return nil, err
		}
	} else if err := s.repo.Create(rec); err != nil {
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
		calcResp, calcErr := s.CalculateServicesWithActor(CalculateRequest{
			InsuranceID:                   req.InsuranceID,
			AdditionalInsuranceID:         req.AdditionalInsuranceID,
			SpecialCodeID:                 req.SpecialCodeID,
			AdditionalInsuranceCoverage:   req.AdditionalInsuranceCoverage,
			AdditionalInsurancePercentage: req.AdditionalInsurancePercentage,
			Services:                      req.Services,
		}, actorUserID, ip)
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

	wasSaved := rec.Status == string(ReceptionStatusSaved)
	rec.InsuranceID = req.InsuranceID
	rec.AdditionalInsuranceID = req.AdditionalInsuranceID
	rec.SpecialCodeID = req.SpecialCodeID
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

	svcLines := toReceptionServices(rec.ID, calcLines)
	// وقتی صندوق/کیف پول درگیر است، بروزرسانی پذیرش و تراکنش‌ها اتمیک هستند.
	if req.Save || wasSaved {
		if err := s.db.Transaction(func(tx *gorm.DB) error {
			txRepo := NewRepository(tx)
			if err := txRepo.Update(rec); err != nil {
				return apperror.New("DB_ERROR", "خطا در بروزرسانی پذیرش.", err.Error(), 500)
			}
			if err := txRepo.ReplaceServices(rec.ID, svcLines); err != nil {
				return apperror.New("DB_ERROR", "خطا در بروزرسانی خدمات پذیرش.", err.Error(), 500)
			}
			return s.recordCashFundAddedTx(tx, rec, calcLines, actorUserID, ip)
		}); err != nil {
			return nil, err
		}
	} else {
		if err := s.repo.Update(rec); err != nil {
			return nil, apperror.New("DB_ERROR", "خطا در بروزرسانی پذیرش.", err.Error(), 500)
		}
		if err := s.repo.ReplaceServices(rec.ID, svcLines); err != nil {
			return nil, apperror.New("DB_ERROR", "خطا در بروزرسانی خدمات پذیرش.", err.Error(), 500)
		}
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
	wasSaved := rec.Status == string(ReceptionStatusSaved)
	if wasSaved {
		if err := s.recordCashFundRemoved(rec, actorUserID, ip); err != nil {
			return err
		}
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
	nc := strings.TrimSpace(req.Patient.NationalCode)
	if req.Patient.IsForeignNational {
		if len(nc) > 20 {
			return apperror.New("E-002", "کد شناسایی اتباع حداکثر ۲۰ کاراکتر است.", "foreign id too long", 400)
		}
	} else if nc != "" {
		if !patient.IsValidIranianNationalCode(nc) {
			return apperror.New("E-002", "کد ملی ایرانی نامعتبر است.", "invalid national code", 400)
		}
	} else if requireComplete {
		return apperror.New("E-002", "کد ملی الزامی است.", "national code required", 400)
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
	if req.SpecialCodeID != nil && *req.SpecialCodeID > 0 && s.specialCodeSvc != nil {
		if _, err := s.specialCodeSvc.Get(*req.SpecialCodeID); err != nil {
			return apperror.New("E-014", "کد خاص یافت نشد.", err.Error(), 404)
		}
	}
	return nil
}

// validateDoctor فعال بودن و نوع کاربر پزشک را بررسی می‌کند.
func (s *Service) validateDoctor(doctorID uint) error {
	u, err := s.userSvc.GetUserByID(int(doctorID), int(doctorID), true)
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
	u, err := s.userSvc.GetUserByID(int(assistantID), int(assistantID), true)
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
		if !input.IsForeignNational || input.FileNumber == "" || input.FirstName == "" || input.LastName == "" || input.BirthDate == "" {
			return 0, apperror.New("BAD_REQUEST", "برای ایجاد بیمار جدید فیلدهای الزامی ناقص است.", "incomplete patient", 400)
		}
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
		IsForeignNational: input.IsForeignNational,
	}, actorID, ip)
	if err != nil {
		return 0, err
	}
	return created.ID, nil
}

// calculateLines مبالغ هر سطر خدمت را بر اساس سه حالت بیمه و درصد کد خاص محاسبه می‌کند.
func (s *Service) calculateLines(
	baseOrgID, suppOrgID *uint,
	suppPct *uint8,
	suppCoverage *int64,
	specialPct uint8,
	inputs []ServiceLineInput,
) ([]CalculatedServiceLine, error) {
	pct := 0.0
	if suppPct != nil {
		pct = float64(*suppPct)
	}
	pct = 100.0 - pct

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
			if pct == 100.0 {
				suppShare = t.SupplementaryShare * int64(qty)
			} else {
				suppShare = int64(float64(amount) * (pct / 100.0))
			}
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
		found := amount - orgShare - suppShare - subsidy

		// محاسبه یارانه
		// اگر صندوق 0 باشد یارانه هم 0 است
		if found == 0 {
			subsidy = 0
		} else if found < 0 {
			subsidy = amount - orgShare - suppShare
		}

		// کد خاص: درصد از سهم صندوق مثبت به یارانه اضافه می‌شود
		found = amount - orgShare - suppShare - subsidy
		if specialPct > 0 && found > 0 {
			specialShare := int64(float64(found) * (float64(specialPct) / 100.0))
			subsidy += specialShare
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
	suppRaw := int64(float64(rate) * (pct / 100.0))
	floor := rate - orgShare
	if suppRaw < floor {
		return suppRaw
	}
	return floor
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
		SpecialCodeID:                 rec.SpecialCodeID,
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
		ReceptionEnded:                rec.ReceptionEnded,
		PhotoCount:                    len(rec.Photos),
		RegisteredByID:                rec.RegisteredByID,
		Services:                      svcResp,
		Deleted:                       rec.DeletedAt.Valid,
	}

	if len(rec.Photos) > 0 {
		photos := make([]ReceptionPhotoResponse, 0, len(rec.Photos))
		for _, p := range rec.Photos {
			photos = append(photos, ReceptionPhotoResponse{
				ID:       p.ID,
				FileName: p.FileName,
				URL:      "/api/reception/" + fmt.Sprintf("%d", rec.ID) + "/photos/" + fmt.Sprintf("%d", p.ID),
			})
		}
		resp.Photos = photos
		resp.PhotoCount = len(photos)
	} else if n, err := s.repo.CountPhotos(rec.ID); err == nil {
		resp.PhotoCount = int(n)
	}

	if rec.SpecialCodeID != nil && s.specialCodeSvc != nil {
		if sc, err := s.specialCodeSvc.Get(*rec.SpecialCodeID); err == nil {
			resp.SpecialCodeName = sc.Name
			resp.SpecialCodeValue = sc.Code
		}
	}

	if p, err := s.patientSvc.Get(rec.PatientID); err == nil {
		resp.Patient = p
	}
	if rec.DoctorID != nil {
		if d, err := s.userSvc.GetUserByID(int(*rec.DoctorID), int(*rec.DoctorID), true); err == nil {
			resp.DoctorName = d.Name + " " + d.Family
			resp.DoctorMedicalCode = d.MedicalCode
		}
	}
	if rec.AssistantID != nil {
		if a, err := s.userSvc.GetUserByID(int(*rec.AssistantID), int(*rec.AssistantID), true); err == nil {
			resp.AssistantName = a.Name + " " + a.Family
		}
	}
	return resp, nil
}

// recordCashFundAdded مبلغ سهم صندوق پذیرش ذخیره‌شده را در دفتر کل ثبت می‌کند.
func (s *Service) recordCashFundAdded(rec *Reception, lines []CalculatedServiceLine, actorID int, ip string) error {
	return s.recordCashFundAddedTx(nil, rec, lines, actorID, ip)
}

// recordCashFundAddedTx ثبت صندوق را داخل تراکنش بیرونی (در صورت وجود) انجام می‌دهد.
func (s *Service) recordCashFundAddedTx(tx *gorm.DB, rec *Reception, lines []CalculatedServiceLine, actorID int, ip string) error {
	if s.walletSvc == nil {
		return nil
	}
	amount, svcLines := cashFromCalcLines(lines, rec.Discount)
	var err error
	if tx != nil {
		err = s.walletSvc.RecordReceptionServiceAddedTx(tx, rec.PatientID, rec.ID, amount, svcLines, actorID, ip)
	} else {
		err = s.walletSvc.RecordReceptionServiceAdded(rec.PatientID, rec.ID, amount, svcLines, actorID, ip)
	}
	if err != nil {
		if ae, ok := err.(*apperror.AppError); ok {
			return ae
		}
		return apperror.New("WALLET_ERROR", "خطا در ثبت تراکنش صندوق پذیرش.", err.Error(), 500)
	}
	return nil
}

// recordCashFundRemoved مبلغ سهم صندوق پذیرش حذف‌شده را از دفتر کل برمی‌گرداند.
func (s *Service) recordCashFundRemoved(rec *Reception, actorID int, ip string) error {
	if s.walletSvc == nil {
		return nil
	}
	amount, svcLines := cashFromReceptionServices(rec.Services, rec.Discount)
	if err := s.walletSvc.RecordReceptionServiceRemoved(rec.PatientID, rec.ID, amount, svcLines, actorID, ip); err != nil {
		if ae, ok := err.(*apperror.AppError); ok {
			return ae
		}
		return apperror.New("WALLET_ERROR", "خطا در برگشت تراکنش صندوق پذیرش.", err.Error(), 500)
	}
	return nil
}

// cashFromCalcLines جمع سهم صندوق و خطوط توضیح را از نتیجه محاسبه می‌سازد.
func cashFromCalcLines(lines []CalculatedServiceLine, discount int64) (int64, []wallet.ReceptionServiceLine) {
	out := make([]wallet.ReceptionServiceLine, 0, len(lines))
	var total int64
	for _, l := range lines {
		cash := l.ServiceAmount - l.ServiceOrganizationShare - l.ServiceSupplementaryInsuranceShare - l.ServiceSubsidyShare
		if cash < 0 {
			cash = 0
		}
		total += cash
		out = append(out, wallet.ReceptionServiceLine{
			ServiceCode: l.ServiceCode,
			ServiceName: l.ServiceName,
			CashAmount:  cash,
		})
	}
	total -= discount
	if total < 0 {
		total = 0
	}
	return total, out
}

// cashFromReceptionServices جمع سهم صندوق و خطوط توضیح را از خدمات ذخیره‌شده می‌سازد.
func cashFromReceptionServices(services []ReceptionService, discount int64) (int64, []wallet.ReceptionServiceLine) {
	out := make([]wallet.ReceptionServiceLine, 0, len(services))
	var total int64
	for _, svc := range services {
		cash := svc.ServiceAmount - svc.ServiceOrganizationShare - svc.ServiceSupplementaryInsuranceShare - svc.ServiceSubsidyShare
		if cash < 0 {
			cash = 0
		}
		total += cash
		out = append(out, wallet.ReceptionServiceLine{
			ServiceCode: fmt.Sprintf("%d", svc.ServiceID),
			ServiceName: svc.ServiceName,
			CashAmount:  cash,
		})
	}
	total -= discount
	if total < 0 {
		total = 0
	}
	return total, out
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

// ListByPatient پذیرش‌های یک پرونده را برمی‌گرداند.
func (s *Service) ListByPatient(patientID uint) ([]ReceptionResponse, error) {
	list, err := s.repo.FindByPatientID(patientID)
	if err != nil {
		return nil, apperror.New("DB_ERROR", "خطا در خواندن پذیرش‌های پرونده.", err.Error(), 500)
	}
	out := make([]ReceptionResponse, 0, len(list))
	for i := range list {
		resp, rErr := s.toResponse(&list[i])
		if rErr != nil {
			return nil, rErr
		}
		out = append(out, *resp)
	}
	return out, nil
}

// PatientServiceHistory تاریخچه خدمات پرونده از اولین پذیرش تا الان را برمی‌گرداند.
func (s *Service) PatientServiceHistory(patientID uint) ([]PatientServiceHistoryItem, error) {
	list, err := s.repo.FindByPatientID(patientID)
	if err != nil {
		return nil, apperror.New("DB_ERROR", "خطا در خواندن تاریخچه خدمات.", err.Error(), 500)
	}
	out := make([]PatientServiceHistoryItem, 0, len(list))
	for _, rec := range list {
		cash, _ := cashFromReceptionServices(rec.Services, rec.Discount)
		names := make([]string, 0, len(rec.Services))
		for _, svc := range rec.Services {
			names = append(names, svc.ServiceName)
		}
		item := PatientServiceHistoryItem{
			ReceptionID:   rec.ID,
			ReceptionDate: rec.ReceptionDate.Format("2006-01-02"),
			CashAmount:    cash,
			ServiceNames:  names,
		}
		if rec.InsuranceID != nil {
			if org, oErr := s.orgSvc.Get(*rec.InsuranceID); oErr == nil {
				item.InsuranceName = org.Name
			}
		}
		if rec.AdditionalInsuranceID != nil {
			if org, oErr := s.orgSvc.Get(*rec.AdditionalInsuranceID); oErr == nil {
				item.AdditionalInsuranceName = org.Name
			}
		}
		out = append(out, item)
	}
	return out, nil
}

// EndReception پایان پذیرش را با بررسی ضوابط و عکس‌ها انجام می‌دهد.
func (s *Service) EndReception(id uint, actorID int, ip string) (*EndReceptionResponse, error) {
	rec, err := s.repo.FindByID(id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperror.ErrNotFound
	}
	if err != nil {
		return nil, apperror.New("DB_ERROR", "خطا در خواندن پذیرش.", err.Error(), 500)
	}
	if rec.Status != string(ReceptionStatusSaved) {
		return nil, apperror.New("E-015", "فقط پذیرش ذخیره‌شده قابل پایان است.", "not saved", 400)
	}
	if rec.ReceptionEnded {
		return &EndReceptionResponse{
			Success:            true,
			ReceptionEnded:     true,
			Message:            "پذیرش قبلاً پایان یافته است.",
			UploadedPhotoCount: len(rec.Photos),
		}, nil
	}

	prev, prevErr := s.repo.FindPreviousForPatient(rec.PatientID, rec.ID)
	if prevErr == nil && prev != nil && !prev.ReceptionEnded {
		prevID := prev.ID
		return &EndReceptionResponse{
			Success:             false,
			ReceptionEnded:      false,
			PreviousReceptionID: &prevID,
			Message:             fmt.Sprintf("پذیرش قبلی شماره %d هنوز پایان نیافته است.", prevID),
		}, nil
	} else if prevErr != nil && !errors.Is(prevErr, gorm.ErrRecordNotFound) {
		return nil, apperror.New("DB_ERROR", "خطا در خواندن پذیرش قبلی.", prevErr.Error(), 500)
	}

	requiredPhotos := 0
	var descriptions []string
	if s.regulationSvc != nil {
		regs, rErr := s.regulationSvc.ListActive()
		if rErr != nil {
			return nil, rErr
		}
		for _, reg := range regs {
			if len(reg.ServiceIDs) == 0 {
				continue
			}
			from := rec.ReceptionDate.AddDate(0, 0, -reg.DurationDays)
			window, wErr := s.repo.FindPatientReceptionsInRange(rec.PatientID, from, rec.ReceptionDate)
			if wErr != nil {
				return nil, apperror.New("DB_ERROR", "خطا در بررسی ضوابط.", wErr.Error(), 500)
			}
			foundServices := map[uint]bool{}
			for _, wr := range window {
				for _, svc := range wr.Services {
					foundServices[svc.ServiceID] = true
				}
			}
			allMatched := true
			for _, sid := range reg.ServiceIDs {
				if !foundServices[sid] {
					allMatched = false
					break
				}
			}
			if allMatched {
				descriptions = append(descriptions, reg.Description)
				if reg.PhotoCount > requiredPhotos {
					requiredPhotos = reg.PhotoCount
				}
			}
		}
	}

	uploaded, _ := s.repo.CountPhotos(rec.ID)
	if requiredPhotos > 0 && int(uploaded) < requiredPhotos {
		return &EndReceptionResponse{
			Success:                false,
			ReceptionEnded:         false,
			RegulationDescriptions: descriptions,
			RequiredPhotoCount:     requiredPhotos,
			UploadedPhotoCount:     int(uploaded),
			Message:                "برای پایان پذیرش نیاز به آپلود عکس دندان است.",
		}, nil
	}

	rec.ReceptionEnded = true
	if err := s.repo.Update(rec); err != nil {
		return nil, apperror.New("DB_ERROR", "خطا در ثبت پایان پذیرش.", err.Error(), 500)
	}
	_ = s.audit.LogEvent(&actorID, ip, audit.EventUserDataChange, fmt.Sprintf("پایان پذیرش %d", rec.ID))
	return &EndReceptionResponse{
		Success:                true,
		ReceptionEnded:         true,
		RegulationDescriptions: descriptions,
		RequiredPhotoCount:     requiredPhotos,
		UploadedPhotoCount:     int(uploaded),
		Message:                "پایان پذیرش با موفقیت ثبت شد.",
	}, nil
}

// UploadReceptionPhoto عکس دندان را برای پذیرش ذخیره می‌کند.
func (s *Service) UploadReceptionPhoto(id uint, fileName string, reader io.Reader, actorID int, ip string) (*ReceptionPhotoResponse, error) {
	rec, err := s.repo.FindByID(id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperror.ErrNotFound
	}
	if err != nil {
		return nil, apperror.New("DB_ERROR", "خطا در خواندن پذیرش.", err.Error(), 500)
	}
	if rec.ReceptionEnded {
		return nil, apperror.New("E-016", "پذیرش پایان‌یافته و امکان آپلود عکس ندارد.", "already ended", 400)
	}

	dir := filepath.Join(s.uploadDir, fmt.Sprintf("%d", id))
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, apperror.New("IO_ERROR", "خطا در ایجاد پوشه آپلود.", err.Error(), 500)
	}
	ext := filepath.Ext(fileName)
	stored := uuid.New().String() + ext
	fullPath := filepath.Join(dir, stored)
	f, err := os.Create(fullPath)
	if err != nil {
		return nil, apperror.New("IO_ERROR", "خطا در ذخیره فایل.", err.Error(), 500)
	}
	defer f.Close()
	if _, err := io.Copy(f, reader); err != nil {
		return nil, apperror.New("IO_ERROR", "خطا در نوشتن فایل.", err.Error(), 500)
	}

	photo := &ReceptionPhoto{
		ReceptionID: id,
		FilePath:    fullPath,
		FileName:    fileName,
	}
	if err := s.repo.AddPhoto(photo); err != nil {
		return nil, apperror.New("DB_ERROR", "خطا در ثبت عکس پذیرش.", err.Error(), 500)
	}
	_ = s.audit.LogEvent(&actorID, ip, audit.EventUpload, fmt.Sprintf("آپلود عکس پذیرش %d", id))
	return &ReceptionPhotoResponse{
		ID:       photo.ID,
		FileName: photo.FileName,
		URL:      fmt.Sprintf("/api/reception/%d/photos/%d", id, photo.ID),
	}, nil
}

// GetReceptionPhotoPath مسیر فایل عکس پذیرش را برمی‌گرداند.
func (s *Service) GetReceptionPhotoPath(receptionID, photoID uint) (string, string, error) {
	var photo ReceptionPhoto
	err := s.db.Where("ID = ? AND ReceptionID = ?", photoID, receptionID).First(&photo).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return "", "", apperror.ErrNotFound
	}
	if err != nil {
		return "", "", apperror.New("DB_ERROR", "خطا در خواندن عکس.", err.Error(), 500)
	}
	return photo.FilePath, photo.FileName, nil
}
