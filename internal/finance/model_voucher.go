package finance

import (
	"time"

	"github.com/shopspring/decimal"
	"github.com/wencai/easyhr/internal/common/model"
)

// Voucher represents an accounting voucher (凭证).
type Voucher struct {
	model.BaseModel
	OrgID     int64         `gorm:"column:org_id;not null;index" json:"-"`
	PeriodID  int64         `gorm:"not null;index" json:"period_id"`
	VoucherNo string        `gorm:"type:varchar(20);index" json:"voucher_no"`
	Date      time.Time     `gorm:"type:date;not null" json:"date"`
	Status    VoucherStatus `gorm:"type:varchar(20);default:'draft'" json:"status"`
	SourceType SourceType   `gorm:"type:varchar(20)" json:"source_type"`
	SourceID  *int64        `gorm:"index" json:"source_id,omitempty"`
	Summary   string        `gorm:"type:varchar(500)" json:"summary"`
	ReversalOf *int64       `gorm:"index" json:"reversal_of,omitempty"`
	Entries   []JournalEntry `gorm:"foreignKey:VoucherID" json:"entries,omitempty"`
}

// TableName returns the table name for Voucher.
func (Voucher) TableName() string {
	return "vouchers"
}

// JournalEntry represents a debit or credit line in a voucher (借贷分录).
type JournalEntry struct {
	model.BaseModel
	OrgID     int64           `gorm:"column:org_id;not null;index:idx_je_org_account,priority:1" json:"-"`
	VoucherID int64           `gorm:"not null;index:idx_je_voucher,priority:1" json:"voucher_id"`
	AccountID int64           `gorm:"not null;index:idx_je_org_account,priority:2" json:"account_id"`
	DC        DCType          `gorm:"type:varchar(10);not null" json:"dc"`
	Amount    decimal.Decimal `gorm:"type:varchar(50);not null" json:"amount"`
	Summary   string          `gorm:"type:varchar(200)" json:"summary"`
}

// TableName returns the table name for JournalEntry.
func (JournalEntry) TableName() string {
	return "journal_entries"
}
