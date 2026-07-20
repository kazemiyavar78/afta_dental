// Package cardreader شبیه‌ساز آزمایشی کارتخوان (POS) بدون اتصال واقعی به دستگاه است.
package cardreader

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"
)

// ChargeRequest درخواست شبیه‌سازی پرداخت کارتخوان.
type ChargeRequest struct {
	Amount int64 `json:"amount" binding:"required,gt=0"`
}

// ChargeResponse خروجی آزمایشی کارتخوان.
type ChargeResponse struct {
	TransactionNumber string `json:"transaction_number"`
	CardNumber        string `json:"card_number"`
	Amount            int64  `json:"amount"`
	Approved          bool   `json:"approved"`
	PaidAt            string `json:"paid_at"`
}

// Service شبیه‌ساز کارتخوان.
type Service struct{}

// NewService سرویس شبیه‌ساز کارتخوان را می‌سازد.
func NewService() *Service {
	return &Service{}
}

// SimulateCharge پاسخ آزمایشی کارتخوان را بدون اتصال به دستگاه تولید می‌کند.
func (s *Service) SimulateCharge(req ChargeRequest) (*ChargeResponse, error) {
	txn, err := randomHex(8)
	if err != nil {
		return nil, err
	}
	cardSuffix, err := randomHex(4)
	if err != nil {
		return nil, err
	}
	return &ChargeResponse{
		TransactionNumber: fmt.Sprintf("POS-%d-%s", time.Now().Unix(), txn),
		CardNumber:        fmt.Sprintf("6037****%s", cardSuffix),
		Amount:            req.Amount,
		Approved:          true,
		PaidAt:            time.Now().UTC().Format(time.RFC3339),
	}, nil
}

// randomHex رشته هگز تصادفی با طول n بایت تولید می‌کند.
func randomHex(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
