package finance

import (
	"time"

	"github.com/shopspring/decimal"
	"github.com/wencai/easyhr/internal/common/model"
)

// InvoiceType represents the type of invoice.
type InvoiceType string

const (
	InvoiceTypeInput  InvoiceType = "INPUT"  // 进项发票
	InvoiceTypeOutput InvoiceType = "OUTPUT" // 销项发票
)

// InvoiceStatus represents the verification status of an invoice.
type InvoiceStatus string

const (
	InvoiceStatusUnverified InvoiceStatus = "unverified" // 未认证
	InvoiceStatusVerified    InvoiceStatus = "verified"   // 已认证
	InvoiceStatusDeducted    InvoiceStatus = "deducted"   // 已抵扣
)

// Invoice represents a tax invoice (发票).
type Invoice struct {
	model.BaseModel
	InvoiceType InvoiceType   `gorm:"type:varchar(20);not null;index:idx_invoice_org_type,priority:2" json:"invoice_type"`
	Code      string          `gorm:"type:varchar(50)" json:"code"`         // 发票代码
	Number    string          `gorm:"type:varchar(50)" json:"number"`       // 发票号码
	Date      time.Time       `gorm:"type:date;not null" json:"date"`       // 开票日期
	Amount    decimal.Decimal `gorm:"type:varchar(50);not null" json:"amount"` // 含税金额
	TaxRate   decimal.Decimal `gorm:"type:varchar(20);not null" json:"tax_rate"` // 税率，如 0.13
	TaxAmount decimal.Decimal `gorm:"type:varchar(50);not null" json:"tax_amount"` // 税额
	Status    InvoiceStatus   `gorm:"type:varchar(20);default:'unverified'" json:"status"`
	VoucherID *int64          `gorm:"index:idx_invoice_voucher" json:"voucher_id,omitempty"` // 关联凭证
	Remark    string          `gorm:"type:varchar(500)" json:"remark"`
}

// TableName returns the table name for Invoice.
func (Invoice) TableName() string {
	return "invoices"
}
