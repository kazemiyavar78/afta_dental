package wallet

import (
	"errors"
	"time"
)

// ErrInsufficientBalance وقتی موجودی کیف پول برای کاهش کافی نباشد.
var ErrInsufficientBalance = errors.New("insufficient wallet balance")

// CashAdjustRequest درخواست افزایش/کاهش اعتبار نقدی.
type CashAdjustRequest struct {
	FileID      uint   `json:"file_id" binding:"required"`
	Amount      int64  `json:"amount" binding:"required,gt=0"`
	Increase    bool   `json:"increase"`
	Description string `json:"description"`
}

// CardToCardAdjustRequest درخواست افزایش/کاهش کارت‌به‌کارت.
type CardToCardAdjustRequest struct {
	FileID           uint    `json:"file_id" binding:"required"`
	Amount           int64   `json:"amount" binding:"required,gt=0"`
	Increase         bool    `json:"increase"`
	BankAccountID    uint    `json:"bank_account_id" binding:"required"`
	CounterpartyCard string  `json:"counterparty_card" binding:"required"`
	TrackingNumber   string  `json:"tracking_number" binding:"required"`
	PaidAt           *string `json:"paid_at"`
	Description      string  `json:"description"`
}

// CardReaderChargeRequest درخواست شارژ از کارتخوان آزمایشی.
type CardReaderChargeRequest struct {
	FileID            uint   `json:"file_id" binding:"required"`
	Amount            int64  `json:"amount" binding:"required,gt=0"`
	TransactionNumber string `json:"transaction_number" binding:"required"`
	CardNumber        string `json:"card_number" binding:"required"`
	Approved          bool   `json:"approved"`
	Description       string `json:"description"`
}

// CreateBankAccountRequest درخواست ایجاد حساب بانکی.
type CreateBankAccountRequest struct {
	BankName      string `json:"bank_name" binding:"required"`
	ShebaNumber   string `json:"sheba_number" binding:"required"`
	AccountNumber string `json:"account_number" binding:"required"`
	CardNumber    string `json:"card_number" binding:"required"`
	AccountName   string `json:"account_name" binding:"required"`
}

// UpdateBankAccountRequest درخواست ویرایش حساب بانکی.
type UpdateBankAccountRequest struct {
	BankName      string `json:"bank_name" binding:"required"`
	ShebaNumber   string `json:"sheba_number" binding:"required"`
	AccountNumber string `json:"account_number" binding:"required"`
	CardNumber    string `json:"card_number" binding:"required"`
	AccountName   string `json:"account_name" binding:"required"`
}

// WalletBalanceResponse موجودی کیف پول بیمار.
type WalletBalanceResponse struct {
	FileID  uint  `json:"file_id"`
	Balance int64 `json:"balance"`
}

// CashFundResponse موجودی صندوق درآمد.
type CashFundResponse struct {
	ID           uint  `json:"id"`
	TotalBalance int64 `json:"total_balance"`
}

// TransactionResponse پاسخ API تراکنش.
// PerformedBy شناسه کاربر عامل است؛ 0 یعنی سیستم. PerformedByName نام نمایشی عامل است.
type TransactionResponse struct {
	ID               uint      `json:"id"`
	FileID           uint      `json:"file_id"`
	ReceptionID      *uint     `json:"reception_id,omitempty"`
	Amount           int64     `json:"amount"`
	Category         string    `json:"category"`
	Action           string    `json:"action"`
	PaymentMethod    *string   `json:"payment_method,omitempty"`
	Description      string    `json:"description"`
	BankAccountID    *uint     `json:"bank_account_id,omitempty"`
	CounterpartyCard *string   `json:"counterparty_card,omitempty"`
	TrackingNumber   *string   `json:"tracking_number,omitempty"`
	BankName         *string   `json:"bank_name,omitempty"`
	PaidAt           *string   `json:"paid_at,omitempty"`
	PerformedBy      uint      `json:"performed_by"`
	PerformedByName  string    `json:"performed_by_name"`
	CreatedAt        time.Time `json:"created_at"`
}

// BankAccountResponse پاسخ API حساب بانکی.
type BankAccountResponse struct {
	ID            uint   `json:"id"`
	BankName      string `json:"bank_name"`
	ShebaNumber   string `json:"sheba_number"`
	AccountNumber string `json:"account_number"`
	CardNumber    string `json:"card_number"`
	AccountName   string `json:"account_name"`
	HasTransactions bool `json:"has_transactions"`
}

// ReceptionServiceLine خلاصه خدمت برای ثبت در توضیح تراکنش پذیرش.
type ReceptionServiceLine struct {
	ServiceCode string
	ServiceName string
	CashAmount  int64
}
