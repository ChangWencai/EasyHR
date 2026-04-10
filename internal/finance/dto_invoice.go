package finance

import "time"

// CreateInvoiceRequest is the request body for creating an invoice.
type CreateInvoiceRequest struct {
	InvoiceType InvoiceType `json:"invoice_type" binding:"required,oneof=INPUT OUTPUT"`
	Code        string      `json:"code"`
	Number      string      `json:"number" binding:"required"`
	Date        string      `json:"date" binding:"required"` // YYYY-MM-DD
	Amount      string      `json:"amount" binding:"required"` // decimal string
	TaxRate     string      `json:"tax_rate" binding:"required"` // decimal string, e.g. "0.13"
	Status      InvoiceStatus `json:"status" binding:"omitempty,oneof=unverified verified deducted"`
	Remark      string      `json:"remark"`
}

// UpdateInvoiceRequest is the request body for updating an invoice.
type UpdateInvoiceRequest struct {
	Code        *string         `json:"code"`
	Number      *string         `json:"number"`
	Date        *string         `json:"date"` // YYYY-MM-DD
	Amount      *string         `json:"amount"`
	TaxRate     *string         `json:"tax_rate"`
	Status      *InvoiceStatus  `json:"status"`
	Remark      *string         `json:"remark"`
}

// LinkVoucherRequest links an invoice to a voucher.
type LinkVoucherRequest struct {
	VoucherID int64 `json:"voucher_id" binding:"required"`
}

// ListInvoiceRequest is the query params for listing invoices.
type ListInvoiceRequest struct {
	Type     *InvoiceType  `form:"type"`
	Status   *InvoiceStatus `form:"status"`
	Year     int           `form:"year"`
	Month    int           `form:"month"`
	Page     int           `form:"page,default=1"`
	Limit    int           `form:"limit,default=20"`
}

// InvoiceResponse is the API response format for an invoice.
type InvoiceResponse struct {
	ID          int64           `json:"id"`
	InvoiceType InvoiceType     `json:"invoice_type"`
	Code        string          `json:"code"`
	Number      string          `json:"number"`
	Date        string          `json:"date"`
	Amount      string          `json:"amount"`
	TaxRate     string          `json:"tax_rate"`
	TaxAmount   string          `json:"tax_amount"`
	Status      InvoiceStatus   `json:"status"`
	VoucherID   *int64          `json:"voucher_id,omitempty"`
	VoucherNo   string          `json:"voucher_no,omitempty"`
	Remark      string          `json:"remark"`
	CreatedAt   time.Time       `json:"created_at"`
}

// MonthlyTaxSummary is the response for monthly VAT summary.
type MonthlyTaxSummary struct {
	Year             int     `json:"year"`
	Month            int     `json:"month"`
	OutputTaxSum     string  `json:"output_tax_sum"`      // 销项税额合计
	InputTaxSum      string  `json:"input_tax_sum"`      // 已认证进项税额合计
	OutputAmountSum  string  `json:"output_amount_sum"`  // 销项含税金额合计
	InputAmountSum   string  `json:"input_amount_sum"`   // 进项含税金额合计
	NetVAT           string  `json:"net_vat"`             // 应纳税额 = 销项 - 进项
}
