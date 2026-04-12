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
	InvoiceType InvoiceType   `gorm:"type:varchar(20);not null;index:idx_invoice_org_type,priority:2;comment:发票类型（INPUT/OUTPUT）" json:"invoice_type"`
	Code        string        `gorm:"type:varchar(50);comment:发票代码" json:"code"`
	Number      string        `gorm:"type:varchar(50);comment:发票号码" json:"number"`
	Date        time.Time     `gorm:"type:date;not null;comment:开票日期" json:"date"`
	Amount      decimal.Decimal `gorm:"type:varchar(50);not null;comment:含税金额" json:"amount"`
	TaxRate     decimal.Decimal `gorm:"type:varchar(20);not null;comment:税率" json:"tax_rate"`
	TaxAmount   decimal.Decimal `gorm:"type:varchar(50);not null;comment:税额" json:"tax_amount"`
	Status      InvoiceStatus `gorm:"type:varchar(20);default:'unverified';comment:认证状态（unverified/verified/deducted）" json:"status"`
	VoucherID   *int64        `gorm:"index:idx_invoice_voucher;comment:关联凭证ID" json:"voucher_id,omitempty"`
	Remark      string        `gorm:"type:varchar(500);comment:备注" json:"remark"`
}

// TableName returns the table name for Invoice.
func (Invoice) TableName() string {
	return "invoices"
}
