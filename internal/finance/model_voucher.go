package finance

import (
	"time"

	"github.com/shopspring/decimal"
	"github.com/wencai/easyhr/internal/common/model"
)

// Voucher represents an accounting voucher (凭证).
type Voucher struct {
	model.BaseModel
	PeriodID    int64         `gorm:"not null;index;comment:会计期间ID" json:"period_id"`
	VoucherNo   string        `gorm:"type:varchar(20);comment:凭证号" json:"voucher_no"`
	Date        time.Time     `gorm:"type:date;not null;comment:凭证日期" json:"date"`
	Status      VoucherStatus `gorm:"type:varchar(20);default:'draft';comment:凭证状态（draft/posted）" json:"status"`
	SourceType  SourceType    `gorm:"type:varchar(20);comment:来源类型" json:"source_type"`
	SourceID    *int64        `gorm:"index;comment:来源单据ID" json:"source_id,omitempty"`
	Summary     string        `gorm:"type:varchar(500);comment:凭证摘要" json:"summary"`
	ReversalOf  *int64        `gorm:"index;comment:被冲销凭证ID" json:"reversal_of,omitempty"`
	Entries     []JournalEntry `gorm:"foreignKey:VoucherID;comment:借贷分录" json:"entries,omitempty"`
}

// TableName returns the table name for Voucher.
func (Voucher) TableName() string {
	return "vouchers"
}

// JournalEntry represents a debit or credit line in a voucher (借贷分录).
type JournalEntry struct {
	model.BaseModel
	VoucherID int64           `gorm:"not null;index:idx_je_voucher,priority:1;comment:凭证ID" json:"voucher_id"`
	AccountID int64           `gorm:"not null;index:idx_je_org_account,priority:2;comment:科目ID" json:"account_id"`
	DC        DCType          `gorm:"type:varchar(10);not null;comment:借贷方向（debit/credit）" json:"dc"`
	Amount    decimal.Decimal `gorm:"type:varchar(50);not null;comment:金额" json:"amount"`
	Summary   string          `gorm:"type:varchar(200);comment:分录摘要" json:"summary"`
}

// TableName returns the table name for JournalEntry.
func (JournalEntry) TableName() string {
	return "journal_entries"
}
