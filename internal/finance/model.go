package finance

import (
	"time"

	"github.com/shopspring/decimal"
	"github.com/wencai/easyhr/internal/common/model"
)

// NormalBalance represents the normal balance side of an account.
type NormalBalance string

const (
	NormalBalanceDebit  NormalBalance = "debit"
	NormalBalanceCredit NormalBalance = "credit"
)

// DCType represents debit or credit direction.
type DCType string

const (
	DCTypeDebit  DCType = "debit"
	DCTypeCredit DCType = "credit"
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

// Account represents an accounting account.
type Account struct {
	model.BaseModel
	Code          string         `gorm:"type:varchar(20);not null;index" json:"code"`
	Name          string         `gorm:"type:varchar(100);not null" json:"name"`
	Category      string         `gorm:"type:varchar(50);not null" json:"category"`
	NormalBalance NormalBalance  `gorm:"type:varchar(10);not null" json:"normal_balance"`
	IsActive      bool           `gorm:"default:true" json:"is_active"`
	IsSystem      bool           `gorm:"default:false" json:"is_system"`
	ParentID      *int64         `gorm:"index" json:"parent_id,omitempty"`
	Level         int            `gorm:"default:1" json:"level"`
}

// Period represents an accounting period.
type Period struct {
	model.BaseModel
	Year   int          `gorm:"not null" json:"year"`
	Month  int          `gorm:"not null" json:"month"`
	Status PeriodStatus `gorm:"type:varchar(10);default:'OPEN'" json:"status"`
}

// Voucher represents an accounting voucher (凭证).
type Voucher struct {
	model.BaseModel
	PeriodID   int64         `gorm:"not null;index" json:"period_id"`
	VoucherNo  string        `gorm:"type:varchar(20);index" json:"voucher_no"`
	Date       time.Time     `gorm:"not null" json:"date"`
	Status     VoucherStatus `gorm:"type:varchar(20);default:'draft'" json:"status"`
	SourceType SourceType    `gorm:"type:varchar(20)" json:"source_type"`
	SourceID   *int64        `gorm:"index" json:"source_id,omitempty"`
	Summary    string        `gorm:"type:varchar(500)" json:"summary"`
	ReversalOf *int64        `gorm:"index" json:"reversal_of,omitempty"`
	Entries    []JournalEntry `gorm:"foreignKey:VoucherID" json:"entries"`
}

// JournalEntry represents a debit or credit entry in a voucher.
type JournalEntry struct {
	model.BaseModel
	VoucherID int64          `gorm:"not null;index" json:"voucher_id"`
	AccountID int64          `gorm:"not null;index" json:"account_id"`
	DC        DCType         `gorm:"type:varchar(10);not null" json:"dc"`
	Amount    decimal.Decimal `gorm:"type:varchar(50);not null" json:"amount"`
	Summary   string         `gorm:"type:varchar(200)" json:"summary"`
}
