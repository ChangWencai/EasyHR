package finance

import (
	"encoding/json"
	"time"

	"github.com/shopspring/decimal"
	"github.com/wencai/easyhr/internal/common/model"
)

// ExpenseType represents the type of expense reimbursement.
type ExpenseType string

const (
	ExpenseTypeTravel       ExpenseType = "travel"       // 差旅费
	ExpenseTypeTransport    ExpenseType = "transport"   // 交通费
	ExpenseTypeEntertainment ExpenseType = "entertainment" // 招待费
	ExpenseTypeOffice       ExpenseType = "office"      // 办公费
	ExpenseTypeOther        ExpenseType = "other"       // 其他
)

// ExpenseStatus represents the status of an expense reimbursement.
type ExpenseStatus string

const (
	ExpenseStatusPending   ExpenseStatus = "pending"   // 待审批
	ExpenseStatusApproved  ExpenseStatus = "approved"  // 已批准
	ExpenseStatusRejected  ExpenseStatus = "rejected"  // 已驳回
	ExpenseStatusPaid      ExpenseStatus = "paid"      // 已支付
)

// ExpenseReimbursement represents an employee expense reimbursement record (费用报销).
type ExpenseReimbursement struct {
	model.BaseModel
	EmployeeID   int64           `gorm:"column:employee_id;not null;index:idx_expense_employee,priority:1" json:"employee_id"`
	Amount       decimal.Decimal `gorm:"type:varchar(50);not null" json:"amount"` // 报销金额
	ExpenseType  ExpenseType     `gorm:"type:varchar(20);not null" json:"expense_type"`
	Description  string          `gorm:"type:varchar(500)" json:"description"`
	Attachments  string          `gorm:"type:varchar(1000)" json:"attachments"` // JSON array of OSS URLs, max 9
	Status       ExpenseStatus  `gorm:"type:varchar(20);default:'pending';index:idx_expense_org_status,priority:2" json:"status"`
	ApproverID   *int64          `gorm:"index" json:"approver_id,omitempty"`
	ApprovedAt   *time.Time      `json:"approved_at,omitempty"`
	ApprovedNote string          `gorm:"type:varchar(500)" json:"approved_note,omitempty"`
	RejectedAt   *time.Time      `json:"rejected_at,omitempty"`
	RejectedNote string          `gorm:"type:varchar(500)" json:"rejected_note,omitempty"`
	PaidAt       *time.Time      `json:"paid_at,omitempty"`
	PaidNote     string          `gorm:"type:varchar(500)" json:"paid_note,omitempty"`
	VoucherID    *int64          `gorm:"index" json:"voucher_id,omitempty"` // 关联的费用凭证
}

// TableName returns the table name for ExpenseReimbursement.
func (ExpenseReimbursement) TableName() string {
	return "expense_reimbursements"
}

// AttachmentURLs parses the Attachments JSON field and returns URLs.
func (e *ExpenseReimbursement) AttachmentURLs() []string {
	if e.Attachments == "" {
		return nil
	}
	var urls []string
	if err := json.Unmarshal([]byte(e.Attachments), &urls); err != nil {
		return nil
	}
	return urls
}
