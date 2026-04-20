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
	EmployeeID      int64      `gorm:"column:employee_id;index;not null;comment:员工ID，外键到employees.id" json:"employee_id"`
	ContractType    string     `gorm:"column:contract_type;type:varchar(20);not null;comment:合同类型（fixed_term/indefinite/intern）" json:"contract_type"`
	StartDate       time.Time  `gorm:"column:start_date;type:date;not null;comment:合同开始日期" json:"start_date"`
	EndDate         *time.Time `gorm:"column:end_date;type:date;comment:合同结束日期（无固定期限为空）" json:"end_date"`
	Salary          float64    `gorm:"column:salary;type:decimal(10,2);comment:月薪" json:"salary"`
	Status          string     `gorm:"column:status;type:varchar(20);not null;default:draft;comment:状态（draft/pending_sign/signed/active/terminated/expired）" json:"status"`
	PDFURL          string     `gorm:"column:pdf_url;type:varchar(500);comment:合同PDF模板URL" json:"pdf_url"`
	SignedPDFURL    string     `gorm:"column:signed_pdf_url;type:varchar(500);comment:已签署合同PDF URL" json:"signed_pdf_url"`
	SignDate        *time.Time `gorm:"column:sign_date;type:date;comment:签署日期" json:"sign_date"`
	TerminateDate   *time.Time `gorm:"column:terminate_date;type:date;comment:终止日期" json:"terminate_date"`
	TerminateReason string     `gorm:"column:terminate_reason;type:varchar(500);comment:终止原因" json:"terminate_reason"`
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

// SignCodeExpiry 验证码有效期：5分钟
const SignCodeExpiry = 5 * time.Minute

// SignLinkExpiry 签署链接有效期：7天
const SignLinkExpiry = 7 * 24 * time.Hour

// SignTokenExpiry 签署 Token 有效期：30分钟（验证码校验后，用于确认签署）
const SignTokenExpiry = 30 * time.Minute

// ContractSignCode 签署验证码记录
type ContractSignCode struct {
	model.BaseModel
	ContractID int64     `gorm:"column:contract_id;index;not null" json:"contract_id"`
	Phone     string    `gorm:"column:phone;type:varchar(20);not null;index" json:"phone"`
	Code      string    `gorm:"column:code;type:varchar(6);not null" json:"code"` // 6位纯数字
	ExpiresAt time.Time `gorm:"column:expires_at;not null;index" json:"expires_at"`
	Verified  bool      `gorm:"column:verified;default:false" json:"verified"`
	SignToken string    `gorm:"column:sign_token;type:varchar(64);index" json:"sign_token"` // 校验通过后生成的签署 token
}

// TableName 指定表名
func (ContractSignCode) TableName() string {
	return "contract_sign_codes"
}
