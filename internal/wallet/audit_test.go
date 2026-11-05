package wallet

import (
	"context"
	"testing"

	"github.com/tpdenta/afta-reception/internal/platform/security/audit"
	"github.com/tpdenta/afta-reception/internal/platform/security/integrity"
	"github.com/tpdenta/afta-reception/internal/platform/testutil"
	"gorm.io/gorm"
)

// bankAuditRepo فقط ایجاد و به‌روزرسانی حساب بانکی را نگه می‌دارد.
// بقیه متدها در این تست صدا زده نمی‌شوند و مقدار صفر برمی‌گردانند.
type bankAuditRepo struct{}

func (bankAuditRepo) WithTx(func(Repository) error) error { return nil }
func (bankAuditRepo) DB() *gorm.DB                        { return nil }
func (bankAuditRepo) GetOrCreateCashFund(*gorm.DB) (*CashFund, error) {
	return nil, nil
}
func (bankAuditRepo) AdjustCashFundBalance(*gorm.DB, int64) error { return nil }
func (bankAuditRepo) GetOrCreatePatientWallet(*gorm.DB, uint) (*PatientWallet, error) {
	return nil, nil
}
func (bankAuditRepo) AdjustPatientWalletBalance(*gorm.DB, uint, int64) error { return nil }
func (bankAuditRepo) AdjustPatientWalletBalanceAllowDebt(*gorm.DB, uint, int64) error {
	return nil
}
func (bankAuditRepo) CreateTransaction(*gorm.DB, *Transaction) error { return nil }
func (bankAuditRepo) UpdateTransaction(*gorm.DB, *Transaction) error { return nil }
func (bankAuditRepo) UpdateTransactionCreditPaid(*Transaction) error { return nil }
func (bankAuditRepo) FindTransactionByID(uint) (*Transaction, error) { return nil, nil }
func (bankAuditRepo) ListTransactionsByFileID(uint) ([]Transaction, error) {
	return nil, nil
}
func (bankAuditRepo) FindReceptionCashAdded(uint) (*Transaction, error) { return nil, nil }
func (bankAuditRepo) SumReceptionCashFundNet(uint) (int64, error)       { return 0, nil }
func (bankAuditRepo) CountTransactionsByBankAccount(uint) (int64, error) {
	return 0, nil
}
func (bankAuditRepo) CreateBankAccount(a *BankAccount) error {
	a.ID = 21
	return nil
}
func (bankAuditRepo) UpdateBankAccount(*BankAccount) error           { return nil }
func (bankAuditRepo) DeleteBankAccount(uint) error                   { return nil }
func (bankAuditRepo) FindBankAccountByID(uint) (*BankAccount, error) { return nil, nil }
func (bankAuditRepo) ListBankAccounts() ([]BankAccount, error)       { return nil, nil }
func (bankAuditRepo) CreateCreditAccount(*CreditAccount) error       { return nil }
func (bankAuditRepo) UpdateCreditAccount(*CreditAccount) error       { return nil }
func (bankAuditRepo) DeleteCreditAccount(uint) error                 { return nil }
func (bankAuditRepo) FindCreditAccountByID(uint) (*CreditAccount, error) {
	return nil, nil
}
func (bankAuditRepo) ListCreditAccounts() ([]CreditAccount, error)       { return nil, nil }
func (bankAuditRepo) ListActiveCreditAccounts() ([]CreditAccount, error) { return nil, nil }
func (bankAuditRepo) TryConsumeCredit(*gorm.DB, uint, int64) (bool, error) {
	return false, nil
}
func (bankAuditRepo) FindPatientWalletByFileID(uint) (*PatientWallet, error) {
	return nil, nil
}
func (bankAuditRepo) GetCashFund() (*CashFund, error) { return nil, nil }

// TestD13_BankAccountCreatedRecordsActorAndTarget ایجاد حساب بانکی را با فاعل و نام حساب ثبت می‌کند.
func TestD13_BankAccountCreatedRecordsActorAndTarget(t *testing.T) {
	rec := &testutil.FakeAuditRecorder{}
	svc := &Service{
		repo:   bankAuditRepo{},
		audit:  rec,
		signer: integrity.NewSigner("test-integrity-key-32bytes-long!!"),
	}
	id := 12
	_, err := svc.CreateBankAccount(CreateBankAccountRequest{
		BankName: "ملی", ShebaNumber: "IR1", AccountNumber: "10", CardNumber: "6037", AccountName: "حساب درمان",
	}, context.Background(), audit.Actor{UserID: &id, Username: "cashier", IP: "10.0.0.12"})
	if err != nil {
		t.Fatal(err)
	}
	if len(rec.Events) != 1 {
		t.Fatalf("یک رویداد ایجاد انتظار می‌رود، دریافت %d", len(rec.Events))
	}
	ev := rec.Events[0].Event
	if ev.Type != audit.EventEntityCreated || ev.Target.Type != audit.TargetBankAccount || ev.Target.ID != "21" || ev.Target.Name != "حساب درمان" {
		t.Fatalf("هدف حساب بانکی نادرست است: %+v", ev.Target)
	}
	if ev.Actor.UserID == nil || *ev.Actor.UserID != 12 || ev.Actor.Username != "cashier" {
		t.Fatalf("فاعل نادرست است: %+v", ev.Actor)
	}
}
