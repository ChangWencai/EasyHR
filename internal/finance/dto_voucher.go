package finance

// JournalEntryInput represents a single debit or credit entry in a create-voucher request.
type JournalEntryInput struct {
	AccountID int64  `json:"account_id" binding:"required"`
	DC        string `json:"dc" binding:"required,oneof=debit credit"`
	Amount    string `json:"amount" binding:"required"`
	Summary   string `json:"summary"`
}

// CreateVoucherRequest is the request body for creating a voucher.
type CreateVoucherRequest struct {
	PeriodID   int64              `json:"period_id" binding:"required"`
	VoucherDate string            `json:"voucher_date" binding:"required"`
	Summary    string             `json:"summary"`
	SourceType SourceType         `json:"source_type"`
	SourceID   *int64             `json:"source_id"`
	Entries    []JournalEntryInput `json:"entries" binding:"required,min=2,dive"`
}

// VoucherResponse is the API response format for a voucher with its entries.
type VoucherResponse struct {
	ID         int64            `json:"id"`
	VoucherNo  string           `json:"voucher_no"`
	PeriodID   int64            `json:"period_id"`
	Date       string           `json:"date"`
	Status     VoucherStatus   `json:"status"`
	SourceType SourceType      `json:"source_type"`
	SourceID   *int64          `json:"source_id,omitempty"`
	Summary    string           `json:"summary"`
	ReversalOf *int64          `json:"reversal_of,omitempty"`
	Entries    []JournalEntryResponse `json:"entries"`
}

// JournalEntryResponse is the API response format for a journal entry.
type JournalEntryResponse struct {
	ID        int64  `json:"id"`
	AccountID int64  `json:"account_id"`
	DC        string `json:"dc"`
	Amount    string `json:"amount"`
	Summary   string `json:"summary"`
}

// ListVoucherRequest is the query params for listing/searching vouchers.
type ListVoucherRequest struct {
	PeriodID  *int64 `form:"period_id"`
	AccountID *int64 `form:"account_id"`
	Keyword   string `form:"keyword"`
	Page      int    `form:"page,default=1"`
	Limit     int    `form:"limit,default=20"`
}

// SubmitVoucherRequest is the request to submit a draft voucher.
type SubmitVoucherRequest struct {
	VoucherID int64 `json:"voucher_id" binding:"required"`
}

// AuditVoucherRequest is the request to audit a submitted voucher.
type AuditVoucherRequest struct {
	VoucherID int64 `json:"voucher_id" binding:"required"`
}

// ReverseVoucherRequest is the request to reverse an audited voucher.
type ReverseVoucherRequest struct {
	VoucherID int64 `json:"voucher_id" binding:"required"`
}

// SubmitVoucher submits a draft voucher.
type SubmitVoucher struct {
	VoucherID int64 `json:"voucher_id"`
}

// AuditVoucher audits a submitted voucher.
type AuditVoucher struct {
	VoucherID int64 `json:"voucher_id"`
}
