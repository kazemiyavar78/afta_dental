package report

import (
	"fmt"
	"strings"
	"time"

	"github.com/tpdenta/afta-reception/internal/platform/apperror"
	"github.com/tpdenta/afta-reception/internal/user"
)

// userTypeSortOrder ترتیب نمایش انواع کاربر در گزارش (مطابق تعریف مدل).
var userTypeSortOrder = map[user.UserType]int{
	user.UserTypeUser:      0,
	user.UserTypeDoctor:    1,
	user.UserTypeAssistant: 2,
	user.UserTypeSpecialist: 3,
}

// parseUserPerformanceBase فیلترهای مشترک گزارش کارکرد کاربران را اعتبارسنجی و parse می‌کند.
func parseUserPerformanceBase(q UserPerformanceBaseQuery) (from, to time.Time, userIDs []uint, err error) {
	userIDs, err = parseUintList(q.UserIDs)
	if err != nil {
		return time.Time{}, time.Time{}, nil, apperror.New("VALIDATION_ERROR", "شناسه کاربر نامعتبر است.", "invalid user id", 400)
	}
	if len(userIDs) == 0 {
		return time.Time{}, time.Time{}, nil, apperror.New("VALIDATION_ERROR", "حداقل یک کاربر انتخاب کنید.", "user_ids required", 400)
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
	return from, to, userIDs, nil
}

// buildUserPerformanceMeta متادیتای چاپ گزارش کارکرد کاربران را می‌سازد.
func (s *Service) buildUserPerformanceMeta(
	reportType string,
	fromDate, toDate, fromTime, toTime string,
	userIDs []uint,
	separateByPatient bool,
) (UserPerformanceMeta, error) {
	infos, err := s.repo.FindUserInfos(userIDs)
	if err != nil {
		return UserPerformanceMeta{}, apperror.New("DB_ERROR", "خطا در خواندن نام کاربران.", err.Error(), 500)
	}
	userNames := make([]string, 0, len(userIDs))
	for _, id := range userIDs {
		if info, ok := infos[id]; ok {
			userNames = append(userNames, strings.TrimSpace(info.Name+" "+info.Family))
		} else {
			userNames = append(userNames, fmt.Sprintf("کاربر %d", id))
		}
	}
	if fromTime == "" {
		fromTime = "00:00"
	}
	if toTime == "" {
		toTime = "23:59"
	}
	return UserPerformanceMeta{
		ReportType:        reportType,
		FromDate:          fromDate,
		ToDate:            toDate,
		FromTime:          fromTime,
		ToTime:            toTime,
		UserNames:         userNames,
		SeparateByPatient: separateByPatient,
	}, nil
}

// sortUserIDsByType کاربران را بر اساس نوع کاربر و سپس نام خانوادگی مرتب می‌کند.
func sortUserIDsByType(userIDs []uint, infos map[uint]userInfoRow) []uint {
	sorted := make([]uint, len(userIDs))
	copy(sorted, userIDs)
	type sorter struct {
		id   uint
		rank int
		name string
	}
	items := make([]sorter, 0, len(sorted))
	for _, id := range sorted {
		info, ok := infos[id]
		rank := 99
		name := fmt.Sprintf("کاربر %d", id)
		if ok {
			if r, exists := userTypeSortOrder[user.UserType(info.UserType)]; exists {
				rank = r
			}
			name = strings.TrimSpace(info.Name + " " + info.Family)
		}
		items = append(items, sorter{id: id, rank: rank, name: name})
	}
	for i := 0; i < len(items); i++ {
		for j := i + 1; j < len(items); j++ {
			swap := false
			if items[j].rank < items[i].rank {
				swap = true
			} else if items[j].rank == items[i].rank && items[j].name < items[i].name {
				swap = true
			}
			if swap {
				items[i], items[j] = items[j], items[i]
			}
		}
	}
	out := make([]uint, len(items))
	for i, item := range items {
		out[i] = item.id
	}
	return out
}

// userDisplayName نام نمایشی کاربر را برمی‌گرداند.
func userDisplayName(id uint, infos map[uint]userInfoRow) string {
	if info, ok := infos[id]; ok {
		return strings.TrimSpace(info.Name + " " + info.Family)
	}
	return fmt.Sprintf("کاربر %d", id)
}

// buildUserReceptionRow سطر تجمیعی پذیرش یک کاربر را می‌سازد.
func buildUserReceptionRow(userID uint, userName string, agg *userReceptionAggRow) UserPerformanceReceptionRow {
	row := UserPerformanceReceptionRow{
		UserID:             userID,
		UserName:           userName,
		ServiceAmount:      agg.ServiceAmount,
		ServiceTariff:      agg.ServiceTariff,
		OrganizationShare:  agg.OrganizationShare,
		SupplementaryShare: agg.SupplementaryShare,
		SubsidyShare:       agg.SubsidyShare,
		ReceptionCount:     agg.ReceptionCount,
	}
	row.PayableAmount = calcPayableAmount(row.ServiceAmount, row.OrganizationShare, row.SupplementaryShare, row.SubsidyShare)
	return row
}

// buildUserReceptionDetailRows سطرهای ریز پذیرش را می‌سازد.
func buildUserReceptionDetailRows(aggRows []userReceptionDetailAggRow) []UserPerformanceReceptionDetailRow {
	rows := make([]UserPerformanceReceptionDetailRow, 0, len(aggRows))
	for _, a := range aggRows {
		row := UserPerformanceReceptionDetailRow{
			ReceptionID:        a.ReceptionID,
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
	}
	return rows
}

// summarizeUserReceptionDetailRows جمع مالی ریز پذیرش‌ها را محاسبه می‌کند.
func summarizeUserReceptionDetailRows(rows []UserPerformanceReceptionDetailRow) UserPerformanceReceptionSummary {
	var summary UserPerformanceReceptionSummary
	summary.RowLabel = "جمع"
	for _, row := range rows {
		summary.ServiceAmount += row.ServiceAmount
		summary.ServiceTariff += row.ServiceTariff
		summary.OrganizationShare += row.OrganizationShare
		summary.SupplementaryShare += row.SupplementaryShare
		summary.SubsidyShare += row.SubsidyShare
		summary.PayableAmount += row.PayableAmount
	}
	summary.ReceptionCount = int64(len(rows))
	return summary
}

// mergeUserReceptionSummaries جمع چند خلاصه پذیرش را محاسبه می‌کند.
func mergeUserReceptionSummaries(summaries []UserPerformanceReceptionSummary) UserPerformanceReceptionSummary {
	out := UserPerformanceReceptionSummary{RowLabel: "جمع کل"}
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

// buildUserTransactionRow سطر تجمیعی تراکنش یک کاربر را می‌سازد.
func buildUserTransactionRow(userID uint, userName string, agg *userTransactionAggRow) UserPerformanceTransactionRow {
	return UserPerformanceTransactionRow{
		UserID:          userID,
		UserName:        userName,
		TotalReceived:   agg.TotalReceived,
		TotalPaid:       agg.TotalPaid,
		ReceivedTxCount: agg.ReceivedTxCount,
		PaidTxCount:     agg.PaidTxCount,
		TotalTxCount:    agg.TotalTxCount,
	}
}

// buildUserTransactionDetailRows سطرهای ریز تراکنش را می‌سازد.
func buildUserTransactionDetailRows(aggRows []userTransactionDetailAggRow) []UserPerformanceTransactionDetailRow {
	rows := make([]UserPerformanceTransactionDetailRow, 0, len(aggRows))
	for _, a := range aggRows {
		rows = append(rows, UserPerformanceTransactionDetailRow{
			TransactionID: a.TransactionID,
			FileNumber:    a.FileNumber,
			PatientName:   strings.TrimSpace(a.PatientFirstName + " " + a.PatientLastName),
			NationalCode:  a.NationalCode,
			Received:      a.Received,
			Paid:          a.Paid,
		})
	}
	return rows
}

// summarizeUserTransactionDetailRows جمع ریز تراکنش‌ها را محاسبه می‌کند.
func summarizeUserTransactionDetailRows(rows []UserPerformanceTransactionDetailRow) UserPerformanceTransactionSummary {
	var summary UserPerformanceTransactionSummary
	summary.RowLabel = "جمع"
	for _, row := range rows {
		summary.TotalReceived += row.Received
		summary.TotalPaid += row.Paid
		if row.Received > 0 {
			summary.ReceivedTxCount++
		}
		if row.Paid > 0 {
			summary.PaidTxCount++
		}
	}
	summary.TotalTxCount = int64(len(rows))
	return summary
}

// mergeUserTransactionSummaries جمع چند خلاصه تراکنش را محاسبه می‌کند.
func mergeUserTransactionSummaries(summaries []UserPerformanceTransactionSummary) UserPerformanceTransactionSummary {
	out := UserPerformanceTransactionSummary{RowLabel: "جمع کل"}
	for _, s := range summaries {
		out.TotalReceived += s.TotalReceived
		out.TotalPaid += s.TotalPaid
		out.ReceivedTxCount += s.ReceivedTxCount
		out.PaidTxCount += s.PaidTxCount
		out.TotalTxCount += s.TotalTxCount
	}
	return out
}

// UserPerformanceByReception گزارش کارکرد کاربران بر اساس پذیرش را برمی‌گرداند.
func (s *Service) UserPerformanceByReception(q UserPerformanceByReceptionQuery) (*UserPerformanceByReceptionResponse, error) {
	from, to, userIDs, err := parseUserPerformanceBase(q.UserPerformanceBaseQuery)
	if err != nil {
		return nil, err
	}

	infos, err := s.repo.FindUserInfos(userIDs)
	if err != nil {
		return nil, apperror.New("DB_ERROR", "خطا در خواندن اطلاعات کاربران.", err.Error(), 500)
	}
	sortedIDs := sortUserIDsByType(userIDs, infos)

	meta, err := s.buildUserPerformanceMeta("by_reception", q.FromDate, q.ToDate, q.FromTime, q.ToTime, sortedIDs, q.SeparateByPatient)
	if err != nil {
		return nil, err
	}

	if q.SeparateByPatient {
		sections := make([]UserPerformanceReceptionSection, 0, len(sortedIDs))
		sectionSummaries := make([]UserPerformanceReceptionSummary, 0, len(sortedIDs))
		for _, userID := range sortedIDs {
			info, ok := infos[userID]
			if !ok {
				continue
			}
			detailRows, err := s.repo.ListUserPerformanceReceptions(from, to, userID, user.UserType(info.UserType))
			if err != nil {
				return nil, apperror.New("DB_ERROR", "خطا در محاسبه گزارش.", err.Error(), 500)
			}
			receptionRows := buildUserReceptionDetailRows(detailRows)
			summary := summarizeUserReceptionDetailRows(receptionRows)
			sections = append(sections, UserPerformanceReceptionSection{
				UserID:   userID,
				UserName: userDisplayName(userID, infos),
				Rows:     receptionRows,
				Summary:  summary,
			})
			sectionSummaries = append(sectionSummaries, summary)
		}
		return &UserPerformanceByReceptionResponse{
			Summary:  mergeUserReceptionSummaries(sectionSummaries),
			Sections: sections,
			Meta:     meta,
		}, nil
	}

	rows := make([]UserPerformanceReceptionRow, 0, len(sortedIDs))
	var summary UserPerformanceReceptionSummary
	summary.RowLabel = "جمع کل"
	for _, userID := range sortedIDs {
		info, ok := infos[userID]
		if !ok {
			continue
		}
		agg, err := s.repo.AggregateUserPerformanceByReception(from, to, userID, user.UserType(info.UserType))
		if err != nil {
			return nil, apperror.New("DB_ERROR", "خطا در محاسبه گزارش.", err.Error(), 500)
		}
		row := buildUserReceptionRow(userID, userDisplayName(userID, infos), agg)
		rows = append(rows, row)
		summary.ServiceAmount += row.ServiceAmount
		summary.ServiceTariff += row.ServiceTariff
		summary.OrganizationShare += row.OrganizationShare
		summary.SupplementaryShare += row.SupplementaryShare
		summary.SubsidyShare += row.SubsidyShare
		summary.PayableAmount += row.PayableAmount
		summary.ReceptionCount += row.ReceptionCount
	}

	return &UserPerformanceByReceptionResponse{
		Rows:    rows,
		Summary: summary,
		Meta:    meta,
	}, nil
}

// UserPerformanceByTransaction گزارش کارکرد کاربران بر اساس تراکنش مالی را برمی‌گرداند.
func (s *Service) UserPerformanceByTransaction(q UserPerformanceByTransactionQuery) (*UserPerformanceByTransactionResponse, error) {
	from, to, userIDs, err := parseUserPerformanceBase(q.UserPerformanceBaseQuery)
	if err != nil {
		return nil, err
	}

	infos, err := s.repo.FindUserInfos(userIDs)
	if err != nil {
		return nil, apperror.New("DB_ERROR", "خطا در خواندن اطلاعات کاربران.", err.Error(), 500)
	}
	sortedIDs := sortUserIDsByType(userIDs, infos)

	meta, err := s.buildUserPerformanceMeta("by_transaction", q.FromDate, q.ToDate, q.FromTime, q.ToTime, sortedIDs, q.SeparateByPatient)
	if err != nil {
		return nil, err
	}

	if q.SeparateByPatient {
		sections := make([]UserPerformanceTransactionSection, 0, len(sortedIDs))
		sectionSummaries := make([]UserPerformanceTransactionSummary, 0, len(sortedIDs))
		for _, userID := range sortedIDs {
			detailRows, err := s.repo.ListUserPerformanceTransactions(from, to, userID)
			if err != nil {
				return nil, apperror.New("DB_ERROR", "خطا در محاسبه گزارش.", err.Error(), 500)
			}
			txRows := buildUserTransactionDetailRows(detailRows)
			summary := summarizeUserTransactionDetailRows(txRows)
			sections = append(sections, UserPerformanceTransactionSection{
				UserID:   userID,
				UserName: userDisplayName(userID, infos),
				Rows:     txRows,
				Summary:  summary,
			})
			sectionSummaries = append(sectionSummaries, summary)
		}
		return &UserPerformanceByTransactionResponse{
			Summary:  mergeUserTransactionSummaries(sectionSummaries),
			Sections: sections,
			Meta:     meta,
		}, nil
	}

	rows := make([]UserPerformanceTransactionRow, 0, len(sortedIDs))
	var summary UserPerformanceTransactionSummary
	summary.RowLabel = "جمع کل"
	for _, userID := range sortedIDs {
		agg, err := s.repo.AggregateUserPerformanceByTransaction(from, to, userID)
		if err != nil {
			return nil, apperror.New("DB_ERROR", "خطا در محاسبه گزارش.", err.Error(), 500)
		}
		row := buildUserTransactionRow(userID, userDisplayName(userID, infos), agg)
		rows = append(rows, row)
		summary.TotalReceived += row.TotalReceived
		summary.TotalPaid += row.TotalPaid
		summary.ReceivedTxCount += row.ReceivedTxCount
		summary.PaidTxCount += row.PaidTxCount
		summary.TotalTxCount += row.TotalTxCount
	}

	return &UserPerformanceByTransactionResponse{
		Rows:    rows,
		Summary: summary,
		Meta:    meta,
	}, nil
}
