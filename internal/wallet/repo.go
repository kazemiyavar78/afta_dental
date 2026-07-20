package wallet

import (
	"errors"

	"gorm.io/gorm"
)

// Repository دسترسی داده‌ای کیف پول، صندوق، تراکنش و حساب بانکی.
type Repository interface {
	WithTx(fn func(txRepo Repository) error) error

	GetOrCreateCashFund(tx *gorm.DB) (*CashFund, error)
	AdjustCashFundBalance(tx *gorm.DB, delta int64) error

	GetOrCreatePatientWallet(tx *gorm.DB, fileID uint) (*PatientWallet, error)
	// AdjustPatientWalletBalance موجودی را تغییر می‌دهد؛ کاهش زیر صفر مجاز نیست (نقدی/کارت‌به‌کارت).
	AdjustPatientWalletBalance(tx *gorm.DB, fileID uint, delta int64) error
	// AdjustPatientWalletBalanceAllowDebt موجودی را تغییر می‌دهد و بدهکار شدن (منفی) را برای پذیرش مجاز می‌کند.
	AdjustPatientWalletBalanceAllowDebt(tx *gorm.DB, fileID uint, delta int64) error

	CreateTransaction(tx *gorm.DB, t *Transaction) error
	UpdateTransaction(tx *gorm.DB, t *Transaction) error
	FindTransactionByID(id uint) (*Transaction, error)
	ListTransactionsByFileID(fileID uint) ([]Transaction, error)
	FindReceptionCashAdded(receptionID uint) (*Transaction, error)
	// SumReceptionCashFundNet خالص مبلغ صندوق پذیرش (service_added − service_removed) را برمی‌گرداند.
	SumReceptionCashFundNet(receptionID uint) (int64, error)
	CountTransactionsByBankAccount(bankAccountID uint) (int64, error)

	CreateBankAccount(a *BankAccount) error
	UpdateBankAccount(a *BankAccount) error
	DeleteBankAccount(id uint) error
	FindBankAccountByID(id uint) (*BankAccount, error)
	ListBankAccounts() ([]BankAccount, error)

	FindPatientWalletByFileID(fileID uint) (*PatientWallet, error)
	GetCashFund() (*CashFund, error)
}

type gormRepo struct{ db *gorm.DB }

// NewRepository ریپازیتوری GORM کیف پول را می‌سازد.
func NewRepository(db *gorm.DB) Repository {
	return &gormRepo{db: db}
}

// WithTx عملیات را داخل تراکنش دیتابیس اجرا می‌کند.
func (r *gormRepo) WithTx(fn func(txRepo Repository) error) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		return fn(&gormRepo{db: tx})
	})
}

func (r *gormRepo) txOrDB(tx *gorm.DB) *gorm.DB {
	if tx != nil {
		return tx
	}
	return r.db
}

// GetOrCreateCashFund رکورد یکتای صندوق را می‌سازد یا برمی‌گرداند.
func (r *gormRepo) GetOrCreateCashFund(tx *gorm.DB) (*CashFund, error) {
	db := r.txOrDB(tx)
	var fund CashFund
	err := db.First(&fund).Error
	if err == nil {
		return &fund, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	fund = CashFund{TotalBalance: 0}
	if err := db.Create(&fund).Error; err != nil {
		return nil, err
	}
	return &fund, nil
}

// AdjustCashFundBalance موجودی صندوق را با UPDATE اتمیک تغییر می‌دهد (سازگار با SQL Server).
func (r *gormRepo) AdjustCashFundBalance(tx *gorm.DB, delta int64) error {
	db := r.txOrDB(tx)
	fund, err := r.GetOrCreateCashFund(db)
	if err != nil {
		return err
	}
	res := db.Model(&CashFund{}).
		Where("id = ?", fund.ID).
		UpdateColumn("total_balance", gorm.Expr("total_balance + ?", delta))
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return errors.New("cash fund update affected 0 rows")
	}
	return nil
}

// GetOrCreatePatientWallet کیف پول پرونده را می‌سازد یا برمی‌گرداند.
func (r *gormRepo) GetOrCreatePatientWallet(tx *gorm.DB, fileID uint) (*PatientWallet, error) {
	db := r.txOrDB(tx)
	var w PatientWallet
	err := db.Where("file_id = ?", fileID).First(&w).Error
	if err == nil {
		return &w, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	w = PatientWallet{FileID: fileID, Balance: 0}
	if err := db.Create(&w).Error; err != nil {
		return nil, err
	}
	return &w, nil
}

// AdjustPatientWalletBalance موجودی کیف پول را اتمیک تغییر می‌دهد؛ کاهش زیر صفر مجاز نیست.
func (r *gormRepo) AdjustPatientWalletBalance(tx *gorm.DB, fileID uint, delta int64) error {
	return r.adjustPatientWalletBalance(tx, fileID, delta, false)
}

// AdjustPatientWalletBalanceAllowDebt موجودی را اتمیک تغییر می‌دهد و اجازه بدهکار شدن بیمار را می‌دهد.
func (r *gormRepo) AdjustPatientWalletBalanceAllowDebt(tx *gorm.DB, fileID uint, delta int64) error {
	return r.adjustPatientWalletBalance(tx, fileID, delta, true)
}

// adjustPatientWalletBalance هسته به‌روزرسانی موجودی؛ allowDebt=true یعنی موجودی منفی مجاز است.
func (r *gormRepo) adjustPatientWalletBalance(tx *gorm.DB, fileID uint, delta int64, allowDebt bool) error {
	db := r.txOrDB(tx)
	w, err := r.GetOrCreatePatientWallet(db, fileID)
	if err != nil {
		return err
	}
	q := db.Model(&PatientWallet{}).Where("id = ?", w.ID)
	if delta < 0 && !allowDebt {
		q = q.Where("balance >= ?", -delta)
	}
	res := q.UpdateColumn("balance", gorm.Expr("balance + ?", delta))
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrInsufficientBalance
	}
	return nil
}

// CreateTransaction تراکنش را درج می‌کند.
func (r *gormRepo) CreateTransaction(tx *gorm.DB, t *Transaction) error {
	return r.txOrDB(tx).Create(t).Error
}

// UpdateTransaction تراکنش را به‌روز می‌کند.
func (r *gormRepo) UpdateTransaction(tx *gorm.DB, t *Transaction) error {
	return r.txOrDB(tx).Save(t).Error
}

// FindTransactionByID تراکنش را با شناسه برمی‌گرداند.
func (r *gormRepo) FindTransactionByID(id uint) (*Transaction, error) {
	var t Transaction
	err := r.db.First(&t, id).Error
	return &t, err
}

// ListTransactionsByFileID تراکنش‌های یک پرونده را برمی‌گرداند.
func (r *gormRepo) ListTransactionsByFileID(fileID uint) ([]Transaction, error) {
	var list []Transaction
	err := r.db.Where("file_id = ?", fileID).Order("id DESC").Find(&list).Error
	return list, err
}

// FindReceptionCashAdded تراکنش service_added مربوط به پذیرش را برمی‌گرداند.
// از Take به‌جای First استفاده می‌شود تا در MSSQL ستون id در ORDER BY تکراری نشود
// (First به‌صورت خودکار primary key را هم به ORDER BY اضافه می‌کند).
func (r *gormRepo) FindReceptionCashAdded(receptionID uint) (*Transaction, error) {
	var t Transaction
	err := r.db.Where(
		"reception_id = ? AND category = ? AND action = ?",
		receptionID, CategoryCashFund, ActionServiceAdded,
	).Order("id DESC").Take(&t).Error
	return &t, err
}

// SumReceptionCashFundNet خالص مبلغ صندوق پذیرش را برمی‌گرداند (جمع افزوده − جمع کاهشی).
func (r *gormRepo) SumReceptionCashFundNet(receptionID uint) (int64, error) {
	var added int64
	if err := r.db.Model(&Transaction{}).
		Where("reception_id = ? AND category = ? AND action = ? AND deleted_at IS NULL",
			receptionID, CategoryCashFund, ActionServiceAdded).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&added).Error; err != nil {
		return 0, err
	}
	var removed int64
	if err := r.db.Model(&Transaction{}).
		Where("reception_id = ? AND category = ? AND action = ? AND deleted_at IS NULL",
			receptionID, CategoryCashFund, ActionServiceRemoved).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&removed).Error; err != nil {
		return 0, err
	}
	return added - removed, nil
}

// CountTransactionsByBankAccount تعداد تراکنش‌های یک حساب بانکی را برمی‌گرداند.
func (r *gormRepo) CountTransactionsByBankAccount(bankAccountID uint) (int64, error) {
	var count int64
	err := r.db.Model(&Transaction{}).Where("bank_account_id = ?", bankAccountID).Count(&count).Error
	return count, err
}

// CreateBankAccount حساب بانکی جدید می‌سازد.
func (r *gormRepo) CreateBankAccount(a *BankAccount) error {
	return r.db.Create(a).Error
}

// UpdateBankAccount حساب بانکی را به‌روز می‌کند.
func (r *gormRepo) UpdateBankAccount(a *BankAccount) error {
	return r.db.Save(a).Error
}

// DeleteBankAccount حساب بانکی را حذف نرم می‌کند.
func (r *gormRepo) DeleteBankAccount(id uint) error {
	return r.db.Delete(&BankAccount{}, id).Error
}

// FindBankAccountByID حساب بانکی را با شناسه برمی‌گرداند.
func (r *gormRepo) FindBankAccountByID(id uint) (*BankAccount, error) {
	var a BankAccount
	err := r.db.First(&a, id).Error
	return &a, err
}

// ListBankAccounts لیست حساب‌های بانکی را برمی‌گرداند.
func (r *gormRepo) ListBankAccounts() ([]BankAccount, error) {
	var list []BankAccount
	err := r.db.Order("id DESC").Find(&list).Error
	return list, err
}

// FindPatientWalletByFileID کیف پول پرونده را برمی‌گرداند.
func (r *gormRepo) FindPatientWalletByFileID(fileID uint) (*PatientWallet, error) {
	var w PatientWallet
	err := r.db.Where("file_id = ?", fileID).First(&w).Error
	return &w, err
}

// GetCashFund رکورد صندوق را برمی‌گرداند.
func (r *gormRepo) GetCashFund() (*CashFund, error) {
	return r.GetOrCreateCashFund(nil)
}
