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
// authMiddleware is applied as a middleware to all finance routes.
func (h *FinanceHandler) RegisterRoutes(rg *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	authGroup := rg.Group("", authMiddleware)

	if h.accountHandler != nil {
		h.accountHandler.RegisterRoutes(authGroup)
	}
	if h.voucherHandler != nil {
		h.voucherHandler.RegisterRoutes(authGroup)
	}
	if h.invoiceHandler != nil {
		h.invoiceHandler.RegisterRoutes(authGroup)
	}
	if h.expenseHandler != nil {
		h.expenseHandler.RegisterRoutes(authGroup)
	}
	if h.bookHandler != nil {
		h.bookHandler.RegisterRoutes(authGroup)
	}
	if h.reportHandler != nil {
		h.reportHandler.RegisterRoutes(authGroup)
	}
}
