package employee

import (
	"time"

	"github.com/wencai/easyhr/internal/common/model"
)

// Contract 员工劳动合同模型
// 支持固定期限/无固定期限/实习三种合同类型
// 合同生命周期：draft -> pending_sign -> signed -> active -> terminated/expired
type Contract struct {
	model.BaseModel
	EmployeeID      int64      `gorm:"column:employee_id;index;not null" json:"employee_id"`
	ContractType    string     `gorm:"column:contract_type;type:varchar(20);not null" json:"contract_type"` // fixed_term/indefinite/intern
	StartDate       time.Time  `gorm:"column:start_date;type:date;not null" json:"start_date"`
	EndDate         *time.Time `gorm:"column:end_date;type:date" json:"end_date"` // 无固定期限为 nil
	Salary          float64    `gorm:"column:salary;type:decimal(10,2)" json:"salary"`
	Status          string     `gorm:"column:status;type:varchar(20);not null;default:draft" json:"status"` // draft/pending_sign/signed/active/terminated/expired
	PDFURL          string     `gorm:"column:pdf_url;type:varchar(500)" json:"pdf_url"`
	SignedPDFURL    string     `gorm:"column:signed_pdf_url;type:varchar(500)" json:"signed_pdf_url"`
	SignDate        *time.Time `gorm:"column:sign_date;type:date" json:"sign_date"`
	TerminateDate   *time.Time `gorm:"column:terminate_date;type:date" json:"terminate_date"`
	TerminateReason string     `gorm:"column:terminate_reason;type:varchar(500)" json:"terminate_reason"`
}

// TableName 指定表名
func (Contract) TableName() string {
	return "contracts"
}

// ContractStatus 合同状态常量
const (
	ContractStatusDraft        = "draft"         // 草稿
	ContractStatusPendingSign  = "pending_sign"  // 待签署
	ContractStatusSigned       = "signed"        // 已签署
	ContractStatusActive       = "active"        // 履行中
	ContractStatusTerminated   = "terminated"    // 已终止
	ContractStatusExpired      = "expired"       // 已到期
)

// ContractType 合同类型常量
const (
	ContractTypeFixedTerm  = "fixed_term"  // 固定期限
	ContractTypeIndefinite = "indefinite"  // 无固定期限
	ContractTypeIntern     = "intern"      // 实习
)
