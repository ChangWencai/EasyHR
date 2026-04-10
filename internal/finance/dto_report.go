package finance

import (
	"time"

	"github.com/shopspring/decimal"
)

// BalanceSheetResponse represents a balance sheet report.
type BalanceSheetResponse struct {
	PeriodID        int64                `json:"period_id"`
	Year            int                  `json:"year"`
	Month           int                  `json:"month"`
	Assets          []BalanceSheetItem   `json:"assets"`
	Liabilities     []BalanceSheetItem   `json:"liabilities"`
	Equity          []BalanceSheetItem   `json:"equity"`
	AssetTotal      decimal.Decimal      `json:"asset_total"`
	LiabilityTotal  decimal.Decimal      `json:"liability_total"`
	EquityTotal     decimal.Decimal     `json:"equity_total"`
	IsBalanced      bool                 `json:"is_balanced"`
	GeneratedAt     time.Time           `json:"generated_at"`
}

// BalanceSheetItem is a line item in the balance sheet (alias for snapshot struct).
// Defined in model_report.go for use in both snapshot storage and API response.

// IncomeStatementResponse represents an income statement (profit & loss) report.
type IncomeStatementResponse struct {
	PeriodID      int64           `json:"period_id"`
	Year          int             `json:"year"`
	Month         int             `json:"month"`
	Revenue       decimal.Decimal `json:"revenue"`
	COGS          decimal.Decimal `json:"cogs"`
	SGA           decimal.Decimal `json:"sga"`
	Tax           decimal.Decimal `json:"tax"`
	NonOpIncome   decimal.Decimal `json:"non_op_income"`
	NonOpExpense  decimal.Decimal `json:"non_op_expense"`
	IncomeTax     decimal.Decimal `json:"income_tax"`
	NetProfit     decimal.Decimal `json:"net_profit"`
	GeneratedAt   time.Time       `json:"generated_at"`
}

// MultiPeriodReportRequest is a request for multi-period report comparison.
type MultiPeriodReportRequest struct {
	PeriodIDs []int64 `form:"period_ids" binding:"required" json:"period_ids"`
}

// PeriodSummary is a brief summary of a period for multi-period display.
type PeriodSummary struct {
	PeriodID int64  `json:"period_id"`
	Year     int    `json:"year"`
	Month    int    `json:"month"`
	Label    string `json:"label"` // e.g. "2026-01"
}

// MultiPeriodBalanceSheetResponse is a multi-period balance sheet comparison.
type MultiPeriodBalanceSheetResponse struct {
	Periods []PeriodSummary       `json:"periods"`
	Items   []MultiPeriodBalanceItem `json:"items"`
}

// MultiPeriodBalanceItem is a balance sheet item with values across multiple periods.
type MultiPeriodBalanceItem struct {
	AccountID int64                `json:"account_id"`
	Code      string               `json:"code"`
	Name      string               `json:"name"`
	Values    []decimal.Decimal    `json:"values"` // one per period
	Diff      decimal.Decimal      `json:"diff"`   // last - first
	PctChange decimal.Decimal      `json:"pct_change"`
}

// VATCalculationResponse is the monthly VAT calculation result.
type VATCalculationResponse struct {
	PeriodID      int64              `json:"period_id"`
	Year          int                `json:"year"`
	Month         int                `json:"month"`
	OutputTax    decimal.Decimal     `json:"output_tax"`
	InputTax     decimal.Decimal     `json:"input_tax"`
	NetVAT       decimal.Decimal     `json:"net_vat"`
	InputInvoices []InvoiceRef       `json:"input_invoices,omitempty"`
	OutputInvoices []InvoiceRef      `json:"output_invoices,omitempty"`
}

// InvoiceRef is a reference to an invoice in tax calculations.
type InvoiceRef struct {
	ID          int64           `json:"id"`
	Code        string          `json:"code"`
	Number      string          `json:"number"`
	Date        string          `json:"date"`
	Amount      decimal.Decimal `json:"amount"`
	TaxAmount   decimal.Decimal `json:"tax_amount"`
	TaxRate     decimal.Decimal `json:"tax_rate"`
}

// CITCalculationResponse is the quarterly CIT (企业所得税) estimate.
type CITCalculationResponse struct {
	Year             int              `json:"year"`
	Quarter          int              `json:"quarter"` // 1-4
	RevenueYTD      decimal.Decimal  `json:"revenue_ytd"`
	CostsYTD        decimal.Decimal  `json:"costs_ytd"`
	ExpensesYTD     decimal.Decimal  `json:"expenses_ytd"`
	ProfitBeforeTax decimal.Decimal  `json:"profit_before_tax"`
	TaxRate         decimal.Decimal  `json:"tax_rate"`  // 0.05 for small enterprise
	EstimatedCIT    decimal.Decimal  `json:"estimated_cit"`
}

// TaxDeclarationExport combines VAT and CIT data for export.
type TaxDeclarationExport struct {
	VAT          VATCalculationResponse  `json:"vat"`
	CIT          CITCalculationResponse  `json:"cit"`
	GeneratedAt  time.Time              `json:"generated_at"`
}

// AccountBalanceByPeriod holds an account's balance for a specific period.
type AccountBalanceByPeriod struct {
	AccountID int64
	Code      string
	Name      string
	Balance   decimal.Decimal
}
