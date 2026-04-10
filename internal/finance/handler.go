package finance

import (
	"github.com/gin-gonic/gin"
)

// FinanceHandler wraps all finance sub-handlers and provides route registration.
type FinanceHandler struct {
	accountHandler  *AccountHandler
	voucherHandler *VoucherHandler
}

// NewFinanceHandler creates a new FinanceHandler.
func NewFinanceHandler(accountHandler *AccountHandler, voucherHandler *VoucherHandler) *FinanceHandler {
	return &FinanceHandler{
		accountHandler:  accountHandler,
		voucherHandler:  voucherHandler,
	}
}

// RegisterRoutes registers all finance routes within the given router group.
func (h *FinanceHandler) RegisterRoutes(rg *gin.RouterGroup) {
	if h.accountHandler != nil {
		h.accountHandler.RegisterRoutes(rg)
	}
	if h.voucherHandler != nil {
		h.voucherHandler.RegisterRoutes(rg)
	}
}
