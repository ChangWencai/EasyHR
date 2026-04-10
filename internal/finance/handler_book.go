package finance

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// BookHandler handles book (账簿) routes.
type BookHandler struct {
	bookSvc *BookService
}

// NewBookHandler creates a new BookHandler.
func NewBookHandler(bookSvc *BookService) *BookHandler {
	return &BookHandler{bookSvc: bookSvc}
}

// RegisterRoutes registers the book routes.
func (h *BookHandler) RegisterRoutes(rg *gin.RouterGroup) {
	books := rg.Group("/books")
	{
		books.GET("/trial-balance", h.GetTrialBalance)
		books.GET("/account-balance", h.GetAccountBalance)
		books.GET("/ledger", h.GetLedger)
	}
}

// GetTrialBalance returns the trial balance for a period.
// GET /books/trial-balance?period_id=1
func (h *BookHandler) GetTrialBalance(c *gin.Context) {
	orgID := c.GetInt64("org_id")
	periodID, err := strconv.ParseInt(c.Query("period_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "period_id is required"})
		return
	}

	result, err := h.bookSvc.GetTrialBalance(c.Request.Context(), orgID, periodID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

// GetAccountBalance returns the account balance and entries for a specific account.
// GET /books/account-balance?period_id=1&account_id=1
func (h *BookHandler) GetAccountBalance(c *gin.Context) {
	orgID := c.GetInt64("org_id")
	periodID, err := strconv.ParseInt(c.Query("period_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "period_id is required"})
		return
	}
	accountID, err := strconv.ParseInt(c.Query("account_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "account_id is required"})
		return
	}

	result, err := h.bookSvc.GetAccountBalance(c.Request.Context(), orgID, periodID, accountID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

// GetLedger returns paginated ledger entries for an account in a period.
// GET /books/ledger?period_id=1&account_id=1&page=1&limit=50
func (h *BookHandler) GetLedger(c *gin.Context) {
	orgID := c.GetInt64("org_id")
	periodID, err := strconv.ParseInt(c.Query("period_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "period_id is required"})
		return
	}
	accountID, err := strconv.ParseInt(c.Query("account_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "account_id is required"})
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 200 {
		limit = 50
	}

	result, err := h.bookSvc.GetLedger(c.Request.Context(), orgID, periodID, accountID, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}
