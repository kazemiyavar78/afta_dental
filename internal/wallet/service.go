package wallet

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/tpdenta/afta-reception/internal/platform/apperror"
	"github.com/tpdenta/afta-reception/internal/platform/security/audit"
	"github.com/tpdenta/afta-reception/internal/platform/security/integrity"
	"gorm.io/gorm"
)

// Service لایه منطق کسب‌وکار کیف پول و صندوق.
type Service struct {
	repo   Repository
	audit  audit.Recorder
	signer *integrity.Signer
	db     *gorm.DB
}

// NewService سرویس کیف پول را با وابستگی‌ها می‌سازد.
func NewService(db *gorm.DB, auditMgr *audit.Manager, signer *integrity.Signer) *Service {
	return &Service{
		repo:   NewRepository(db),
		audit:  auditMgr,
		signer: signer,
		db:     db,
	}
}

// GetWalletBalance موجودی کیف پول پرونده را برمی‌گرداند.
func (s *Service) GetWalletBalance(fileID uint) (*WalletBalanceResponse, error) {
	w, err := s.repo.FindPatientWalletByFileID(fileID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return &WalletBalanceResponse{FileID: fileID, Balance: 0}, nil
	}
	if err != nil {
		return nil, apperror.New("DB_ERROR", "خطا در خواندن کیف پول.", err.Error(), 500)
	}
	return &WalletBalanceResponse{FileID: w.FileID, Balance: w.Balance}, nil
}

// GetCashFund موجودی صندوق درآمد را برمی‌گرداند.
func (s *Service) GetCashFund() (*CashFundResponse, error) {
	fund, err := s.repo.GetCashFund()
	if err != nil {
		return nil, apperror.New("DB_ERROR", "خطا در خواندن صندوق.", err.Error(), 500)
	}
	return &CashFundResponse{ID: fund.ID, TotalBalance: fund.TotalBalance}, nil
}

// ListTransactionsByFile تراکنش‌های یک پرونده را برمی‌گرداند.
func (s *Service) ListTransactionsByFile(fileID uint) ([]TransactionResponse, error) {
	list, err := s.repo.ListTransactionsByFileID(fileID)
	if err != nil {
		return nil, apperror.New("DB_ERROR", "خطا در خواندن تراکنش‌ها.", err.Error(), 500)
	}
	names := s.lookupPerformedByNames(list)
	out := make([]TransactionResponse, 0, len(list))
	for i := range list {
		// if !VerifyTransactionIntegrity(s.signer, &list[i]) {
		// 	return nil, apperror.ErrIntegrity
		// }
		out = append(out, toTransactionResponse(&list[i], names[list[i].PerformedBy]))
	}
	return out, nil
}

// AdjustCash اعتبار نقدی بیمار را افزایش یا کاهش می‌دهد.
func (s *Service) AdjustCash(req CashAdjustRequest, ctx context.Context, actor audit.Actor) (*TransactionResponse, error) {
	if req.Amount <= 0 {
		return nil, apperror.New("BAD_REQUEST", "مبلغ باید بزرگ‌تر از صفر باشد.", "invalid amount", 400)
	}
	method := MethodCash
	action := ActionCharge
	delta := req.Amount
	if !req.Increase {
		action = ActionPayment
		delta = -req.Amount
	}
	desc := strings.TrimSpace(req.Description)
	if desc == "" {
		if req.Increase {
			desc = "افزایش اعتبار نقدی"
		} else {
			desc = "کاهش اعتبار نقدی"
		}
	}

	var created *Transaction
	err := s.repo.WithTx(func(txRepo Repository) error {
		if err := txRepo.AdjustPatientWalletBalance(nil, req.FileID, delta); err != nil {
			return err
		}
		t := &Transaction{
			FileID:        req.FileID,
			Amount:        req.Amount,
			Category:      CategoryWallet,
			Action:        action,
			PaymentMethod: &method,
			Description:   desc,
			PerformedBy:   uint(maxInt(actor.ID(), 0)),
		}
		if err := txRepo.CreateTransaction(nil, t); err != nil {
			return err
		}
		t.IntegrityHash = SignTransactionIntegrityHash(s.signer, t)
		if err := txRepo.UpdateTransaction(nil, t); err != nil {
			return err
		}
		created = t
		return s.recordTx(txRepo.DB(), walletMoneyEvent(actor, t, "نقدی", string(method), ""))
	})
	if err != nil {
		return nil, mapWalletErr(err)
	}
	return ptrTransactionResponse(created), nil
}

// AdjustCardToCard اعتبار کارت‌به‌کارت را افزایش یا کاهش می‌دهد.
func (s *Service) AdjustCardToCard(req CardToCardAdjustRequest, ctx context.Context, actor audit.Actor) (*TransactionResponse, error) {
	if req.Amount <= 0 {
		return nil, apperror.New("BAD_REQUEST", "مبلغ باید بزرگ‌تر از صفر باشد.", "invalid amount", 400)
	}
	account, err := s.repo.FindBankAccountByID(req.BankAccountID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperror.New("NOT_FOUND", "حساب بانکی یافت نشد.", "bank account not found", 404)
	}
	if err != nil {
		return nil, apperror.New("DB_ERROR", "خطا در خواندن حساب بانکی.", err.Error(), 500)
	}
	if !VerifyBankAccountIntegrity(s.signer, account) {
		return nil, apperror.ErrIntegrity
	}

	paidAt := time.Now().UTC()
	if req.PaidAt != nil && strings.TrimSpace(*req.PaidAt) != "" {
		raw := strings.TrimSpace(*req.PaidAt)
		var parsed time.Time
		var pErr error
		for _, layout := range []string{
			time.RFC3339,
			"2006-01-02T15:04:05",
			"2006-01-02T15:04",
			"2006-01-02 15:04:05",
		} {
			parsed, pErr = time.Parse(layout, raw)
			if pErr == nil {
				break
			}
		}
		if pErr != nil {
			return nil, apperror.New("BAD_REQUEST", "زمان پرداخت نامعتبر است.", pErr.Error(), 400)
		}
		paidAt = parsed.UTC()
	}

	method := MethodCardToCard
	action := ActionCharge
	delta := req.Amount
	if !req.Increase {
		action = ActionPayment
		delta = -req.Amount
	}
	desc := strings.TrimSpace(req.Description)
	if desc == "" {
		if req.Increase {
			desc = "افزایش اعتبار کارت‌به‌کارت"
		} else {
			desc = "کاهش اعتبار کارت‌به‌کارت"
		}
	}
	card := strings.TrimSpace(req.CounterpartyCard)
	tracking := strings.TrimSpace(req.TrackingNumber)
	bankName := account.BankName
	bankID := account.ID

	var created *Transaction
	err = s.repo.WithTx(func(txRepo Repository) error {
		if err := txRepo.AdjustPatientWalletBalance(nil, req.FileID, delta); err != nil {
			return err
		}
		t := &Transaction{
			FileID:           req.FileID,
			Amount:           req.Amount,
			Category:         CategoryWallet,
			Action:           action,
			PaymentMethod:    &method,
			Description:      desc,
			BankAccountID:    &bankID,
			CounterpartyCard: &card,
			TrackingNumber:   &tracking,
			BankName:         &bankName,
			PaidAt:           &paidAt,
			PerformedBy:      uint(maxInt(actor.ID(), 0)),
		}
		if err := txRepo.CreateTransaction(nil, t); err != nil {
			return err
		}
		t.IntegrityHash = SignTransactionIntegrityHash(s.signer, t)
		if err := txRepo.UpdateTransaction(nil, t); err != nil {
			return err
		}
		created = t
		return s.recordTx(txRepo.DB(), walletMoneyEvent(actor, t, bankName, string(method), tracking))
	})
	if err != nil {
		return nil, mapWalletErr(err)
	}
	return ptrTransactionResponse(created), nil
}

// ChargeFromCardReader شارژ کیف پول از پاسخ آزمایشی کارتخوان را ثبت می‌کند (فقط افزایش).
func (s *Service) ChargeFromCardReader(req CardReaderChargeRequest, ctx context.Context, actor audit.Actor) (*TransactionResponse, error) {
	if !req.Approved {
		return nil, apperror.New("BAD_REQUEST", "تراکنش کارتخوان تأیید نشده است.", "pos not approved", 400)
	}
	if req.Amount <= 0 {
		return nil, apperror.New("BAD_REQUEST", "مبلغ باید بزرگ‌تر از صفر باشد.", "invalid amount", 400)
	}
	method := MethodCardReader
	desc := strings.TrimSpace(req.Description)
	if desc == "" {
		desc = "شارژ از کارتخوان"
	}
	card := strings.TrimSpace(req.CardNumber)
	tracking := strings.TrimSpace(req.TransactionNumber)
	paidAt := time.Now().UTC()

	var created *Transaction
	err := s.repo.WithTx(func(txRepo Repository) error {
		if err := txRepo.AdjustPatientWalletBalance(nil, req.FileID, req.Amount); err != nil {
			return err
		}
		t := &Transaction{
			FileID:           req.FileID,
			Amount:           req.Amount,
			Category:         CategoryWallet,
			Action:           ActionCharge,
			PaymentMethod:    &method,
			Description:      desc,
			CounterpartyCard: &card,
			TrackingNumber:   &tracking,
			PaidAt:           &paidAt,
			PerformedBy:      uint(maxInt(actor.ID(), 0)),
		}
		if err := txRepo.CreateTransaction(nil, t); err != nil {
			return err
		}
		t.IntegrityHash = SignTransactionIntegrityHash(s.signer, t)
		if err := txRepo.UpdateTransaction(nil, t); err != nil {
			return err
		}
		created = t
		return s.recordTx(txRepo.DB(), walletMoneyEvent(actor, t, "کارتخوان", string(method), tracking))
	})
	if err != nil {
		return nil, mapWalletErr(err)
	}
	return ptrTransactionResponse(created), nil
}

// ChargeFromCreditAccount مبلغ را از سقف حساب اعتباری کسر و کیف پول پرونده را شارژ می‌کند.
// ورودی: درخواست شارژ، شناسه کاربر عامل و آی‌پی. خروجی: تراکنش ثبت‌شده یا خطا.
func (s *Service) ChargeFromCreditAccount(req CreditChargeRequest, ctx context.Context, actor audit.Actor) (*TransactionResponse, error) {
	if req.Amount <= 0 {
		return nil, apperror.New("BAD_REQUEST", "مبلغ باید بزرگ‌تر از صفر باشد.", "invalid amount", 400)
	}

	account, err := s.repo.FindCreditAccountByID(req.CreditAccountID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperror.New("NOT_FOUND", "حساب اعتباری یافت نشد.", "credit account not found", 404)
	}
	if err != nil {
		return nil, apperror.New("DB_ERROR", "خطا در خواندن حساب اعتباری.", err.Error(), 500)
	}
	if !VerifyCreditAccountIntegrity(s.signer, account) {
		return nil, apperror.ErrIntegrity
	}
	if !account.IsActive {
		return nil, apperror.New("BAD_REQUEST", "حساب اعتباری غیرفعال است.", "credit account inactive", 400)
	}

	method := MethodCreditAccount
	desc := strings.TrimSpace(req.Description)
	if desc == "" {
		desc = fmt.Sprintf("شارژ از حساب اعتباری %s", account.Name)
	}
	accountName := account.Name
	accountID := account.ID
	paidAt := time.Now().UTC()

	var created *Transaction
	err = s.repo.WithTx(func(txRepo Repository) error {
		ok, cErr := txRepo.TryConsumeCredit(nil, accountID, req.Amount)
		if cErr != nil {
			return cErr
		}
		if !ok {
			return s.creditConsumeFailure(txRepo, accountID, req.Amount)
		}
		current, fErr := txRepo.FindCreditAccountByID(accountID)
		if fErr != nil {
			return fErr
		}
		current.IntegrityHash = SignCreditAccountIntegrityHash(s.signer, current)
		if uErr := txRepo.UpdateCreditAccount(current); uErr != nil {
			return uErr
		}
		if aErr := txRepo.AdjustPatientWalletBalance(nil, req.FileID, req.Amount); aErr != nil {
			return aErr
		}
		t := &Transaction{
			FileID:            req.FileID,
			Amount:            req.Amount,
			Category:          CategoryWallet,
			Action:            ActionCharge,
			PaymentMethod:     &method,
			Description:       desc,
			CreditAccountID:   &accountID,
			CreditAccountName: &accountName,
			PaidAt:            &paidAt,
			PerformedBy:       uint(maxInt(actor.ID(), 0)),
		}
		if tErr := txRepo.CreateTransaction(nil, t); tErr != nil {
			return tErr
		}
		t.IntegrityHash = SignTransactionIntegrityHash(s.signer, t)
		if tErr := txRepo.UpdateTransaction(nil, t); tErr != nil {
			return tErr
		}
		created = t
		return s.recordTx(txRepo.DB(), walletMoneyEvent(actor, t, accountName, string(method), ""))
	})
	if err != nil {
		return nil, mapWalletErr(err)
	}
	return ptrTransactionResponse(created), nil
}

// SetCreditTransactionPaid وضعیت پرداخت‌شدن یک تراکنش کسر از اعتبار را ثبت یا برمی‌دارد.
// ورودی: شناسه تراکنش، پرداخت‌شده بودن، شناسه کاربر عامل و آی‌پی. خروجی: تراکنش به‌روزشده یا خطا.
func (s *Service) SetCreditTransactionPaid(id uint, paid bool, ctx context.Context, actor audit.Actor) (*TransactionResponse, error) {
	t, err := s.repo.FindTransactionByID(id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperror.ErrNotFound
	}
	if err != nil {
		return nil, apperror.New("DB_ERROR", "خطا در خواندن تراکنش.", err.Error(), 500)
	}
	if !VerifyTransactionIntegrity(s.signer, t) {
		return nil, apperror.ErrIntegrity
	}
	if t.PaymentMethod == nil || *t.PaymentMethod != MethodCreditAccount || t.Action != ActionCharge {
		return nil, apperror.New("BAD_REQUEST", "فقط تراکنش کسر از اعتبار قابل تیک پرداخت است.", "not a credit charge", 400)
	}

	beforePaid := t.CreditPaid
	t.CreditPaid = paid
	if paid {
		if t.CreditPaidAt == nil {
			now := time.Now().UTC()
			t.CreditPaidAt = &now
		}
	} else {
		t.CreditPaidAt = nil
	}
	t.IntegrityHash = SignTransactionIntegrityHash(s.signer, t)
	if err := s.repo.UpdateTransactionCreditPaid(t); err != nil {
		return nil, apperror.New("DB_ERROR", "خطا در ثبت وضعیت پرداخت.", err.Error(), 500)
	}
	if err := s.recordCreditPaid(ctx, actor, beforePaid, t); err != nil {
		return nil, err
	}
	return ptrTransactionResponse(t), nil
}

// creditConsumeFailure دلیل ناموفق بودن کسر از سقف اعتبار را به خطای دامنه تبدیل می‌کند.
// ورودی: ریپازیتوری تراکنش، شناسه حساب و مبلغ. خروجی: خطای دامنه.
func (s *Service) creditConsumeFailure(txRepo Repository, id uint, amount int64) error {
	current, err := txRepo.FindCreditAccountByID(id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return apperror.New("NOT_FOUND", "حساب اعتباری یافت نشد.", "credit account not found", 404)
	}
	if err != nil {
		return err
	}
	if !current.IsActive {
		return ErrCreditAccountInactive
	}
	if current.UsedAmount+amount > current.CreditLimit {
		return ErrCreditLimitExceeded
	}
	return apperror.New("CONFLICT", "کسر از حساب اعتباری انجام نشد.", "credit consume failed", 409)
}

// RecordReceptionServiceAdded هنگام ذخیره پذیرش، مبلغ صندوق را با service_added ثبت می‌کند.
func (s *Service) RecordReceptionServiceAdded(
	fileID uint,
	receptionID uint,
	amount int64,
	services []ReceptionServiceLine,
	ctx context.Context,
	actor audit.Actor,
) error {
	return s.recordReceptionServiceAdded(nil, fileID, receptionID, amount, services, ctx, actor)
}

// RecordReceptionServiceAddedTx همان ثبت صندوق پذیرش را داخل تراکنش بیرونی اجرا می‌کند.
func (s *Service) RecordReceptionServiceAddedTx(
	tx *gorm.DB,
	fileID uint,
	receptionID uint,
	amount int64,
	services []ReceptionServiceLine,
	ctx context.Context,
	actor audit.Actor,
) error {
	if tx == nil {
		return s.RecordReceptionServiceAdded(fileID, receptionID, amount, services, ctx, actor)
	}
	return s.recordReceptionServiceAdded(tx, fileID, receptionID, amount, services, ctx, actor)
}

// recordReceptionServiceAdded فقط اختلاف صندوق نسبت به خالص قبلی را ثبت می‌کند.
// بدهی خدمات فقط با رکورد cash_fund است؛ رکورد کیف پول فقط وقتی اعتبار قبلی مصرف شود.
func (s *Service) recordReceptionServiceAdded(
	tx *gorm.DB,
	fileID uint,
	receptionID uint,
	amount int64,
	services []ReceptionServiceLine,
	ctx context.Context,
	actor audit.Actor,
) error {
	if amount < 0 {
		amount = 0
	}
	fn := func(txRepo Repository) error {
		prev, err := txRepo.SumReceptionCashFundNet(receptionID)
		if err != nil {
			return err
		}
		delta := amount - prev
		if delta == 0 {
			return nil
		}
		return s.applyReceptionCashDelta(txRepo, fileID, receptionID, delta, services, ctx, actor)
	}
	var err error
	if tx != nil {
		err = fn(&gormRepo{db: tx})
	} else {
		err = s.repo.WithTx(fn)
	}
	if err != nil {
		return mapWalletErr(err)
	}
	return nil
}

// applyReceptionCashDelta اختلاف صندوق پذیرش را اعمال می‌کند (مثبت=بدهی بیشتر، منفی=کاهش بدهی).
func (s *Service) applyReceptionCashDelta(
	txRepo Repository,
	fileID uint,
	receptionID uint,
	delta int64,
	services []ReceptionServiceLine,
	ctx context.Context,
	actor audit.Actor,
) error {
	creditBefore, err := s.patientCreditBalance(txRepo, fileID)
	if err != nil {
		return err
	}

	if delta > 0 {
		desc := buildServicesDescription(services, "افزایش")
		if err := txRepo.AdjustCashFundBalance(nil, delta); err != nil {
			return err
		}
		t := &Transaction{
			FileID:      fileID,
			ReceptionID: &receptionID,
			Amount:      delta,
			Category:    CategoryCashFund,
			Action:      ActionServiceAdded,
			Description: desc,
			PerformedBy: uint(maxInt(actor.ID(), 0)),
		}
		if err := s.persistSignedTx(txRepo, t); err != nil {
			return err
		}
		// بدهی خدمات: موجودی کم می‌شود؛ رکورد جداگانه کیف پول برای کل بدهی ساخته نمی‌شود
		if err := txRepo.AdjustPatientWalletBalanceAllowDebt(nil, fileID, -delta); err != nil {
			return err
		}
		// اگر اعتبار داشت، تا سقف اعتبار یک رکورد پرداخت از کیف پول ثبت شود (بدون تغییر دوباره موجودی)
		if creditBefore > 0 {
			pay := creditBefore
			if pay > delta {
				pay = delta
			}
			method := MethodWalletCredit
			p := &Transaction{
				FileID:        fileID,
				ReceptionID:   &receptionID,
				Amount:        pay,
				Category:      CategoryWallet,
				Action:        ActionPayment,
				PaymentMethod: &method,
				Description:   fmt.Sprintf("پرداخت از اعتبار کیف پول بابت پذیرش %d", receptionID),
				PerformedBy:   uint(maxInt(actor.ID(), 0)),
			}
			if err := s.persistSignedTx(txRepo, p); err != nil {
				return err
			}
		}
		return s.recordTx(txRepo.DB(), walletMoneyEvent(actor, t, "پذیرش", string(ActionServiceAdded), ""))
	}

	// delta < 0 — کاهش بدهی / خدمات
	abs := -delta
	desc := buildServicesDescription(services, "کاهش")
	if err := txRepo.AdjustCashFundBalance(nil, -abs); err != nil {
		return err
	}
	t := &Transaction{
		FileID:      fileID,
		ReceptionID: &receptionID,
		Amount:      abs,
		Category:    CategoryCashFund,
		Action:      ActionServiceRemoved,
		Description: desc,
		PerformedBy: uint(maxInt(actor.ID(), 0)),
	}
	if err := s.persistSignedTx(txRepo, t); err != nil {
		return err
	}
	if err := txRepo.AdjustPatientWalletBalanceAllowDebt(nil, fileID, abs); err != nil {
		return err
	}
	return s.recordTx(txRepo.DB(), walletMoneyEvent(actor, t, "پذیرش", string(ActionServiceRemoved), ""))
}

// patientCreditBalance اعتبار مثبت فعلی بیمار را برمی‌گرداند (بدهی به‌عنوان صفر اعتبار).
func (s *Service) patientCreditBalance(txRepo Repository, fileID uint) (int64, error) {
	w, err := txRepo.FindPatientWalletByFileID(fileID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	if w.Balance > 0 {
		return w.Balance, nil
	}
	return 0, nil
}

// persistSignedTx تراکنش را درج و با هش یکپارچگی به‌روز می‌کند.
func (s *Service) persistSignedTx(txRepo Repository, t *Transaction) error {
	if err := txRepo.CreateTransaction(nil, t); err != nil {
		return err
	}
	t.IntegrityHash = SignTransactionIntegrityHash(s.signer, t)
	return txRepo.UpdateTransaction(nil, t)
}

// RecordReceptionServiceRemoved هنگام حذف پذیرش، خالص صندوق پذیرش را برمی‌گرداند.
func (s *Service) RecordReceptionServiceRemoved(
	fileID uint,
	receptionID uint,
	amount int64,
	services []ReceptionServiceLine,
	ctx context.Context,
	actor audit.Actor,
) error {
	_ = amount // مبلغ ورودی نادیده گرفته می‌شود؛ خالص واقعی از دفتر کل خوانده می‌شود
	err := s.repo.WithTx(func(txRepo Repository) error {
		prev, err := txRepo.SumReceptionCashFundNet(receptionID)
		if err != nil {
			return err
		}
		if prev == 0 {
			return nil
		}
		// اختلاف منفی = برگشت کامل بدهی ثبت‌شده
		return s.applyReceptionCashDelta(txRepo, fileID, receptionID, -prev, services, ctx, actor)
	})
	if err != nil {
		return mapWalletErr(err)
	}
	return nil
}

// CreateBankAccount حساب بانکی جدید ایجاد می‌کند.
func (s *Service) CreateBankAccount(req CreateBankAccountRequest, ctx context.Context, actor audit.Actor) (*BankAccountResponse, error) {
	a := &BankAccount{
		BankName:      strings.TrimSpace(req.BankName),
		ShebaNumber:   strings.TrimSpace(req.ShebaNumber),
		AccountNumber: strings.TrimSpace(req.AccountNumber),
		CardNumber:    strings.TrimSpace(req.CardNumber),
		AccountName:   strings.TrimSpace(req.AccountName),
	}
	if err := s.repo.CreateBankAccount(a); err != nil {
		return nil, apperror.New("DB_ERROR", "خطا در ایجاد حساب بانکی.", err.Error(), 500)
	}
	a.IntegrityHash = SignBankAccountIntegrityHash(s.signer, a)
	if err := s.repo.UpdateBankAccount(a); err != nil {
		return nil, apperror.New("DB_ERROR", "خطا در ذخیره هش حساب بانکی.", err.Error(), 500)
	}
	if err := s.recordBankCreated(ctx, actor, a); err != nil {
		return nil, err
	}
	return &BankAccountResponse{
		ID: a.ID, BankName: a.BankName, ShebaNumber: a.ShebaNumber,
		AccountNumber: a.AccountNumber, CardNumber: a.CardNumber,
		AccountName: a.AccountName, HasTransactions: false,
	}, nil
}

// UpdateBankAccount حساب بانکی بدون تراکنش را ویرایش می‌کند.
func (s *Service) UpdateBankAccount(id uint, req UpdateBankAccountRequest, ctx context.Context, actor audit.Actor) (*BankAccountResponse, error) {
	a, err := s.repo.FindBankAccountByID(id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperror.ErrNotFound
	}
	if err != nil {
		return nil, apperror.New("DB_ERROR", "خطا در خواندن حساب بانکی.", err.Error(), 500)
	}
	if !VerifyBankAccountIntegrity(s.signer, a) {
		return nil, apperror.ErrIntegrity
	}
	count, err := s.repo.CountTransactionsByBankAccount(id)
	if err != nil {
		return nil, apperror.New("DB_ERROR", "خطا در بررسی تراکنش‌های حساب.", err.Error(), 500)
	}
	if count > 0 {
		return nil, apperror.New("CONFLICT", "حساب بانکی دارای تراکنش است و قابل ویرایش نیست.", "bank account has transactions", 409)
	}
	before := *a
	a.BankName = strings.TrimSpace(req.BankName)
	a.ShebaNumber = strings.TrimSpace(req.ShebaNumber)
	a.AccountNumber = strings.TrimSpace(req.AccountNumber)
	a.CardNumber = strings.TrimSpace(req.CardNumber)
	a.AccountName = strings.TrimSpace(req.AccountName)
	a.IntegrityHash = SignBankAccountIntegrityHash(s.signer, a)
	if err := s.repo.UpdateBankAccount(a); err != nil {
		return nil, apperror.New("DB_ERROR", "خطا در ویرایش حساب بانکی.", err.Error(), 500)
	}
	if err := s.recordBankUpdated(ctx, actor, &before, a); err != nil {
		return nil, err
	}
	return &BankAccountResponse{
		ID: a.ID, BankName: a.BankName, ShebaNumber: a.ShebaNumber,
		AccountNumber: a.AccountNumber, CardNumber: a.CardNumber,
		AccountName: a.AccountName, HasTransactions: false,
	}, nil
}

// DeleteBankAccount حساب بانکی بدون تراکنش را حذف می‌کند.
func (s *Service) DeleteBankAccount(id uint, ctx context.Context, actor audit.Actor) error {
	a, err := s.repo.FindBankAccountByID(id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return apperror.ErrNotFound
	}
	if err != nil {
		return apperror.New("DB_ERROR", "خطا در خواندن حساب بانکی.", err.Error(), 500)
	}
	if !VerifyBankAccountIntegrity(s.signer, a) {
		return apperror.ErrIntegrity
	}
	count, err := s.repo.CountTransactionsByBankAccount(id)
	if err != nil {
		return apperror.New("DB_ERROR", "خطا در بررسی تراکنش‌های حساب.", err.Error(), 500)
	}
	if count > 0 {
		return apperror.New("CONFLICT", "حساب بانکی دارای تراکنش است و قابل حذف نیست.", "bank account has transactions", 409)
	}
	if err := s.repo.DeleteBankAccount(id); err != nil {
		return apperror.New("DB_ERROR", "خطا در حذف حساب بانکی.", err.Error(), 500)
	}
	return s.recordBankDeleted(ctx, actor, a)
}

// ListBankAccounts لیست حساب‌های بانکی را برمی‌گرداند.
func (s *Service) ListBankAccounts() ([]BankAccountResponse, error) {
	list, err := s.repo.ListBankAccounts()
	if err != nil {
		return nil, apperror.New("DB_ERROR", "خطا در خواندن حساب‌های بانکی.", err.Error(), 500)
	}
	out := make([]BankAccountResponse, 0, len(list))
	for i := range list {
		if !VerifyBankAccountIntegrity(s.signer, &list[i]) {
			return nil, apperror.ErrIntegrity
		}
		count, cErr := s.repo.CountTransactionsByBankAccount(list[i].ID)
		if cErr != nil {
			return nil, apperror.New("DB_ERROR", "خطا در بررسی تراکنش‌های حساب.", cErr.Error(), 500)
		}
		out = append(out, BankAccountResponse{
			ID: list[i].ID, BankName: list[i].BankName, ShebaNumber: list[i].ShebaNumber,
			AccountNumber: list[i].AccountNumber, CardNumber: list[i].CardNumber,
			AccountName: list[i].AccountName, HasTransactions: count > 0,
		})
	}
	return out, nil
}

// GetBankAccount یک حساب بانکی را برمی‌گرداند.
func (s *Service) GetBankAccount(id uint) (*BankAccountResponse, error) {
	a, err := s.repo.FindBankAccountByID(id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperror.ErrNotFound
	}
	if err != nil {
		return nil, apperror.New("DB_ERROR", "خطا در خواندن حساب بانکی.", err.Error(), 500)
	}
	if !VerifyBankAccountIntegrity(s.signer, a) {
		return nil, apperror.ErrIntegrity
	}
	count, err := s.repo.CountTransactionsByBankAccount(id)
	if err != nil {
		return nil, apperror.New("DB_ERROR", "خطا در بررسی تراکنش‌های حساب.", err.Error(), 500)
	}
	return &BankAccountResponse{
		ID: a.ID, BankName: a.BankName, ShebaNumber: a.ShebaNumber,
		AccountNumber: a.AccountNumber, CardNumber: a.CardNumber,
		AccountName: a.AccountName, HasTransactions: count > 0,
	}, nil
}

// CreateCreditAccount حساب اعتباری جدید ایجاد می‌کند.
// ورودی: درخواست ایجاد، شناسه کاربر عامل و آی‌پی. خروجی: حساب ایجادشده یا خطا.
func (s *Service) CreateCreditAccount(req CreateCreditAccountRequest, ctx context.Context, actor audit.Actor) (*CreditAccountResponse, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, apperror.New("BAD_REQUEST", "نام اعتبار الزامی است.", "name required", 400)
	}
	if req.CreditLimit <= 0 {
		return nil, apperror.New("BAD_REQUEST", "سقف اعتبار باید بزرگ‌تر از صفر باشد.", "invalid credit limit", 400)
	}
	a := &CreditAccount{
		Name:        name,
		Description: strings.TrimSpace(req.Description),
		CreditLimit: req.CreditLimit,
		UsedAmount:  0,
		IsActive:    req.IsActive,
	}
	if err := s.repo.CreateCreditAccount(a); err != nil {
		return nil, apperror.New("DB_ERROR", "خطا در ایجاد حساب اعتباری.", err.Error(), 500)
	}
	a.IntegrityHash = SignCreditAccountIntegrityHash(s.signer, a)
	if err := s.repo.UpdateCreditAccount(a); err != nil {
		return nil, apperror.New("DB_ERROR", "خطا در ذخیره هش حساب اعتباری.", err.Error(), 500)
	}
	if err := s.recordCreditCreated(ctx, actor, a); err != nil {
		return nil, err
	}
	resp := toCreditAccountResponse(a)
	return &resp, nil
}

// UpdateCreditAccount حساب اعتباری را ویرایش می‌کند.
// ورودی: شناسه، درخواست ویرایش، شناسه کاربر عامل و آی‌پی. خروجی: حساب ویرایش‌شده یا خطا.
func (s *Service) UpdateCreditAccount(id uint, req UpdateCreditAccountRequest, ctx context.Context, actor audit.Actor) (*CreditAccountResponse, error) {
	a, err := s.repo.FindCreditAccountByID(id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperror.ErrNotFound
	}
	if err != nil {
		return nil, apperror.New("DB_ERROR", "خطا در خواندن حساب اعتباری.", err.Error(), 500)
	}
	if !VerifyCreditAccountIntegrity(s.signer, a) {
		return nil, apperror.ErrIntegrity
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, apperror.New("BAD_REQUEST", "نام اعتبار الزامی است.", "name required", 400)
	}
	if req.CreditLimit <= 0 {
		return nil, apperror.New("BAD_REQUEST", "سقف اعتبار باید بزرگ‌تر از صفر باشد.", "invalid credit limit", 400)
	}
	if req.CreditLimit < a.UsedAmount {
		return nil, apperror.New("BAD_REQUEST", "سقف اعتبار نمی‌تواند کمتر از مبلغ استفاده‌شده باشد.", "credit limit below used amount", 400)
	}
	beforeCredit := *a
	a.Name = name
	a.Description = strings.TrimSpace(req.Description)
	a.CreditLimit = req.CreditLimit
	a.IsActive = req.IsActive
	a.IntegrityHash = SignCreditAccountIntegrityHash(s.signer, a)
	if err := s.repo.UpdateCreditAccount(a); err != nil {
		return nil, apperror.New("DB_ERROR", "خطا در ویرایش حساب اعتباری.", err.Error(), 500)
	}
	if err := s.recordCreditUpdated(ctx, actor, &beforeCredit, a); err != nil {
		return nil, err
	}
	resp := toCreditAccountResponse(a)
	return &resp, nil
}

// DeleteCreditAccount حساب اعتباری بدون پرداخت را حذف می‌کند.
// ورودی: شناسه، شناسه کاربر عامل و آی‌پی. خروجی: خطا در صورت عدم امکان حذف.
func (s *Service) DeleteCreditAccount(id uint, ctx context.Context, actor audit.Actor) error {
	a, err := s.repo.FindCreditAccountByID(id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return apperror.ErrNotFound
	}
	if err != nil {
		return apperror.New("DB_ERROR", "خطا در خواندن حساب اعتباری.", err.Error(), 500)
	}
	if !VerifyCreditAccountIntegrity(s.signer, a) {
		return apperror.ErrIntegrity
	}
	if a.UsedAmount > 0 {
		return apperror.New("CONFLICT", "حساب اعتباری دارای پرداخت است و قابل حذف نیست.", "credit account has payments", 409)
	}
	if err := s.repo.DeleteCreditAccount(id); err != nil {
		return apperror.New("DB_ERROR", "خطا در حذف حساب اعتباری.", err.Error(), 500)
	}
	return s.recordCreditDeleted(ctx, actor, a)
}

// ListCreditAccounts لیست حساب‌های اعتباری را برمی‌گرداند.
// خروجی: لیست حساب‌ها یا خطا.
func (s *Service) ListCreditAccounts() ([]CreditAccountResponse, error) {
	return s.listCreditAccounts(false)
}

// ListActiveCreditAccounts فقط حساب‌های اعتباری فعال را برمی‌گرداند.
// خروجی: لیست حساب‌های فعال یا خطا.
func (s *Service) ListActiveCreditAccounts() ([]CreditAccountResponse, error) {
	return s.listCreditAccounts(true)
}

// listCreditAccounts حساب‌های اعتباری را می‌خواند و یکپارچگی هر رکورد را بررسی می‌کند.
// ورودی: activeOnly برای محدود کردن به حساب‌های فعال. خروجی: لیست پاسخ یا خطا.
func (s *Service) listCreditAccounts(activeOnly bool) ([]CreditAccountResponse, error) {
	var (
		list []CreditAccount
		err  error
	)
	if activeOnly {
		list, err = s.repo.ListActiveCreditAccounts()
	} else {
		list, err = s.repo.ListCreditAccounts()
	}
	if err != nil {
		return nil, apperror.New("DB_ERROR", "خطا در خواندن حساب‌های اعتباری.", err.Error(), 500)
	}
	out := make([]CreditAccountResponse, 0, len(list))
	for i := range list {
		if !VerifyCreditAccountIntegrity(s.signer, &list[i]) {
			return nil, apperror.ErrIntegrity
		}
		out = append(out, toCreditAccountResponse(&list[i]))
	}
	return out, nil
}

// GetCreditAccount یک حساب اعتباری را برمی‌گرداند.
// ورودی: شناسه. خروجی: حساب یا خطا.
func (s *Service) GetCreditAccount(id uint) (*CreditAccountResponse, error) {
	a, err := s.repo.FindCreditAccountByID(id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperror.ErrNotFound
	}
	if err != nil {
		return nil, apperror.New("DB_ERROR", "خطا در خواندن حساب اعتباری.", err.Error(), 500)
	}
	if !VerifyCreditAccountIntegrity(s.signer, a) {
		return nil, apperror.ErrIntegrity
	}
	resp := toCreditAccountResponse(a)
	return &resp, nil
}

// toCreditAccountResponse مدل حساب اعتباری را به پاسخ API تبدیل می‌کند.
// ورودی: حساب. خروجی: DTO پاسخ شامل مانده سقف.
func toCreditAccountResponse(a *CreditAccount) CreditAccountResponse {
	remaining := a.CreditLimit - a.UsedAmount
	if remaining < 0 {
		remaining = 0
	}
	return CreditAccountResponse{
		ID:              a.ID,
		Name:            a.Name,
		Description:     a.Description,
		CreditLimit:     a.CreditLimit,
		UsedAmount:      a.UsedAmount,
		Remaining:       remaining,
		IsActive:        a.IsActive,
		HasTransactions: a.UsedAmount > 0,
	}
}

// buildServicesDescription توضیح تراکنش از لیست خدمات می‌سازد.
func buildServicesDescription(services []ReceptionServiceLine, verb string) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("%s خدمات صندوق:\n", verb))
	var total int64
	for _, svc := range services {
		b.WriteString(fmt.Sprintf("- %s (%s): %d ریال\n", svc.ServiceName, svc.ServiceCode, svc.CashAmount))
		total += svc.CashAmount
	}
	b.WriteString(fmt.Sprintf("جمع: %d ریال", total))
	return b.String()
}

// mapWalletErr خطای دامنه را به AppError نگاشت می‌کند.
func mapWalletErr(err error) error {
	if errors.Is(err, ErrInsufficientBalance) {
		return apperror.New("INSUFFICIENT_BALANCE", "موجودی کیف پول کافی نیست؛ بدهکار شدن بیمار مجاز نیست.", err.Error(), 400)
	}
	if errors.Is(err, ErrCreditLimitExceeded) {
		return apperror.New("CREDIT_LIMIT", "مانده سقف حساب اعتباری برای این مبلغ کافی نیست.", err.Error(), 400)
	}
	if errors.Is(err, ErrCreditAccountInactive) {
		return apperror.New("BAD_REQUEST", "حساب اعتباری غیرفعال است.", err.Error(), 400)
	}
	if ae, ok := err.(*apperror.AppError); ok {
		return ae
	}
	return apperror.New("DB_ERROR", "خطا در ثبت تراکنش کیف پول.", err.Error(), 500)
}

// toTransactionResponse مدل تراکنش را به DTO پاسخ تبدیل می‌کند.
// performedByName نام نمایشی عامل است؛ اگر خالی باشد از شناسه یا «سیستم» استفاده می‌شود.
func toTransactionResponse(t *Transaction, performedByName string) TransactionResponse {
	var method *string
	if t.PaymentMethod != nil {
		m := string(*t.PaymentMethod)
		method = &m
	}
	var paidAt *string
	if t.PaidAt != nil {
		v := t.PaidAt.UTC().Format(time.RFC3339)
		paidAt = &v
	}
	var creditPaidAt *string
	if t.CreditPaidAt != nil {
		v := t.CreditPaidAt.UTC().Format(time.RFC3339)
		creditPaidAt = &v
	}
	name := performedByName
	if name == "" {
		if t.PerformedBy == 0 {
			name = "سیستم"
		} else {
			name = fmt.Sprintf("#%d", t.PerformedBy)
		}
	}
	return TransactionResponse{
		ID: t.ID, FileID: t.FileID, ReceptionID: t.ReceptionID,
		Amount: t.Amount, Category: string(t.Category), Action: string(t.Action),
		PaymentMethod: method, Description: t.Description,
		BankAccountID: t.BankAccountID, CounterpartyCard: t.CounterpartyCard,
		TrackingNumber: t.TrackingNumber, BankName: t.BankName,
		CreditAccountID: t.CreditAccountID, CreditAccountName: t.CreditAccountName,
		CreditPaid: t.CreditPaid, CreditPaidAt: creditPaidAt,
		PaidAt: paidAt, PerformedBy: t.PerformedBy, PerformedByName: name,
		CreatedAt: t.CreatedAt,
	}
}

// ptrTransactionResponse اشاره‌گر پاسخ تراکنش می‌سازد.
func ptrTransactionResponse(t *Transaction) *TransactionResponse {
	if t == nil {
		return nil
	}
	resp := toTransactionResponse(t, "")
	return &resp
}

// lookupPerformedByNames نام کاربران عامل تراکنش‌ها را از جدول Users می‌خواند.
func (s *Service) lookupPerformedByNames(list []Transaction) map[uint]string {
	ids := make([]uint, 0)
	seen := map[uint]struct{}{}
	for i := range list {
		id := list[i].PerformedBy
		if id == 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		ids = append(ids, id)
	}
	out := map[uint]string{0: "سیستم"}
	if len(ids) == 0 {
		return out
	}
	type row struct {
		ID     int
		Name   string
		Family string
	}
	var rows []row
	if err := s.db.Table("Users").Select("ID, Name, Family").Where("ID IN ?", ids).Find(&rows).Error; err != nil {
		return out
	}
	for _, r := range rows {
		name := strings.TrimSpace(r.Name + " " + r.Family)
		if name == "" {
			name = fmt.Sprintf("#%d", r.ID)
		}
		out[uint(r.ID)] = name
	}
	return out
}

// maxInt مقدار بزرگ‌تر بین a و b را برمی‌گرداند.
func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
