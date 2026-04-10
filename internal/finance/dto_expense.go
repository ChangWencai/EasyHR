package finance

import (
	"time"
)

// CreateExpenseRequest is the request body for creating/submitting an expense.
type CreateExpenseRequest struct {
	EmployeeID  int64          `json:"employee_id" binding:"required"`
	Amount      string         `json:"amount" binding:"required"` // decimal string
	ExpenseType ExpenseType    `json:"expense_type" binding:"required,oneof=travel transport entertainment office other"`
	Description string         `json:"description" binding:"required"`
	Attachments []string        `json:"attachments"` // max 9 OSS URLs
}

// ApproveExpenseRequest is the request to approve an expense.
type ApproveExpenseRequest struct {
	Note string `json:"note"` // optional approval note
}

// RejectExpenseRequest is the request to reject an expense.
type RejectExpenseRequest struct {
	Note string `json:"note"` // required rejection reason
}

// MarkPaidRequest is the request to mark an expense as paid.
type MarkPaidRequest struct {
	Note string `json:"note"` // optional payment note
}

// ExpenseResponse is the API response format for an expense reimbursement.
type ExpenseResponse struct {
	ID           int64         `json:"id"`
	EmployeeID   int64         `json:"employee_id"`
	Amount       string        `json:"amount"`
	ExpenseType  ExpenseType   `json:"expense_type"`
	Description  string        `json:"description"`
	Attachments  []string      `json:"attachments"`
	Status       ExpenseStatus `json:"status"`
	ApproverID   *int64        `json:"approver_id,omitempty"`
	ApprovedAt   *string       `json:"approved_at,omitempty"`
	ApprovedNote string        `json:"approved_note,omitempty"`
	RejectedAt   *string       `json:"rejected_at,omitempty"`
	RejectedNote string        `json:"rejected_note,omitempty"`
	PaidAt       *string       `json:"paid_at,omitempty"`
	PaidNote     string        `json:"paid_note,omitempty"`
	VoucherID    *int64        `json:"voucher_id,omitempty"`
	VoucherNo    string        `json:"voucher_no,omitempty"`
	CreatedAt    time.Time     `json:"created_at"`
}

// ListExpenseRequest is the query params for listing expenses.
type ListExpenseRequest struct {
	Status     *ExpenseStatus `form:"status"`
	EmployeeID *int64         `form:"employee_id"`
	Page       int            `form:"page,default=1"`
	Limit      int            `form:"limit,default=20"`
}

// MonthlyExpenseSummary is a summary of expenses for dashboard display.
type MonthlyExpenseSummary struct {
	Year        int    `json:"year"`
	Month       int    `json:"month"`
	TotalCount  int64  `json:"total_count"`
	TotalAmount string `json:"total_amount"`
	PendingCount int64 `json:"pending_count"`
	ApprovedAmount string `json:"approved_amount"`
	PaidAmount  string `json:"paid_amount"`
}
