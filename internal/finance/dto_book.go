package finance

import "github.com/shopspring/decimal"

// AccountBalanceResponse represents a single account's balance in the trial balance.
type AccountBalanceResponse struct {
	AccountID     int64           `json:"account_id"`
	Code          string          `json:"code"`
	Name          string          `json:"name"`
	Category      AccountCategory `json:"category"`
	DebitSum      decimal.Decimal `json:"debit_sum"`
	CreditSum     decimal.Decimal `json:"credit_sum"`
	Balance       decimal.Decimal `json:"balance"`
	IsLeaf        bool            `json:"is_leaf"`
}

// TrialBalanceResponse represents the trial balance (科目余额表) for a period.
type TrialBalanceResponse struct {
	Items       []AccountBalanceResponse `json:"items"`
	TotalDebit  decimal.Decimal          `json:"total_debit"`
	TotalCredit decimal.Decimal          `json:"total_credit"`
	IsBalanced  bool                     `json:"is_balanced"`
	PeriodID    int64                    `json:"period_id"`
	Year        int                      `json:"year"`
	Month       int                      `json:"month"`
}

// LedgerEntryResponse represents a single line in an account ledger.
type LedgerEntryResponse struct {
	VoucherNo      string          `json:"voucher_no"`
	VoucherDate    string          `json:"voucher_date"`
	Description    string          `json:"description"`
	AccountName    string          `json:"account_name"`
	DC             string          `json:"dc"`
	Amount         decimal.Decimal `json:"amount"`
	BalanceAfter   decimal.Decimal `json:"balance_after"`
}

// LedgerResponse represents the detailed ledger (明细账) for an account in a period.
type LedgerResponse struct {
	AccountID          int64                   `json:"account_id"`
	AccountCode       string                  `json:"account_code"`
	AccountName       string                  `json:"account_name"`
	PeriodID          int64                   `json:"period_id"`
	Year              int                    `json:"year"`
	Month             int                    `json:"month"`
	Entries           []LedgerEntryResponse   `json:"entries"`
	PeriodDebitSum    decimal.Decimal         `json:"period_debit_sum"`
	PeriodCreditSum   decimal.Decimal         `json:"period_credit_sum"`
	EndingBalance     decimal.Decimal         `json:"ending_balance"`
	OpeningBalance    decimal.Decimal         `json:"opening_balance"`
}

// BookExportRequest represents a request to export book data.
type BookExportRequest struct {
	PeriodID int64  `form:"period_id" binding:"required"`
	AccountID *int64 `form:"account_id"`
	Format    string `form:"format"` // "excel" (default)
}

// AccountBalanceRequest represents a request to get a specific account's balance.
type AccountBalanceRequest struct {
	PeriodID  int64 `form:"period_id" binding:"required" json:"period_id"`
	AccountID int64 `form:"account_id" binding:"required" json:"account_id"`
}
