package wallet

import (
	"github.com/gin-gonic/gin"
	"github.com/tpdenta/afta-reception/internal/platform/middleware"
)

// RegisterRoutes مسیرهای API کیف پول و حساب بانکی را ثبت می‌کند.
func RegisterRoutes(r *gin.RouterGroup, h *Handler) {
	r.GET("/wallet/cash-fund", middleware.RequirePermission("wallet.read"), middleware.AuthorizationMiddleware(), h.GetCashFund)
	r.GET("/wallet/:fileId/balance", middleware.RequirePermission("wallet.read"), middleware.AuthorizationMiddleware(), h.GetWalletBalance)
	r.GET("/wallet/:fileId/transactions", middleware.RequirePermission("wallet.read"), middleware.AuthorizationMiddleware(), h.ListTransactions)
	r.POST("/wallet/cash", middleware.RequirePermission("wallet.cash"), middleware.AuthorizationMiddleware(), h.AdjustCash)
	r.POST("/wallet/card-to-card", middleware.RequirePermission("wallet.card_to_card"), middleware.AuthorizationMiddleware(), h.AdjustCardToCard)
	r.POST("/wallet/card-reader", middleware.RequirePermission("wallet.card_reader"), middleware.AuthorizationMiddleware(), h.ChargeFromCardReader)

	r.GET("/bank-accounts", middleware.RequirePermission("bank_account.read"), middleware.AuthorizationMiddleware(), h.ListBankAccounts)
	r.POST("/bank-accounts", middleware.RequirePermission("bank_account.create"), middleware.AuthorizationMiddleware(), h.CreateBankAccount)
	r.GET("/bank-accounts/:id", middleware.RequirePermission("bank_account.read"), middleware.AuthorizationMiddleware(), h.GetBankAccount)
	r.PUT("/bank-accounts/:id", middleware.RequirePermission("bank_account.update"), middleware.AuthorizationMiddleware(), h.UpdateBankAccount)
	r.DELETE("/bank-accounts/:id", middleware.RequirePermission("bank_account.delete"), middleware.AuthorizationMiddleware(), h.DeleteBankAccount)
}
