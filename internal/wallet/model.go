package wallet

import (
	"strconv"
	"time"

	"github.com/tpdenta/afta-reception/internal/platform/security/integrity"
	"gorm.io/gorm"
)

// دسته‌بندی اصلی تراکنش
type TransactionCategory string

const (
	CategoryCashFund TransactionCategory = "cash_fund" // صندوق درآمد (خدمات)
	CategoryWallet   TransactionCategory = "wallet"    // کیف پول بیمار
)

// نوع عملیات (Action)
type TransactionAction string

const (
	// فقط برای صندوق (cash_fund)
	ActionServiceAdded   TransactionAction = "service_added"   // افزایش صندوق
	ActionServiceRemoved TransactionAction = "service_removed" // کاهش صندوق

	// فقط برای کیف پول (wallet)
	ActionCharge  TransactionAction = "charge"  // شارژ کیف پول
	ActionPayment TransactionAction = "payment" // پرداخت از کیف پول
	ActionRefund  TransactionAction = "refund"  // برگشت وجه به کیف پول
)

// روش پرداخت (فقط برای تراکنش‌های کیف پول معنی دارد)
type PaymentMethod string

const (
	MethodCash          PaymentMethod = "cash"
	MethodCardReader    PaymentMethod = "card_reader"
	MethodCardToCard    PaymentMethod = "card_to_card"
	MethodCheck         PaymentMethod = "check"
	MethodWalletCredit  PaymentMethod = "wallet_credit"
	MethodCreditAccount PaymentMethod = "credit_account" // کسر از حساب اعتباری
)

// CashFund موجودی لحظه‌ای صندوق درآمد کلینیک است (فقط یک رکورد).
type CashFund struct {
	gorm.Model
	TotalBalance int64 `json:"total_balance" gorm:"column:total_balance;not null;default:0"`
}

// TableName نام جدول صندوق درآمد.
func (CashFund) TableName() string { return "cash_funds" }

// PatientWallet کیف پول هر بیمار (یک رکورد به ازای هر پرونده).
type PatientWallet struct {
	gorm.Model
	FileID  uint  `json:"file_id" gorm:"column:file_id;uniqueIndex;not null"`
	Balance int64 `json:"balance" gorm:"column:balance;not null;default:0"`
}

// TableName نام جدول کیف پول بیمار.
func (PatientWallet) TableName() string { return "patient_wallets" }

// BankAccount حساب بانکی تعریف‌شده برای دریافت/پرداخت کارت‌به‌کارت.
type BankAccount struct {
	gorm.Model
	BankName      string `json:"bank_name" gorm:"column:bank_name;size:100;not null"`
	ShebaNumber   string `json:"sheba_number" gorm:"column:sheba_number;size:34;not null"`
	AccountNumber string `json:"account_number" gorm:"column:account_number;size:50;not null"`
	CardNumber    string `json:"card_number" gorm:"column:card_number;size:20;not null"`
	AccountName   string `json:"account_name" gorm:"column:account_name;size:200;not null"`
	IntegrityHash string `json:"-" gorm:"column:integrity_hash;size:128;not null"`
}

// TableName نام جدول حساب‌های بانکی.
func (BankAccount) TableName() string { return "bank_accounts" }

// Transaction یک رکورد در دفتر کل سیستم است.
// PerformedBy شناسه کاربری است که دریافت/پرداخت را انجام داده؛ مقدار 0 یعنی خود سیستم.
type Transaction struct {
	gorm.Model
	FileID            uint                `json:"file_id" gorm:"column:file_id;index;not null"`
	ReceptionID       *uint               `json:"reception_id" gorm:"column:reception_id;index"`
	CheckID           *string             `json:"check_id" gorm:"column:check_id;type:varchar(50)"`
	Amount            int64               `json:"amount" gorm:"column:amount;not null"`
	Category          TransactionCategory `json:"category" gorm:"column:category;type:varchar(20);not null"`
	Action            TransactionAction   `json:"action" gorm:"column:action;type:varchar(30);not null"`
	PaymentMethod     *PaymentMethod      `json:"payment_method" gorm:"column:payment_method;type:varchar(30)"`
	Description       string              `json:"description" gorm:"column:description;type:nvarchar(MAX)"`
	BankAccountID     *uint               `json:"bank_account_id" gorm:"column:bank_account_id;index"`
	CounterpartyCard  *string             `json:"counterparty_card" gorm:"column:counterparty_card;size:20"`
	TrackingNumber    *string             `json:"tracking_number" gorm:"column:tracking_number;size:100"`
	BankName          *string             `json:"bank_name" gorm:"column:bank_name;size:100"`
	PaidAt            *time.Time          `json:"paid_at" gorm:"column:paid_at"`
	PerformedBy       uint                `json:"performed_by" gorm:"column:performed_by;not null;default:0;index"`
	CreditAccountID   *uint               `json:"credit_account_id" gorm:"column:credit_account_id;index"`
	CreditAccountName *string             `json:"credit_account_name" gorm:"column:credit_account_name;size:200"`
	// آیا مبلغ کسر از اعتبار واقعاً پرداخت شده است (فقط برای روش credit_account)
	CreditPaid    bool       `json:"credit_paid" gorm:"column:credit_paid;not null;default:false"`
	CreditPaidAt  *time.Time `json:"credit_paid_at" gorm:"column:credit_paid_at"`
	IntegrityHash string     `json:"-" gorm:"column:integrity_hash;size:128;not null"`
}

// TableName نام جدول تراکنش‌ها.
func (Transaction) TableName() string { return "wallet_transactions" }

// BuildTransactionIntegrityFields فیلدهای HMAC تراکنش را برمی‌گرداند.
// ورودی: تراکنش. خروجی: فیلدهای امضا شامل حساب اعتباری و وضعیت پرداخت‌شدن.
func BuildTransactionIntegrityFields(t *Transaction) []string {
	paid := "0"
	if t.CreditPaid {
		paid = "1"
	}
	paidAt := ""
	if t.CreditPaidAt != nil {
		paidAt = t.CreditPaidAt.UTC().Format(time.RFC3339)
	}
	return append(buildTransactionIntegrityFieldsWithCredit(t), paid, paidAt)
}

// buildTransactionIntegrityFieldsWithCredit فیلدهای HMAC قبل از وضعیت پرداخت‌شدن را برمی‌گرداند.
// ورودی: تراکنش. خروجی: فیلدهای امضا شامل شناسه و نام حساب اعتباری.
func buildTransactionIntegrityFieldsWithCredit(t *Transaction) []string {
	creditAccountID := ""
	if t.CreditAccountID != nil {
		creditAccountID = strconv.FormatUint(uint64(*t.CreditAccountID), 10)
	}
	creditAccountName := ""
	if t.CreditAccountName != nil {
		creditAccountName = *t.CreditAccountName
	}
	return append(buildTransactionIntegrityFieldsLegacy(t), creditAccountID, creditAccountName)
}

// buildTransactionIntegrityFieldsLegacy فیلدهای HMAC تراکنش‌های قبل از حساب اعتباری را برمی‌گرداند.
// ورودی: تراکنش. خروجی: فیلدهای امضای قدیمی.
func buildTransactionIntegrityFieldsLegacy(t *Transaction) []string {
	receptionID := ""
	if t.ReceptionID != nil {
		receptionID = strconv.FormatUint(uint64(*t.ReceptionID), 10)
	}
	paymentMethod := ""
	if t.PaymentMethod != nil {
		paymentMethod = string(*t.PaymentMethod)
	}
	checkID := ""
	if t.CheckID != nil {
		checkID = *t.CheckID
	}
	bankAccountID := ""
	if t.BankAccountID != nil {
		bankAccountID = strconv.FormatUint(uint64(*t.BankAccountID), 10)
	}
	counterpartyCard := ""
	if t.CounterpartyCard != nil {
		counterpartyCard = *t.CounterpartyCard
	}
	trackingNumber := ""
	if t.TrackingNumber != nil {
		trackingNumber = *t.TrackingNumber
	}
	bankName := ""
	if t.BankName != nil {
		bankName = *t.BankName
	}
	paidAt := ""
	if t.PaidAt != nil {
		paidAt = t.PaidAt.UTC().Format(time.RFC3339)
	}
	return []string{
		strconv.FormatUint(uint64(t.ID), 10),
		strconv.FormatUint(uint64(t.FileID), 10),
		receptionID,
		strconv.FormatInt(t.Amount, 10),
		string(t.Category),
		string(t.Action),
		paymentMethod,
		checkID,
		bankAccountID,
		counterpartyCard,
		trackingNumber,
		bankName,
		paidAt,
		strconv.FormatUint(uint64(t.PerformedBy), 10),
		t.Description,
	}
}

// SignTransactionIntegrityHash هش یکپارچگی تراکنش را محاسبه می‌کند.
func SignTransactionIntegrityHash(signer *integrity.Signer, t *Transaction) string {
	return signer.Sign(BuildTransactionIntegrityFields(t)...)
}

// VerifyTransactionIntegrity یکپارچگی تراکنش را بررسی می‌کند.
// ورودی: امضاکننده و تراکنش. خروجی: true اگر هش جدید یا هش قدیمی (بدون حساب اعتباری) معتبر باشد.
func VerifyTransactionIntegrity(signer *integrity.Signer, t *Transaction) bool {
	if signer.Verify(t.IntegrityHash, BuildTransactionIntegrityFields(t)...) {
		return true
	}
	if signer.Verify(t.IntegrityHash, buildTransactionIntegrityFieldsWithCredit(t)...) {
		return true
	}
	return signer.Verify(t.IntegrityHash, buildTransactionIntegrityFieldsLegacy(t)...)
}

// CreditAccount حساب اعتباری برای شارژ پرونده مالی از محل کسر سقف اعتبار.
type CreditAccount struct {
	gorm.Model
	// نام اعتبار
	Name string `json:"name" gorm:"column:name;size:200;not null"`
	// توضیحات اعتبار
	Description string `json:"description" gorm:"column:description;size:500;not null"`
	// سقف اعتبار به ریال
	CreditLimit int64 `json:"credit_limit" gorm:"column:credit_limit;not null"`
	// مبلغ استفاده‌شده از سقف به ریال
	UsedAmount int64 `json:"used_amount" gorm:"column:used_amount;not null"`
	// فعال یا غیرفعال
	IsActive bool `json:"is_active" gorm:"column:is_active;not null"`
	// هش یکپارچگی
	IntegrityHash string `json:"-" gorm:"column:integrity_hash;size:128;not null"`
}

// TableName نام جدول حساب‌های اعتباری.
func (CreditAccount) TableName() string { return "credit_accounts" }

// BuildCreditAccountIntegrityFields فیلدهای HMAC حساب اعتباری را برمی‌گرداند.
// ورودی: حساب اعتباری. خروجی: فیلدهای امضا.
func BuildCreditAccountIntegrityFields(a *CreditAccount) []string {
	active := "0"
	if a.IsActive {
		active = "1"
	}
	return []string{
		strconv.FormatUint(uint64(a.ID), 10),
		a.Name,
		a.Description,
		strconv.FormatInt(a.CreditLimit, 10),
		strconv.FormatInt(a.UsedAmount, 10),
		active,
	}
}

// SignCreditAccountIntegrityHash هش یکپارچگی حساب اعتباری را محاسبه می‌کند.
// ورودی: امضاکننده و حساب. خروجی: هش HMAC.
func SignCreditAccountIntegrityHash(signer *integrity.Signer, a *CreditAccount) string {
	return signer.Sign(BuildCreditAccountIntegrityFields(a)...)
}

// VerifyCreditAccountIntegrity یکپارچگی حساب اعتباری را بررسی می‌کند.
// ورودی: امضاکننده و حساب. خروجی: true اگر هش معتبر باشد.
func VerifyCreditAccountIntegrity(signer *integrity.Signer, a *CreditAccount) bool {
	return signer.Verify(a.IntegrityHash, BuildCreditAccountIntegrityFields(a)...)
}

// BuildBankAccountIntegrityFields فیلدهای HMAC حساب بانکی را برمی‌گرداند.
func BuildBankAccountIntegrityFields(a *BankAccount) []string {
	return []string{
		strconv.FormatUint(uint64(a.ID), 10),
		a.BankName,
		a.ShebaNumber,
		a.AccountNumber,
		a.CardNumber,
		a.AccountName,
	}
}

// SignBankAccountIntegrityHash هش یکپارچگی حساب بانکی را محاسبه می‌کند.
func SignBankAccountIntegrityHash(signer *integrity.Signer, a *BankAccount) string {
	return signer.Sign(BuildBankAccountIntegrityFields(a)...)
}

// VerifyBankAccountIntegrity یکپارچگی حساب بانکی را بررسی می‌کند.
func VerifyBankAccountIntegrity(signer *integrity.Signer, a *BankAccount) bool {
	return signer.Verify(a.IntegrityHash, BuildBankAccountIntegrityFields(a)...)
}
