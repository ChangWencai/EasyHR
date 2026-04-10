package finance

import (
	"github.com/gin-gonic/gin"
)

// FinanceHandler wraps all finance sub-handlers and provides route registration.
type FinanceHandler struct {
	accountHandler  *AccountHandler
	voucherHandler  *VoucherHandler
	invoiceHandler  *InvoiceHandler
	expenseHandler  *ExpenseHandler
	bookHandler     *BookHandler
	reportHandler   *ReportHandler
}

// NewFinanceHandler creates a new FinanceHandler.
func NewFinanceHandler(
	accountHandler *AccountHandler,
	voucherHandler *VoucherHandler,
	invoiceHandler *InvoiceHandler,
	expenseHandler *ExpenseHandler,
	bookHandler *BookHandler,
	reportHandler *ReportHandler,
) *FinanceHandler {
	return &FinanceHandler{
		accountHandler:  accountHandler,
		voucherHandler:  voucherHandler,
		invoiceHandler:  invoiceHandler,
		expenseHandler:  expenseHandler,
		bookHandler:    bookHandler,
		reportHandler:  reportHandler,
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
	if h.invoiceHandler != nil {
		h.invoiceHandler.RegisterRoutes(rg)
	}
	if h.expenseHandler != nil {
		h.expenseHandler.RegisterRoutes(rg)
	}
	if h.bookHandler != nil {
		h.bookHandler.RegisterRoutes(rg)
	}
	if h.reportHandler != nil {
		h.reportHandler.RegisterRoutes(rg)
	}
}
