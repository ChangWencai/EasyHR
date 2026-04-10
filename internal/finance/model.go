package finance

// NormalBalance represents the normal balance side of an account.
type NormalBalance string

const (
	NormalBalanceDebit  NormalBalance = "debit"
	NormalBalanceCredit NormalBalance = "credit"
)

// DCType represents debit or credit direction.
type DCType string

const (
	DCDebit  DCType = "debit"
	DCCredit DCType = "credit"
)

// VoucherStatus represents the status of a voucher.
type VoucherStatus string

const (
	VoucherStatusDraft     VoucherStatus = "draft"
	VoucherStatusSubmitted VoucherStatus = "submitted"
	VoucherStatusAudited   VoucherStatus = "audited"
	VoucherStatusClosed    VoucherStatus = "closed"
)

// PeriodStatus represents the status of an accounting period.
type PeriodStatus string

const (
	PeriodStatusOpen   PeriodStatus = "OPEN"
	PeriodStatusLocked PeriodStatus = "LOCKED"
	PeriodStatusClosed PeriodStatus = "CLOSED"
)

// SourceType represents the source of a voucher.
type SourceType string

const (
	SourceTypeManual   SourceType = "manual"
	SourceTypePayroll  SourceType = "payroll"
	SourceTypeExpense  SourceType = "expense"
)

// AccountCategory represents the five major categories of accounting accounts.
type AccountCategory string

const (
	AccountCategoryAsset     AccountCategory = "ASSET"
	AccountCategoryLiability AccountCategory = "LIABILITY"
	AccountCategoryEquity    AccountCategory = "EQUITY"
	AccountCategoryCost      AccountCategory = "COST"
	AccountCategoryProfit    AccountCategory = "PROFIT"
)
