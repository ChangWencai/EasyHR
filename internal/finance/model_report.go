package finance

import (
	"time"

	"github.com/shopspring/decimal"
	"github.com/wencai/easyhr/internal/common/model"
)

// ReportType represents the type of a financial report.
type ReportType string

const (
	ReportTypeBalanceSheet     ReportType = "balance_sheet"
	ReportTypeIncomeStatement  ReportType = "income_statement"
)

// ReportSnapshot represents a snapshot of a financial report generated at period close.
type ReportSnapshot struct {
	model.BaseModel
	PeriodID     int64       `gorm:"not null;index:idx_snapshot_org_period_type,priority:1" json:"period_id"`
	ReportType   ReportType  `gorm:"type:varchar(30);not null;index:idx_snapshot_org_period_type,priority:2" json:"report_type"`
	Data         string      `gorm:"type:text;not null" json:"data"` // JSON string holding the computed report
	GeneratedBy  int64       `gorm:"not null" json:"generated_by"`
	GeneratedAt  time.Time   `gorm:"not null" json:"generated_at"`
	IsValid      bool        `gorm:"default:true" json:"is_valid"` // false when period is reopened
}

// TableName returns the table name for ReportSnapshot.
func (ReportSnapshot) TableName() string {
	return "report_snapshots"
}

// BalanceSheetData holds the computed balance sheet values stored in ReportSnapshot.Data.
type BalanceSheetData struct {
	Year            int                              `json:"year"`
	Month           int                              `json:"month"`
	Assets          []BalanceSheetItem               `json:"assets"`
	Liabilities     []BalanceSheetItem               `json:"liabilities"`
	Equity          []BalanceSheetItem               `json:"equity"`
	AssetTotal      decimal.Decimal                  `json:"asset_total"`
	LiabilityTotal  decimal.Decimal                  `json:"liability_total"`
	EquityTotal     decimal.Decimal                  `json:"equity_total"`
	GeneratedAt     time.Time                        `json:"generated_at"`
}

// BalanceSheetItem is a single line item in the balance sheet.
type BalanceSheetItem struct {
	AccountID   int64           `json:"account_id"`
	Code        string           `json:"code"`
	Name        string           `json:"name"`
	Balance     decimal.Decimal `json:"balance"`
	IsParent    bool            `json:"is_parent,omitempty"` // true for aggregate lines like "货币资金"
}

// IncomeStatementData holds the computed income statement values stored in ReportSnapshot.Data.
type IncomeStatementData struct {
	Year          int              `json:"year"`
	Month         int              `json:"month"`
	Revenue       decimal.Decimal  `json:"revenue"`
	COGS          decimal.Decimal  `json:"cogs"`
	SGA           decimal.Decimal  `json:"sga"`
	Tax           decimal.Decimal  `json:"tax"`
	NonOpIncome   decimal.Decimal  `json:"non_op_income"`
	NonOpExpense  decimal.Decimal  `json:"non_op_expense"`
	IncomeTax     decimal.Decimal  `json:"income_tax"`
	NetProfit     decimal.Decimal  `json:"net_profit"`
	GeneratedAt   time.Time        `json:"generated_at"`
}
