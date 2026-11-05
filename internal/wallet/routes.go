package wallet

import (
	"github.com/gin-gonic/gin"
	"github.com/tpdenta/afta-reception/internal/platform/middleware"
)

// RegisterRoutes مسیرهای API کیف پول، حساب بانکی و حساب اعتباری را ثبت می‌کند.
func RegisterRoutes(r *gin.RouterGroup, h *Handler) {
	r.GET("/wallet/cash-fund", middleware.RequirePermission("wallet.read"), middleware.AuthorizationMiddleware(), h.GetCashFund)
	r.GET("/wallet/:fileId/balance", middleware.RequirePermission("wallet.read"), middleware.AuthorizationMiddleware(), h.GetWalletBalance)
	r.GET("/wallet/:fileId/transactions", middleware.RequirePermission("wallet.read"), middleware.AuthorizationMiddleware(), h.ListTransactions)
	r.POST("/wallet/cash", middleware.RequirePermission("wallet.cash"), middleware.AuthorizationMiddleware(), h.AdjustCash)
	r.POST("/wallet/card-to-card", middleware.RequirePermission("wallet.card_to_card"), middleware.AuthorizationMiddleware(), h.AdjustCardToCard)
	r.POST("/wallet/card-reader", middleware.RequirePermission("wallet.card_reader"), middleware.AuthorizationMiddleware(), h.ChargeFromCardReader)
	r.POST("/wallet/credit-charge", middleware.RequirePermission("credit_charge.create"), middleware.AuthorizationMiddleware(), h.ChargeFromCreditAccount)
	r.PATCH("/wallet/transactions/:id/credit-paid", middleware.RequirePermission("credit_charge.update"), middleware.AuthorizationMiddleware(), h.SetCreditTransactionPaid)

	r.GET("/credit-accounts", middleware.RequirePermission("credit_account.read", "reports.read"), middleware.AuthorizationMiddleware(), h.ListCreditAccounts)
	r.GET("/credit-accounts/active", middleware.RequirePermission("credit_charge.create", "credit_account.read"), middleware.AuthorizationMiddleware(), h.ListActiveCreditAccounts)
	r.POST("/credit-accounts", middleware.RequirePermission("credit_account.create"), middleware.AuthorizationMiddleware(), h.CreateCreditAccount)
	r.GET("/credit-accounts/:id", middleware.RequirePermission("credit_account.read"), middleware.AuthorizationMiddleware(), h.GetCreditAccount)
	r.PUT("/credit-accounts/:id", middleware.RequirePermission("credit_account.update"), middleware.AuthorizationMiddleware(), h.UpdateCreditAccount)
	r.DELETE("/credit-accounts/:id", middleware.RequirePermission("credit_account.delete"), middleware.AuthorizationMiddleware(), h.DeleteCreditAccount)

	r.GET("/bank-accounts", middleware.RequirePermission("bank_account.read"), middleware.AuthorizationMiddleware(), h.ListBankAccounts)
	r.POST("/bank-accounts", middleware.RequirePermission("bank_account.create"), middleware.AuthorizationMiddleware(), h.CreateBankAccount)
	r.GET("/bank-accounts/:id", middleware.RequirePermission("bank_account.read"), middleware.AuthorizationMiddleware(), h.GetBankAccount)
	r.PUT("/bank-accounts/:id", middleware.RequirePermission("bank_account.update"), middleware.AuthorizationMiddleware(), h.UpdateBankAccount)
	r.DELETE("/bank-accounts/:id", middleware.RequirePermission("bank_account.delete"), middleware.AuthorizationMiddleware(), h.DeleteBankAccount)
}
