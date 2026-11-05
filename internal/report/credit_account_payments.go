package report

import (
	"fmt"
	"strings"
	"time"

	"github.com/tpdenta/afta-reception/internal/platform/apperror"
	"gorm.io/gorm"
)

// CreditAccountPayments گزارش پرداخت‌های کسر از حساب‌های اعتباری را برمی‌گرداند.
// ورودی: فیلتر تاریخ و شناسه اختیاری حساب. خروجی: سطرها، جمع و متادیتای فیلتر.
func (s *Service) CreditAccountPayments(q CreditAccountPaymentsQuery) (*CreditAccountPaymentsResponse, error) {
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

	accountName := "همه حساب‌ها"
	if q.CreditAccountID > 0 {
		name, nErr := s.repo.FindCreditAccountName(q.CreditAccountID)
		if nErr == gorm.ErrRecordNotFound {
			return nil, apperror.New("NOT_FOUND", "حساب اعتباری یافت نشد.", "credit account not found", 404)
		}
		if nErr != nil {
			return nil, apperror.New("DB_ERROR", "خطا در خواندن حساب اعتباری.", nErr.Error(), 500)
		}
		accountName = name
	}

	rows, err := s.repo.ListCreditAccountPayments(from, to, q.CreditAccountID)
	if err != nil {
		return nil, apperror.New("DB_ERROR", "خطا در خواندن پرداخت‌های اعتباری.", err.Error(), 500)
	}

	out := make([]CreditAccountPaymentRow, 0, len(rows))
	var total int64
	for _, row := range rows {
		performer := strings.TrimSpace(row.PerformerName + " " + row.PerformerFamily)
		if row.PerformedBy == 0 {
			performer = "سیستم"
		} else if performer == "" {
			performer = fmt.Sprintf("#%d", row.PerformedBy)
		}
		creditPaidAt := ""
		if row.CreditPaidAt != nil {
			creditPaidAt = row.CreditPaidAt.UTC().Format(time.RFC3339)
		}
		out = append(out, CreditAccountPaymentRow{
			ID:                row.ID,
			PaidAt:            row.PaidAt.UTC().Format(time.RFC3339),
			CreditAccountID:   row.CreditAccountID,
			CreditAccountName: row.CreditAccountName,
			FileNumber:        row.FileNumber,
			PatientName:       strings.TrimSpace(row.PatientName),
			Amount:            row.Amount,
			Description:       row.Description,
			PerformedByName:   performer,
			CreditPaid:        row.CreditPaid,
			CreditPaidAt:      creditPaidAt,
		})
		total += row.Amount
	}

	fromTime := q.FromTime
	if fromTime == "" {
		fromTime = "00:00"
	}
	toTime := q.ToTime
	if toTime == "" {
		toTime = "23:59"
	}

	return &CreditAccountPaymentsResponse{
		Rows: out,
		Summary: CreditAccountPaymentsSummary{
			Amount:   total,
			RowCount: int64(len(out)),
			RowLabel: "جمع",
		},
		Meta: CreditAccountPaymentsMeta{
			FromDate:          q.FromDate,
			ToDate:            q.ToDate,
			FromTime:          fromTime,
			ToTime:            toTime,
			CreditAccountName: accountName,
		},
	}, nil
}
