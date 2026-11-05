package wallet

import (
	"context"
	"strconv"

	"github.com/tpdenta/afta-reception/internal/platform/apperror"
	"github.com/tpdenta/afta-reception/internal/platform/security/audit"
	"gorm.io/gorm"
)

// bankFields فیلدهای قابل مقایسه حساب بانکی است.
var bankFields = []string{"bank_name", "sheba_number", "account_number", "card_number", "account_name"}

// creditFields فیلدهای قابل مقایسه حساب اعتباری است.
var creditFields = []string{"name", "description", "credit_limit", "is_active"}

// bankSnap فیلدهای قابل ممیزی حساب بانکی است.
type bankSnap struct {
	BankName      string `json:"bank_name"`
	ShebaNumber   string `json:"sheba_number"`
	AccountNumber string `json:"account_number"`
	CardNumber    string `json:"card_number"`
	AccountName   string `json:"account_name"`
}

// creditSnap فیلدهای قابل ممیزی حساب اعتباری است.
type creditSnap struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	CreditLimit int64  `json:"credit_limit"`
	IsActive    bool   `json:"is_active"`
}

// record رویداد را بیرون از تراکنش می‌نویسد.
// ctx زمینه است. e رویداد است. nil یعنی ثبت شد.
func (s *Service) record(ctx context.Context, e audit.Event) error {
	if s.audit == nil {
		return audit.Unavailable(apperror.ErrInternal)
	}
	if ctx == nil {
		ctx = context.Background()
	}
	return audit.Unavailable(s.audit.Record(ctx, e))
}

// recordTx رویداد را داخل تراکنش کیف پول می‌نویسد.
// tx تراکنش باز است. e رویداد است. خطا باید تراکنش را برگرداند.
func (s *Service) recordTx(tx *gorm.DB, e audit.Event) error {
	if s.audit == nil {
		return audit.Unavailable(apperror.ErrInternal)
	}
	return audit.Unavailable(s.audit.RecordTx(tx, e))
}

// walletMoneyEvent رویداد مبلغ کیف پول را با مبلغ، روش و شماره پیگیری می‌سازد.
// actor فاعل، tx تراکنش ذخیره‌شده، name نام خوانا، method روش و tracking شماره پیگیری است.
// شماره پیگیری خالی در جزئیات نمی‌آید. ماسک کردن آن لازم نیست.
func walletMoneyEvent(actor audit.Actor, tx *Transaction, name, method, tracking string) audit.Event {
	details := map[string]any{"amount": tx.Amount, "method": method}
	if tracking != "" {
		details["tracking_number"] = tracking
	}
	id := ""
	if tx != nil {
		id = strconv.FormatUint(uint64(tx.ID), 10)
	}
	return audit.Event{
		Type:        audit.EventEntityCreated,
		Result:      audit.ResultSuccess,
		Actor:       actor,
		Target:      audit.Target{Type: audit.TargetWallet, ID: id, Name: name},
		Details:     details,
		Description: "تراکنش کیف پول «" + name + "»",
	}
}

// recordCreditPaid تغییر تیک پرداخت اعتبار را ثبت می‌کند.
// ctx زمینه است. actor ویرایشگر، before وضعیت قبلی و after تراکنش جدید است.
func (s *Service) recordCreditPaid(ctx context.Context, actor audit.Actor, before bool, after *Transaction) error {
	if before == after.CreditPaid {
		return nil
	}
	return s.record(ctx, audit.Event{
		Type:   audit.EventEntityUpdated,
		Result: audit.ResultSuccess,
		Actor:  actor,
		Target: audit.Target{
			Type: audit.TargetWallet,
			ID:   strconv.FormatUint(uint64(after.ID), 10),
			Name: "اعتبار",
		},
		Details: audit.Diff(
			struct {
				CreditPaid bool `json:"credit_paid"`
			}{before},
			struct {
				CreditPaid bool `json:"credit_paid"`
			}{after.CreditPaid},
			[]string{"credit_paid"},
		),
		Description: "تغییر وضعیت پرداخت اعتبار",
	})
}

// recordBankCreated ایجاد حساب بانکی را ثبت می‌کند.
// ctx زمینه است. actor سازنده و account حساب جدید است. نام هدف نام حساب است.
func (s *Service) recordBankCreated(ctx context.Context, actor audit.Actor, account *BankAccount) error {
	return s.record(ctx, audit.Event{
		Type: audit.EventEntityCreated, Result: audit.ResultSuccess, Actor: actor,
		Target:      bankTarget(account),
		Description: "ایجاد حساب بانکی «" + account.AccountName + "»",
	})
}

// recordBankUpdated اختلاف حساب بانکی را ثبت می‌کند.
// ctx زمینه است. before حساب قبلی و after حساب جدید است. بدون تغییر رویدادی نوشته نمی‌شود.
func (s *Service) recordBankUpdated(ctx context.Context, actor audit.Actor, before, after *BankAccount) error {
	details := audit.Diff(snapBank(before), snapBank(after), bankFields)
	if len(details) == 0 {
		return nil
	}
	return s.record(ctx, audit.Event{
		Type: audit.EventEntityUpdated, Result: audit.ResultSuccess, Actor: actor,
		Target: bankTarget(after), Details: details, Description: "به‌روزرسانی حساب بانکی «" + after.AccountName + "»",
	})
}

// recordBankDeleted حذف حساب بانکی را ثبت می‌کند.
// ctx زمینه است. actor حذف‌کننده و account حساب حذف‌شده است. خروجی خطای ممیزی است.
func (s *Service) recordBankDeleted(ctx context.Context, actor audit.Actor, account *BankAccount) error {
	return s.record(ctx, audit.Event{
		Type: audit.EventEntityDeleted, Result: audit.ResultSuccess, Actor: actor,
		Target: bankTarget(account), Description: "حذف حساب بانکی «" + account.AccountName + "»",
	})
}

// recordCreditCreated ایجاد حساب اعتباری را ثبت می‌کند.
// ctx زمینه است. actor سازنده و account حساب جدید است. نام هدف نام اعتبار است.
func (s *Service) recordCreditCreated(ctx context.Context, actor audit.Actor, account *CreditAccount) error {
	return s.record(ctx, audit.Event{
		Type: audit.EventEntityCreated, Result: audit.ResultSuccess, Actor: actor,
		Target: creditTarget(account), Description: "ایجاد حساب اعتباری «" + account.Name + "»",
	})
}

// recordCreditUpdated اختلاف حساب اعتباری را ثبت می‌کند.
// ctx زمینه است. before حساب قبلی و after حساب جدید است. بدون تغییر رویدادی نوشته نمی‌شود.
func (s *Service) recordCreditUpdated(ctx context.Context, actor audit.Actor, before, after *CreditAccount) error {
	details := audit.Diff(snapCredit(before), snapCredit(after), creditFields)
	if len(details) == 0 {
		return nil
	}
	return s.record(ctx, audit.Event{
		Type: audit.EventEntityUpdated, Result: audit.ResultSuccess, Actor: actor,
		Target: creditTarget(after), Details: details, Description: "به‌روزرسانی حساب اعتباری «" + after.Name + "»",
	})
}

// recordCreditDeleted حذف حساب اعتباری را ثبت می‌کند.
// ctx زمینه است. actor حذف‌کننده و account حساب حذف‌شده است. خروجی خطای ممیزی است.
func (s *Service) recordCreditDeleted(ctx context.Context, actor audit.Actor, account *CreditAccount) error {
	return s.record(ctx, audit.Event{
		Type: audit.EventEntityDeleted, Result: audit.ResultSuccess, Actor: actor,
		Target: creditTarget(account), Description: "حذف حساب اعتباری «" + account.Name + "»",
	})
}

// bankTarget هدف حساب بانکی را می‌سازد.
// account حساب است. نام هدف نام حساب است.
func bankTarget(account *BankAccount) audit.Target {
	if account == nil {
		return audit.Target{Type: audit.TargetBankAccount}
	}
	return audit.Target{Type: audit.TargetBankAccount, ID: strconv.FormatUint(uint64(account.ID), 10), Name: account.AccountName}
}

// creditTarget هدف حساب اعتباری را می‌سازد.
// account حساب است. نام هدف نام اعتبار است.
func creditTarget(account *CreditAccount) audit.Target {
	if account == nil {
		return audit.Target{Type: audit.TargetCreditAccount}
	}
	return audit.Target{Type: audit.TargetCreditAccount, ID: strconv.FormatUint(uint64(account.ID), 10), Name: account.Name}
}

// snapBank فیلدهای قابل مقایسه حساب بانکی را می‌سازد.
// account حساب است. خروجی ورودی audit.Diff است.
func snapBank(account *BankAccount) bankSnap {
	if account == nil {
		return bankSnap{}
	}
	return bankSnap{
		BankName: account.BankName, ShebaNumber: account.ShebaNumber, AccountNumber: account.AccountNumber,
		CardNumber: account.CardNumber, AccountName: account.AccountName,
	}
}

// snapCredit فیلدهای قابل مقایسه حساب اعتباری را می‌سازد.
// account حساب است. خروجی ورودی audit.Diff است.
func snapCredit(account *CreditAccount) creditSnap {
	if account == nil {
		return creditSnap{}
	}
	return creditSnap{Name: account.Name, Description: account.Description, CreditLimit: account.CreditLimit, IsActive: account.IsActive}
}
