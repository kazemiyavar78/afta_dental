package wallet

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/tpdenta/afta-reception/internal/platform/middleware"
)

// Handler هندلر HTTP کیف پول و حساب بانکی.
type Handler struct{ service *Service }

// NewHandler هندلر کیف پول را می‌سازد.
func NewHandler(s *Service) *Handler { return &Handler{service: s} }

func actorID(c *gin.Context) int {
	uid, _ := c.Get(middleware.ContextKeyUserID)
	if id, ok := uid.(int); ok {
		return id
	}
	return 0
}

// GetWalletBalance موجودی کیف پول پرونده را برمی‌گرداند.
func (h *Handler) GetWalletBalance(c *gin.Context) {
	fileID, err := strconv.ParseUint(c.Param("fileId"), 10, 64)
	if err != nil {
		middleware.WriteError(c, err)
		return
	}
	resp, err := h.service.GetWalletBalance(uint(fileID))
	if err != nil {
		middleware.WriteError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// GetCashFund موجودی صندوق درآمد را برمی‌گرداند.
func (h *Handler) GetCashFund(c *gin.Context) {
	resp, err := h.service.GetCashFund()
	if err != nil {
		middleware.WriteError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// ListTransactions تراکنش‌های یک پرونده را برمی‌گرداند.
func (h *Handler) ListTransactions(c *gin.Context) {
	fileID, err := strconv.ParseUint(c.Param("fileId"), 10, 64)
	if err != nil {
		middleware.WriteError(c, err)
		return
	}
	list, err := h.service.ListTransactionsByFile(uint(fileID))
	if err != nil {
		middleware.WriteError(c, err)
		return
	}
	c.JSON(http.StatusOK, list)
}

// AdjustCash افزایش/کاهش اعتبار نقدی.
func (h *Handler) AdjustCash(c *gin.Context) {
	var req CashAdjustRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.WriteError(c, err)
		return
	}
	resp, err := h.service.AdjustCash(req, actorID(c), c.ClientIP())
	if err != nil {
		middleware.WriteError(c, err)
		return
	}
	c.JSON(http.StatusCreated, resp)
}

// AdjustCardToCard افزایش/کاهش کارت‌به‌کارت.
func (h *Handler) AdjustCardToCard(c *gin.Context) {
	var req CardToCardAdjustRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.WriteError(c, err)
		return
	}
	resp, err := h.service.AdjustCardToCard(req, actorID(c), c.ClientIP())
	if err != nil {
		middleware.WriteError(c, err)
		return
	}
	c.JSON(http.StatusCreated, resp)
}

// ChargeFromCardReader شارژ از کارتخوان آزمایشی.
func (h *Handler) ChargeFromCardReader(c *gin.Context) {
	var req CardReaderChargeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.WriteError(c, err)
		return
	}
	resp, err := h.service.ChargeFromCardReader(req, actorID(c), c.ClientIP())
	if err != nil {
		middleware.WriteError(c, err)
		return
	}
	c.JSON(http.StatusCreated, resp)
}

// ChargeFromCreditAccount شارژ پرونده از محل کسر حساب اعتباری.
func (h *Handler) ChargeFromCreditAccount(c *gin.Context) {
	var req CreditChargeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.WriteError(c, err)
		return
	}
	resp, err := h.service.ChargeFromCreditAccount(req, actorID(c), c.ClientIP())
	if err != nil {
		middleware.WriteError(c, err)
		return
	}
	c.JSON(http.StatusCreated, resp)
}

// SetCreditTransactionPaid تیک پرداخت‌شدن تراکنش کسر از اعتبار را ثبت می‌کند.
func (h *Handler) SetCreditTransactionPaid(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		middleware.WriteError(c, err)
		return
	}
	var req SetCreditPaidRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.WriteError(c, err)
		return
	}
	resp, err := h.service.SetCreditTransactionPaid(uint(id), req.CreditPaid, actorID(c), c.ClientIP())
	if err != nil {
		middleware.WriteError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// ListBankAccounts لیست حساب‌های بانکی.
func (h *Handler) ListBankAccounts(c *gin.Context) {
	list, err := h.service.ListBankAccounts()
	if err != nil {
		middleware.WriteError(c, err)
		return
	}
	c.JSON(http.StatusOK, list)
}

// GetBankAccount یک حساب بانکی.
func (h *Handler) GetBankAccount(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		middleware.WriteError(c, err)
		return
	}
	resp, err := h.service.GetBankAccount(uint(id))
	if err != nil {
		middleware.WriteError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// CreateBankAccount ایجاد حساب بانکی.
func (h *Handler) CreateBankAccount(c *gin.Context) {
	var req CreateBankAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.WriteError(c, err)
		return
	}
	resp, err := h.service.CreateBankAccount(req, actorID(c), c.ClientIP())
	if err != nil {
		middleware.WriteError(c, err)
		return
	}
	c.JSON(http.StatusCreated, resp)
}

// UpdateBankAccount ویرایش حساب بانکی.
func (h *Handler) UpdateBankAccount(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		middleware.WriteError(c, err)
		return
	}
	var req UpdateBankAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.WriteError(c, err)
		return
	}
	resp, err := h.service.UpdateBankAccount(uint(id), req, actorID(c), c.ClientIP())
	if err != nil {
		middleware.WriteError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// DeleteBankAccount حذف حساب بانکی.
func (h *Handler) DeleteBankAccount(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		middleware.WriteError(c, err)
		return
	}
	if err := h.service.DeleteBankAccount(uint(id), actorID(c), c.ClientIP()); err != nil {
		middleware.WriteError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// ListCreditAccounts لیست حساب‌های اعتباری.
func (h *Handler) ListCreditAccounts(c *gin.Context) {
	list, err := h.service.ListCreditAccounts()
	if err != nil {
		middleware.WriteError(c, err)
		return
	}
	c.JSON(http.StatusOK, list)
}

// ListActiveCreditAccounts لیست حساب‌های اعتباری فعال برای فرم پرداخت.
func (h *Handler) ListActiveCreditAccounts(c *gin.Context) {
	list, err := h.service.ListActiveCreditAccounts()
	if err != nil {
		middleware.WriteError(c, err)
		return
	}
	c.JSON(http.StatusOK, list)
}

// GetCreditAccount یک حساب اعتباری.
func (h *Handler) GetCreditAccount(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		middleware.WriteError(c, err)
		return
	}
	resp, err := h.service.GetCreditAccount(uint(id))
	if err != nil {
		middleware.WriteError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// CreateCreditAccount ایجاد حساب اعتباری.
func (h *Handler) CreateCreditAccount(c *gin.Context) {
	var req CreateCreditAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.WriteError(c, err)
		return
	}
	resp, err := h.service.CreateCreditAccount(req, actorID(c), c.ClientIP())
	if err != nil {
		middleware.WriteError(c, err)
		return
	}
	c.JSON(http.StatusCreated, resp)
}

// UpdateCreditAccount ویرایش حساب اعتباری.
func (h *Handler) UpdateCreditAccount(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		middleware.WriteError(c, err)
		return
	}
	var req UpdateCreditAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.WriteError(c, err)
		return
	}
	resp, err := h.service.UpdateCreditAccount(uint(id), req, actorID(c), c.ClientIP())
	if err != nil {
		middleware.WriteError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

// DeleteCreditAccount حذف حساب اعتباری.
func (h *Handler) DeleteCreditAccount(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		middleware.WriteError(c, err)
		return
	}
	if err := h.service.DeleteCreditAccount(uint(id), actorID(c), c.ClientIP()); err != nil {
		middleware.WriteError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
