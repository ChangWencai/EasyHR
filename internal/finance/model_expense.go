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
	EmployeeID   int64           `gorm:"column:employee_id;not null;index:idx_expense_employee,priority:1;comment:申请人ID" json:"employee_id"`
	Amount       decimal.Decimal `gorm:"type:varchar(50);not null;comment:报销金额" json:"amount"`
	ExpenseType  ExpenseType     `gorm:"type:varchar(20);not null;comment:费用类型（travel/transport/entertainment/office/other）" json:"expense_type"`
	Description  string          `gorm:"type:varchar(500);comment:费用说明" json:"description"`
	Attachments  string          `gorm:"type:varchar(1000);comment:附件URL列表（JSON格式，最多9张）" json:"attachments"`
	Status       ExpenseStatus  `gorm:"type:varchar(20);default:'pending';index:idx_expense_org_status,priority:2;comment:审批状态（pending/approved/rejected/paid）" json:"status"`
	ApproverID   *int64          `gorm:"index;comment:审批人ID" json:"approver_id,omitempty"`
	ApprovedAt   *time.Time      `gorm:"column:approved_at;comment:审批时间" json:"approved_at,omitempty"`
	ApprovedNote string          `gorm:"type:varchar(500);comment:审批备注" json:"approved_note,omitempty"`
	RejectedAt   *time.Time      `gorm:"column:rejected_at;comment:驳回时间" json:"rejected_at,omitempty"`
	RejectedNote string          `gorm:"type:varchar(500);comment:驳回原因" json:"rejected_note,omitempty"`
	PaidAt       *time.Time      `gorm:"column:paid_at;comment:支付时间" json:"paid_at,omitempty"`
	PaidNote     string          `gorm:"type:varchar(500);comment:支付备注" json:"paid_note,omitempty"`
	VoucherID    *int64          `gorm:"index;comment:关联凭证ID" json:"voucher_id,omitempty"`
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
