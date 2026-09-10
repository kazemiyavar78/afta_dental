package report

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/tpdenta/afta-reception/internal/organization"
	"github.com/tpdenta/afta-reception/internal/platform/apperror"
	"gorm.io/gorm"
)

// Service لایه منطق گزارش‌ها.
type Service struct {
	repo    Repository
	orgRepo organization.Repository
}

// NewService نمونه Service گزارش می‌سازد.
func NewService(db *gorm.DB) *Service {
	return &Service{
		repo:    NewRepository(db),
		orgRepo: organization.NewRepository(db),
	}
}

// parseUintList رشته شناسه‌های جدا شده با کاما را به آرایه uint تبدیل می‌کند.
func parseUintList(raw string) ([]uint, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}
	parts := strings.Split(raw, ",")
	out := make([]uint, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		n, err := strconv.ParseUint(p, 10, 64)
		if err != nil {
			return nil, apperror.New("VALIDATION_ERROR", "شناسه سازمان نامعتبر است.", "invalid organization id", 400)
		}
		out = append(out, uint(n))
	}
	return out, nil
}

// combineDateTime تاریخ و ساعت را به time.Time ترکیب می‌کند.
func combineDateTime(dateStr, timeStr string, endOfDay bool) (time.Time, error) {
	if timeStr == "" {
		if endOfDay {
			timeStr = "23:59:59"
		} else {
			timeStr = "00:00:00"
		}
	} else if len(timeStr) == 5 {
		if endOfDay {
			timeStr += ":59"
		} else {
			timeStr += ":00"
		}
	}
	parsed, err := time.ParseInLocation("2006-01-02 15:04:05", dateStr+" "+timeStr, time.Local)
	if err != nil {
		return time.Time{}, apperror.New("VALIDATION_ERROR", "فرمت تاریخ یا ساعت نامعتبر است.", err.Error(), 400)
	}
	return parsed, nil
}

type orgMaps struct {
	names  map[uint]string
	isFree map[uint]bool
}

// buildOrgMaps نقشه نام و وضعیت آزاد سازمان‌ها را می‌سازد.
func (s *Service) buildOrgMaps() (*orgMaps, error) {
	orgs, err := s.orgRepo.FindAll()
	if err != nil {
		return nil, apperror.New("DB_ERROR", "خطا در خواندن سازمان‌ها.", err.Error(), 500)
	}
	m := &orgMaps{
		names:  make(map[uint]string, len(orgs)),
		isFree: make(map[uint]bool, len(orgs)),
	}
	for _, o := range orgs {
		m.names[o.ID] = o.Name
		m.isFree[o.ID] = o.IsFree
	}
	return m, nil
}

// resolveFreeOrganizationID سازمان پایه آزاد را برای حالت «فقط تکمیلی» برمی‌گرداند.
func (s *Service) resolveFreeOrganizationID() (uint, string, error) {
	freeOrg, err := s.orgRepo.FindFree()
	if err == gorm.ErrRecordNotFound {
		return 0, "", apperror.New("VALIDATION_ERROR", "سازمان آزاد تعریف نشده است.", "free organization not found", 400)
	}
	if err != nil {
		return 0, "", apperror.New("DB_ERROR", "خطا در خواندن سازمان آزاد.", err.Error(), 500)
	}
	if freeOrg.IsTakmili {
		return 0, "", apperror.New("VALIDATION_ERROR", "سازمان آزاد نمی‌تواند بیمه تکمیلی باشد.", "invalid free organization", 400)
	}
	return freeOrg.ID, freeOrg.Name, nil
}

// RevenueByOrganization گزارش درآمد به تفکیک سازمان را برمی‌گرداند.
func (s *Service) RevenueByOrganization(q RevenueByOrganizationQuery) (*RevenueByOrganizationResponse, error) {
	suppIDs, err := parseUintList(q.SupplementaryOrganizationIDs)
	if err != nil {
		return nil, err
	}

	var (
		baseIDs            []uint
		freeOrganizationName string
	)
	if q.OnlySupplementary {
		freeID, freeName, err := s.resolveFreeOrganizationID()
		if err != nil {
			return nil, err
		}
		baseIDs = []uint{freeID}
		freeOrganizationName = freeName
	} else {
		baseIDs, err = parseUintList(q.BaseOrganizationIDs)
		if err != nil {
			return nil, err
		}
		if len(baseIDs) == 0 {
			return nil, apperror.New("VALIDATION_ERROR", "حداقل یک سازمان پایه انتخاب کنید.", "base_organization_ids required", 400)
		}
	}

	from, err := combineDateTime(q.FromDate, q.FromTime, false)
	if err != nil {
		return nil, err
	}
	to, err := combineDateTime(q.ToDate, q.ToTime, true)
	if err != nil {
		return nil, err
	}
	if to.Before(from) {
		return nil, apperror.New("VALIDATION_ERROR", "تاریخ پایان نمی‌تواند قبل از تاریخ شروع باشد.", "invalid date range", 400)
	}

	hasSuppFilter := len(suppIDs) > 0
	detailSupplementary := q.DetailSupplementary && hasSuppFilter && !q.OnlySupplementary

	aggRows, err := s.repo.AggregateRevenueByOrganization(from, to, baseIDs, suppIDs, detailSupplementary, q.OnlySupplementary)
	if err != nil {
		return nil, apperror.New("DB_ERROR", "خطا در محاسبه گزارش.", err.Error(), 500)
	}

	orgInfo, err := s.buildOrgMaps()
	if err != nil {
		return nil, err
	}

	rows := make([]RevenueByOrganizationRow, 0, len(aggRows))
	var summary RevenueByOrganizationRow
	summary.RowLabel = "جمع کل"

	for _, a := range aggRows {
		baseName := orgInfo.names[a.InsuranceID]
		if baseName == "" {
			baseName = fmt.Sprintf("سازمان %d", a.InsuranceID)
		}

		row := RevenueByOrganizationRow{
			BaseOrganizationID:   a.InsuranceID,
			BaseOrganizationName: baseName,
			ServiceAmount:        a.ServiceAmount,
			ServiceTariff:        a.ServiceTariff,
			OrganizationShare:    a.OrganizationShare,
			SubsidyShare:         a.SubsidyShare,
			SupplementaryShare:   a.SupplementaryShare,
			ReceptionCount:       a.ReceptionCount,
		}
		row.PayableAmount = row.ServiceAmount - row.OrganizationShare - row.SupplementaryShare - row.SubsidyShare

		if q.OnlySupplementary {
			if a.AdditionalInsuranceID == nil {
				continue
			}
			suppName := orgInfo.names[*a.AdditionalInsuranceID]
			if suppName == "" {
				suppName = fmt.Sprintf("سازمان %d", *a.AdditionalInsuranceID)
			}
			row.SupplementaryOrganizationID = a.AdditionalInsuranceID
			row.SupplementaryOrganizationName = suppName
			row.RowLabel = suppName
		} else if hasSuppFilter && detailSupplementary {
			row.SupplementaryOrganizationID = a.AdditionalInsuranceID
			if a.AdditionalInsuranceID == nil {
				row.SupplementaryOrganizationName = "بدون بیمه تکمیلی"
				row.RowLabel = baseName + " - بدون بیمه تکمیلی"
			} else {
				suppName := orgInfo.names[*a.AdditionalInsuranceID]
				if suppName == "" {
					suppName = fmt.Sprintf("سازمان %d", *a.AdditionalInsuranceID)
				}
				row.SupplementaryOrganizationName = suppName
				if orgInfo.isFree[a.InsuranceID] {
					row.RowLabel = suppName
				} else {
					row.RowLabel = baseName + " - " + suppName
				}
			}
		} else {
			row.RowLabel = baseName
		}

		rows = append(rows, row)

		summary.ServiceAmount += row.ServiceAmount
		summary.ServiceTariff += row.ServiceTariff
		summary.OrganizationShare += row.OrganizationShare
		summary.SupplementaryShare += row.SupplementaryShare
		summary.SubsidyShare += row.SubsidyShare
		summary.PayableAmount += row.PayableAmount
		summary.ReceptionCount += row.ReceptionCount
	}

	sort.Slice(rows, func(i, j int) bool {
		if q.OnlySupplementary {
			return rows[i].SupplementaryOrganizationName < rows[j].SupplementaryOrganizationName
		}
		if rows[i].BaseOrganizationName != rows[j].BaseOrganizationName {
			return rows[i].BaseOrganizationName < rows[j].BaseOrganizationName
		}
		return rows[i].SupplementaryOrganizationName < rows[j].SupplementaryOrganizationName
	})

	fromTime := q.FromTime
	if fromTime == "" {
		fromTime = "00:00"
	}
	toTime := q.ToTime
	if toTime == "" {
		toTime = "23:59"
	}

	return &RevenueByOrganizationResponse{
		Rows:    rows,
		Summary: summary,
		Meta: RevenueByOrganizationMeta{
			FromDate:               q.FromDate,
			ToDate:                 q.ToDate,
			FromTime:               fromTime,
			ToTime:                 toTime,
			DetailSupplementary:    detailSupplementary,
			HasSupplementaryFilter: hasSuppFilter,
			OnlySupplementary:      q.OnlySupplementary,
			FreeOrganizationName:   freeOrganizationName,
		},
	}, nil
}

// parseDoctorPerformanceBase فیلترهای مشترک گزارش کارکرد پزشکان را اعتبارسنجی و parse می‌کند.
func parseDoctorPerformanceBase(q DoctorPerformanceBaseQuery) (from, to time.Time, doctorIDs []uint, err error) {
	doctorIDs, err = parseUintList(q.DoctorIDs)
	if err != nil {
		return time.Time{}, time.Time{}, nil, err
	}
	if len(doctorIDs) == 0 {
		return time.Time{}, time.Time{}, nil, apperror.New("VALIDATION_ERROR", "حداقل یک پزشک انتخاب کنید.", "doctor_ids required", 400)
	}

	from, err = combineDateTime(q.FromDate, q.FromTime, false)
	if err != nil {
		return time.Time{}, time.Time{}, nil, err
	}
	to, err = combineDateTime(q.ToDate, q.ToTime, true)
	if err != nil {
		return time.Time{}, time.Time{}, nil, err
	}
	if to.Before(from) {
		return time.Time{}, time.Time{}, nil, apperror.New("VALIDATION_ERROR", "تاریخ پایان نمی‌تواند قبل از تاریخ شروع باشد.", "invalid date range", 400)
	}
	return from, to, doctorIDs, nil
}

// buildDoctorPerformanceMeta متادیتای چاپ گزارش کارکرد پزشکان را می‌سازد.
func (s *Service) buildDoctorPerformanceMeta(
	reportType string,
	fromDate, toDate, fromTime, toTime string,
	doctorIDs []uint,
) (DoctorPerformanceMeta, error) {
	namesMap, err := s.repo.FindDoctorNames(doctorIDs)
	if err != nil {
		return DoctorPerformanceMeta{}, apperror.New("DB_ERROR", "خطا در خواندن نام پزشکان.", err.Error(), 500)
	}
	doctorNames := make([]string, 0, len(doctorIDs))
	for _, id := range doctorIDs {
		if name, ok := namesMap[id]; ok {
			doctorNames = append(doctorNames, name)
		} else {
			doctorNames = append(doctorNames, fmt.Sprintf("پزشک %d", id))
		}
	}
	if fromTime == "" {
		fromTime = "00:00"
	}
	if toTime == "" {
		toTime = "23:59"
	}
	return DoctorPerformanceMeta{
		ReportType:  reportType,
		FromDate:    fromDate,
		ToDate:      toDate,
		FromTime:    fromTime,
		ToTime:      toTime,
		DoctorNames: doctorNames,
	}, nil
}

// calcPayableAmount مبلغ قابل پرداخت را از اجزای سهم محاسبه می‌کند.
func calcPayableAmount(amount, orgShare, suppShare, subsidy int64) int64 {
	return amount - orgShare - suppShare - subsidy
}

// shouldSeparateDoctors بررسی می‌کند آیا گزارش باید به‌صورت جداگانه برای هر پزشک برگردد.
func shouldSeparateDoctors(separateDoctors bool, doctorIDs []uint) bool {
	return separateDoctors && len(doctorIDs) > 1
}

// mergeDoctorPerformanceSummaries جمع چند خلاصه مالی را محاسبه می‌کند.
func mergeDoctorPerformanceSummaries(summaries []DoctorPerformanceSummaryRow) DoctorPerformanceSummaryRow {
	out := DoctorPerformanceSummaryRow{RowLabel: "جمع کل"}
	for _, s := range summaries {
		out.ServiceAmount += s.ServiceAmount
		out.ServiceTariff += s.ServiceTariff
		out.OrganizationShare += s.OrganizationShare
		out.SupplementaryShare += s.SupplementaryShare
		out.SubsidyShare += s.SubsidyShare
		out.PayableAmount += s.PayableAmount
		out.ReceptionCount += s.ReceptionCount
	}
	return out
}

type doctorPerformanceOrgBuildOpts struct {
	onlySupplementary   bool
	hasSuppFilter       bool
	detailSupplementary bool
}

// buildDoctorPerformanceOrganizationRows سطرهای گزارش سازمان را از نتیجه تجمیع می‌سازد.
func buildDoctorPerformanceOrganizationRows(
	aggRows []revenueAggRow,
	orgInfo *orgMaps,
	opts doctorPerformanceOrgBuildOpts,
) ([]DoctorPerformanceByOrganizationRow, DoctorPerformanceSummaryRow) {
	rows := make([]DoctorPerformanceByOrganizationRow, 0, len(aggRows))
	var summary DoctorPerformanceSummaryRow
	summary.RowLabel = "جمع کل"

	for _, a := range aggRows {
		baseName := orgInfo.names[a.InsuranceID]
		if baseName == "" {
			baseName = fmt.Sprintf("سازمان %d", a.InsuranceID)
		}

		row := DoctorPerformanceByOrganizationRow{
			BaseOrganizationID:   a.InsuranceID,
			BaseOrganizationName: baseName,
			ServiceAmount:        a.ServiceAmount,
			ServiceTariff:        a.ServiceTariff,
			OrganizationShare:    a.OrganizationShare,
			SupplementaryShare:   a.SupplementaryShare,
			SubsidyShare:         a.SubsidyShare,
			ReceptionCount:       a.ReceptionCount,
		}
		row.PayableAmount = calcPayableAmount(row.ServiceAmount, row.OrganizationShare, row.SupplementaryShare, row.SubsidyShare)

		if opts.onlySupplementary {
			if a.AdditionalInsuranceID == nil {
				continue
			}
			suppName := orgInfo.names[*a.AdditionalInsuranceID]
			if suppName == "" {
				suppName = fmt.Sprintf("سازمان %d", *a.AdditionalInsuranceID)
			}
			row.SupplementaryOrganizationID = a.AdditionalInsuranceID
			row.SupplementaryOrganizationName = suppName
			row.RowLabel = suppName
		} else if opts.hasSuppFilter && opts.detailSupplementary {
			row.SupplementaryOrganizationID = a.AdditionalInsuranceID
			if a.AdditionalInsuranceID == nil {
				row.SupplementaryOrganizationName = "بدون بیمه تکمیلی"
				row.RowLabel = baseName + " - بدون بیمه تکمیلی"
			} else {
				suppName := orgInfo.names[*a.AdditionalInsuranceID]
				if suppName == "" {
					suppName = fmt.Sprintf("سازمان %d", *a.AdditionalInsuranceID)
				}
				row.SupplementaryOrganizationName = suppName
				if orgInfo.isFree[a.InsuranceID] {
					row.RowLabel = suppName
				} else {
					row.RowLabel = baseName + " - " + suppName
				}
			}
		} else {
			row.RowLabel = baseName
		}

		rows = append(rows, row)

		summary.ServiceAmount += row.ServiceAmount
		summary.ServiceTariff += row.ServiceTariff
		summary.OrganizationShare += row.OrganizationShare
		summary.SupplementaryShare += row.SupplementaryShare
		summary.SubsidyShare += row.SubsidyShare
		summary.PayableAmount += row.PayableAmount
		summary.ReceptionCount += row.ReceptionCount
	}

	sort.Slice(rows, func(i, j int) bool {
		if opts.onlySupplementary {
			return rows[i].SupplementaryOrganizationName < rows[j].SupplementaryOrganizationName
		}
		if rows[i].BaseOrganizationName != rows[j].BaseOrganizationName {
			return rows[i].BaseOrganizationName < rows[j].BaseOrganizationName
		}
		return rows[i].SupplementaryOrganizationName < rows[j].SupplementaryOrganizationName
	})

	return rows, summary
}

// buildDoctorPerformanceServiceRows سطرهای گزارش خدمت را از نتیجه تجمیع می‌سازد.
func buildDoctorPerformanceServiceRows(aggRows []doctorServiceAggRow) ([]DoctorPerformanceByServiceRow, DoctorPerformanceSummaryRow) {
	rows := make([]DoctorPerformanceByServiceRow, 0, len(aggRows))
	var summary DoctorPerformanceSummaryRow
	summary.RowLabel = "جمع کل"

	for _, a := range aggRows {
		row := DoctorPerformanceByServiceRow{
			ServiceID:          a.ServiceID,
			ServiceName:        a.ServiceName,
			ServiceAmount:      a.ServiceAmount,
			ServiceTariff:      a.ServiceTariff,
			OrganizationShare:  a.OrganizationShare,
			SupplementaryShare: a.SupplementaryShare,
			SubsidyShare:       a.SubsidyShare,
			ReceptionCount:     a.ReceptionCount,
		}
		row.PayableAmount = calcPayableAmount(row.ServiceAmount, row.OrganizationShare, row.SupplementaryShare, row.SubsidyShare)
		rows = append(rows, row)

		summary.ServiceAmount += row.ServiceAmount
		summary.ServiceTariff += row.ServiceTariff
		summary.OrganizationShare += row.OrganizationShare
		summary.SupplementaryShare += row.SupplementaryShare
		summary.SubsidyShare += row.SubsidyShare
		summary.PayableAmount += row.PayableAmount
		summary.ReceptionCount += row.ReceptionCount
	}

	return rows, summary
}

// buildDoctorPerformancePatientRows سطرهای لیست بیماران را از نتیجه تجمیع می‌سازد.
func buildDoctorPerformancePatientRows(aggRows []doctorPatientAggRow) ([]DoctorPerformancePatientRow, DoctorPerformanceSummaryRow) {
	rows := make([]DoctorPerformancePatientRow, 0, len(aggRows))
	var summary DoctorPerformanceSummaryRow
	summary.RowLabel = "جمع کل"

	for _, a := range aggRows {
		row := DoctorPerformancePatientRow{
			PatientID:          a.PatientID,
			PatientName:        strings.TrimSpace(a.PatientFirstName + " " + a.PatientLastName),
			NationalCode:       a.NationalCode,
			ServiceAmount:      a.ServiceAmount,
			ServiceTariff:      a.ServiceTariff,
			OrganizationShare:  a.OrganizationShare,
			SupplementaryShare: a.SupplementaryShare,
			SubsidyShare:       a.SubsidyShare,
			ServicesReceived:   a.ServicesReceived,
		}
		row.PayableAmount = calcPayableAmount(row.ServiceAmount, row.OrganizationShare, row.SupplementaryShare, row.SubsidyShare)
		rows = append(rows, row)

		summary.ServiceAmount += row.ServiceAmount
		summary.ServiceTariff += row.ServiceTariff
		summary.OrganizationShare += row.OrganizationShare
		summary.SupplementaryShare += row.SupplementaryShare
		summary.SubsidyShare += row.SubsidyShare
		summary.PayableAmount += row.PayableAmount
	}

	return rows, summary
}

// DoctorPerformanceByOrganization گزارش کارکرد پزشکان به تفکیک سازمان را برمی‌گرداند.
func (s *Service) DoctorPerformanceByOrganization(q DoctorPerformanceByOrganizationQuery) (*DoctorPerformanceByOrganizationResponse, error) {
	from, to, doctorIDs, err := parseDoctorPerformanceBase(q.DoctorPerformanceBaseQuery)
	if err != nil {
		return nil, err
	}

	suppIDs, err := parseUintList(q.SupplementaryOrganizationIDs)
	if err != nil {
		return nil, err
	}

	var (
		baseIDs              []uint
		freeOrganizationName string
	)
	if q.OnlySupplementary {
		freeID, freeName, err := s.resolveFreeOrganizationID()
		if err != nil {
			return nil, err
		}
		baseIDs = []uint{freeID}
		freeOrganizationName = freeName
	} else {
		baseIDs, err = parseUintList(q.BaseOrganizationIDs)
		if err != nil {
			return nil, err
		}
		if len(baseIDs) == 0 {
			return nil, apperror.New("VALIDATION_ERROR", "حداقل یک سازمان پایه انتخاب کنید.", "base_organization_ids required", 400)
		}
	}

	hasSuppFilter := len(suppIDs) > 0
	detailSupplementary := q.DetailSupplementary && hasSuppFilter && !q.OnlySupplementary
	buildOpts := doctorPerformanceOrgBuildOpts{
		onlySupplementary:   q.OnlySupplementary,
		hasSuppFilter:       hasSuppFilter,
		detailSupplementary: detailSupplementary,
	}

	orgInfo, err := s.buildOrgMaps()
	if err != nil {
		return nil, err
	}

	meta, err := s.buildDoctorPerformanceMeta("by_organization", q.FromDate, q.ToDate, q.FromTime, q.ToTime, doctorIDs)
	if err != nil {
		return nil, err
	}
	meta.DetailSupplementary = detailSupplementary
	meta.HasSupplementaryFilter = hasSuppFilter
	meta.OnlySupplementary = q.OnlySupplementary
	meta.FreeOrganizationName = freeOrganizationName

	separate := shouldSeparateDoctors(q.SeparateDoctors, doctorIDs)
	if separate {
		meta.SeparateDoctors = true
		namesMap, err := s.repo.FindDoctorNames(doctorIDs)
		if err != nil {
			return nil, apperror.New("DB_ERROR", "خطا در خواندن نام پزشکان.", err.Error(), 500)
		}

		sections := make([]DoctorPerformanceOrganizationSection, 0, len(doctorIDs))
		allRows := make([]DoctorPerformanceByOrganizationRow, 0)
		sectionSummaries := make([]DoctorPerformanceSummaryRow, 0, len(doctorIDs))

		for _, doctorID := range doctorIDs {
			aggRows, err := s.repo.AggregateDoctorPerformanceByOrganization(
				from, to, []uint{doctorID}, baseIDs, suppIDs, detailSupplementary, q.OnlySupplementary,
			)
			if err != nil {
				return nil, apperror.New("DB_ERROR", "خطا در محاسبه گزارش.", err.Error(), 500)
			}
			rows, summary := buildDoctorPerformanceOrganizationRows(aggRows, orgInfo, buildOpts)
			doctorName := namesMap[doctorID]
			if doctorName == "" {
				doctorName = fmt.Sprintf("پزشک %d", doctorID)
			}
			sections = append(sections, DoctorPerformanceOrganizationSection{
				DoctorID:   doctorID,
				DoctorName: doctorName,
				Rows:       rows,
				Summary:    summary,
			})
			allRows = append(allRows, rows...)
			sectionSummaries = append(sectionSummaries, summary)
		}

		return &DoctorPerformanceByOrganizationResponse{
			Rows:     allRows,
			Summary:  mergeDoctorPerformanceSummaries(sectionSummaries),
			Sections: sections,
			Meta:     meta,
		}, nil
	}

	aggRows, err := s.repo.AggregateDoctorPerformanceByOrganization(from, to, doctorIDs, baseIDs, suppIDs, detailSupplementary, q.OnlySupplementary)
	if err != nil {
		return nil, apperror.New("DB_ERROR", "خطا در محاسبه گزارش.", err.Error(), 500)
	}

	rows, summary := buildDoctorPerformanceOrganizationRows(aggRows, orgInfo, buildOpts)

	return &DoctorPerformanceByOrganizationResponse{
		Rows:    rows,
		Summary: summary,
		Meta:    meta,
	}, nil
}

// DoctorPerformanceByService گزارش کارکرد پزشکان به تفکیک خدمت را برمی‌گرداند.
func (s *Service) DoctorPerformanceByService(q DoctorPerformanceByServiceQuery) (*DoctorPerformanceByServiceResponse, error) {
	from, to, doctorIDs, err := parseDoctorPerformanceBase(q.DoctorPerformanceBaseQuery)
	if err != nil {
		return nil, err
	}

	serviceIDs, err := parseUintList(q.ServiceIDs)
	if err != nil {
		return nil, err
	}
	if len(serviceIDs) == 0 {
		return nil, apperror.New("VALIDATION_ERROR", "حداقل یک خدمت انتخاب کنید.", "service_ids required", 400)
	}

	meta, err := s.buildDoctorPerformanceMeta("by_service", q.FromDate, q.ToDate, q.FromTime, q.ToTime, doctorIDs)
	if err != nil {
		return nil, err
	}

	separate := shouldSeparateDoctors(q.SeparateDoctors, doctorIDs)
	if separate {
		meta.SeparateDoctors = true
		namesMap, err := s.repo.FindDoctorNames(doctorIDs)
		if err != nil {
			return nil, apperror.New("DB_ERROR", "خطا در خواندن نام پزشکان.", err.Error(), 500)
		}

		sections := make([]DoctorPerformanceServiceSection, 0, len(doctorIDs))
		allRows := make([]DoctorPerformanceByServiceRow, 0)
		sectionSummaries := make([]DoctorPerformanceSummaryRow, 0, len(doctorIDs))

		for _, doctorID := range doctorIDs {
			aggRows, err := s.repo.AggregateDoctorPerformanceByService(from, to, []uint{doctorID}, serviceIDs)
			if err != nil {
				return nil, apperror.New("DB_ERROR", "خطا در محاسبه گزارش.", err.Error(), 500)
			}
			rows, summary := buildDoctorPerformanceServiceRows(aggRows)
			doctorName := namesMap[doctorID]
			if doctorName == "" {
				doctorName = fmt.Sprintf("پزشک %d", doctorID)
			}
			sections = append(sections, DoctorPerformanceServiceSection{
				DoctorID:   doctorID,
				DoctorName: doctorName,
				Rows:       rows,
				Summary:    summary,
			})
			allRows = append(allRows, rows...)
			sectionSummaries = append(sectionSummaries, summary)
		}

		serviceNames := make([]string, 0, len(allRows))
		for _, row := range allRows {
			serviceNames = append(serviceNames, row.ServiceName)
		}
		meta.ServiceNames = serviceNames

		return &DoctorPerformanceByServiceResponse{
			Rows:     allRows,
			Summary:  mergeDoctorPerformanceSummaries(sectionSummaries),
			Sections: sections,
			Meta:     meta,
		}, nil
	}

	aggRows, err := s.repo.AggregateDoctorPerformanceByService(from, to, doctorIDs, serviceIDs)
	if err != nil {
		return nil, apperror.New("DB_ERROR", "خطا در محاسبه گزارش.", err.Error(), 500)
	}

	rows, summary := buildDoctorPerformanceServiceRows(aggRows)

	serviceNames := make([]string, 0, len(rows))
	for _, row := range rows {
		serviceNames = append(serviceNames, row.ServiceName)
	}
	meta.ServiceNames = serviceNames

	return &DoctorPerformanceByServiceResponse{
		Rows:    rows,
		Summary: summary,
		Meta:    meta,
	}, nil
}

// DoctorPerformancePatientList لیست بیماران پذیرش‌شده توسط پزشکان را برمی‌گرداند.
func (s *Service) DoctorPerformancePatientList(q DoctorPerformancePatientListQuery) (*DoctorPerformancePatientListResponse, error) {
	from, to, doctorIDs, err := parseDoctorPerformanceBase(q.DoctorPerformanceBaseQuery)
	if err != nil {
		return nil, err
	}

	meta, err := s.buildDoctorPerformanceMeta("patient_list", q.FromDate, q.ToDate, q.FromTime, q.ToTime, doctorIDs)
	if err != nil {
		return nil, err
	}

	separate := shouldSeparateDoctors(q.SeparateDoctors, doctorIDs)
	if separate {
		meta.SeparateDoctors = true
		namesMap, err := s.repo.FindDoctorNames(doctorIDs)
		if err != nil {
			return nil, apperror.New("DB_ERROR", "خطا در خواندن نام پزشکان.", err.Error(), 500)
		}

		sections := make([]DoctorPerformancePatientSection, 0, len(doctorIDs))
		allRows := make([]DoctorPerformancePatientRow, 0)
		sectionSummaries := make([]DoctorPerformanceSummaryRow, 0, len(doctorIDs))

		for _, doctorID := range doctorIDs {
			aggRows, err := s.repo.AggregateDoctorPerformancePatientList(from, to, []uint{doctorID})
			if err != nil {
				return nil, apperror.New("DB_ERROR", "خطا در محاسبه گزارش.", err.Error(), 500)
			}
			rows, summary := buildDoctorPerformancePatientRows(aggRows)
			doctorName := namesMap[doctorID]
			if doctorName == "" {
				doctorName = fmt.Sprintf("پزشک %d", doctorID)
			}
			sections = append(sections, DoctorPerformancePatientSection{
				DoctorID:   doctorID,
				DoctorName: doctorName,
				Rows:       rows,
				Summary:    summary,
			})
			allRows = append(allRows, rows...)
			sectionSummaries = append(sectionSummaries, summary)
		}

		return &DoctorPerformancePatientListResponse{
			Rows:     allRows,
			Summary:  mergeDoctorPerformanceSummaries(sectionSummaries),
			Sections: sections,
			Meta:     meta,
		}, nil
	}

	aggRows, err := s.repo.AggregateDoctorPerformancePatientList(from, to, doctorIDs)
	if err != nil {
		return nil, apperror.New("DB_ERROR", "خطا در محاسبه گزارش.", err.Error(), 500)
	}

	rows, summary := buildDoctorPerformancePatientRows(aggRows)

	return &DoctorPerformancePatientListResponse{
		Rows:    rows,
		Summary: summary,
		Meta:    meta,
	}, nil
}

// buildRevenueByDoctorsRows سطرهای گزارش درآمد به تفکیک پزشک را می‌سازد.
func buildRevenueByDoctorsRows(aggRows []revenueByDoctorAggRow) ([]RevenueByDoctorsRow, DoctorPerformanceSummaryRow) {
	rows := make([]RevenueByDoctorsRow, 0, len(aggRows))
	var summary DoctorPerformanceSummaryRow
	summary.RowLabel = "جمع کل"

	for _, a := range aggRows {
		medicalCode := ""
		if a.MedicalCode != nil {
			medicalCode = *a.MedicalCode
		}
		row := RevenueByDoctorsRow{
			DoctorID:           a.DoctorID,
			MedicalCode:        medicalCode,
			DoctorName:         strings.TrimSpace(a.DoctorFirstName + " " + a.DoctorLastName),
			ServiceAmount:      a.ServiceAmount,
			ServiceTariff:      a.ServiceTariff,
			OrganizationShare:  a.OrganizationShare,
			SupplementaryShare: a.SupplementaryShare,
			SubsidyShare:       a.SubsidyShare,
			ReceptionCount:     a.ReceptionCount,
		}
		row.PayableAmount = calcPayableAmount(row.ServiceAmount, row.OrganizationShare, row.SupplementaryShare, row.SubsidyShare)
		rows = append(rows, row)

		summary.ServiceAmount += row.ServiceAmount
		summary.ServiceTariff += row.ServiceTariff
		summary.OrganizationShare += row.OrganizationShare
		summary.SupplementaryShare += row.SupplementaryShare
		summary.SubsidyShare += row.SubsidyShare
		summary.PayableAmount += row.PayableAmount
		summary.ReceptionCount += row.ReceptionCount
	}

	return rows, summary
}

// RevenueByDoctors گزارش درآمد به تفکیک پزشکان را برمی‌گرداند.
func (s *Service) RevenueByDoctors(q RevenueByDoctorsQuery) (*RevenueByDoctorsResponse, error) {
	from, to, doctorIDs, err := parseDoctorPerformanceBase(q.DoctorPerformanceBaseQuery)
	if err != nil {
		return nil, err
	}

	aggRows, err := s.repo.AggregateRevenueByDoctors(from, to, doctorIDs)
	if err != nil {
		return nil, apperror.New("DB_ERROR", "خطا در محاسبه گزارش.", err.Error(), 500)
	}

	rows, summary := buildRevenueByDoctorsRows(aggRows)

	meta, err := s.buildDoctorPerformanceMeta("revenue_by_doctors", q.FromDate, q.ToDate, q.FromTime, q.ToTime, doctorIDs)
	if err != nil {
		return nil, err
	}

	return &RevenueByDoctorsResponse{
		Rows:    rows,
		Summary: summary,
		Meta:    meta,
	}, nil
}
