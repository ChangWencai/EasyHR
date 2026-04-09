package salary

import (
	"time"

	"github.com/wencai/easyhr/internal/common/model"
)

// ========== 薪资模板模型 ==========

// SalaryTemplateItem 薪资项模板（全局预置 + 企业启用状态）
// OrgID=0 为全局预置项，OrgID=企业ID 为企业启用/禁用覆盖
type SalaryTemplateItem struct {
	model.BaseModel
	Name      string `gorm:"column:name;type:varchar(50);not null" json:"name"`
	Type      string `gorm:"column:type;type:varchar(20);not null" json:"type"` // income/deduction
	SortOrder int    `gorm:"column:sort_order;not null;default:0" json:"sort_order"`
	IsRequired bool  `gorm:"column:is_required;not null;default:false" json:"is_required"`
	IsEnabled  bool  `gorm:"column:is_enabled;not null" json:"is_enabled"`
}

func (SalaryTemplateItem) TableName() string { return "salary_template_items" }

// ========== 员工薪资项模型 ==========

// SalaryItem 员工薪资项金额（每员工每启用的薪资项一条记录）
type SalaryItem struct {
	model.BaseModel
	EmployeeID     int64   `gorm:"column:employee_id;index;not null" json:"employee_id"`
	TemplateItemID int64   `gorm:"column:template_item_id;index;not null" json:"template_item_id"`
	Amount         float64 `gorm:"column:amount;type:decimal(10,2);not null;default:0" json:"amount"`
	EffectiveMonth string  `gorm:"column:effective_month;type:varchar(7);not null" json:"effective_month"` // "2026-04"
}

func (SalaryItem) TableName() string { return "salary_items" }

// ========== 工资核算模型 ==========

// PayrollRecord 月度工资核算主表（每员工每月一条）
type PayrollRecord struct {
	model.BaseModel
	EmployeeID      int64      `gorm:"column:employee_id;index;not null" json:"employee_id"`
	EmployeeName    string     `gorm:"column:employee_name;type:varchar(50);not null" json:"employee_name"`
	Year            int        `gorm:"column:year;not null" json:"year"`
	Month           int        `gorm:"column:month;not null" json:"month"`
	Status          string     `gorm:"column:status;type:varchar(20);not null;default:draft" json:"status"` // draft/calculated/confirmed/paid
	GrossIncome     float64    `gorm:"column:gross_income;type:decimal(12,2);not null;default:0" json:"gross_income"`
	SIDeduction     float64    `gorm:"column:si_deduction;type:decimal(12,2);not null;default:0" json:"si_deduction"`
	Tax             float64    `gorm:"column:tax;type:decimal(12,2);not null;default:0" json:"tax"`
	TotalDeductions float64    `gorm:"column:total_deductions;type:decimal(12,2);not null;default:0" json:"total_deductions"`
	NetIncome       float64    `gorm:"column:net_income;type:decimal(12,2);not null;default:0" json:"net_income"`
	PayMethod       string     `gorm:"column:pay_method;type:varchar(20)" json:"pay_method"` // bank_transfer/cash/other
	PayDate         *time.Time `gorm:"column:pay_date" json:"pay_date"`
	PayNote         string     `gorm:"column:pay_note;type:varchar(200)" json:"pay_note"`
}

func (PayrollRecord) TableName() string { return "payroll_records" }

// PayrollItem 工资核算明细（每员工多条，快照薪资项名称和金额）
type PayrollItem struct {
	model.BaseModel
	PayrollRecordID int64   `gorm:"column:payroll_record_id;index;not null" json:"payroll_record_id"`
	ItemName        string  `gorm:"column:item_name;type:varchar(50);not null" json:"item_name"`
	ItemType        string  `gorm:"column:item_type;type:varchar(20);not null" json:"item_type"` // income/deduction
	Amount          float64 `gorm:"column:amount;type:decimal(12,2);not null;default:0" json:"amount"`
}

func (PayrollItem) TableName() string { return "payroll_items" }

// ========== 工资单模型 ==========

// PayrollSlip 工资单（含唯一 token 用于 H5 查看链接）
type PayrollSlip struct {
	model.BaseModel
	PayrollRecordID int64      `gorm:"column:payroll_record_id;index;not null" json:"payroll_record_id"`
	EmployeeID      int64      `gorm:"column:employee_id;index;not null" json:"employee_id"`
	Token           string     `gorm:"column:token;type:varchar(64);uniqueIndex;not null" json:"token"`
	PhoneEncrypted  string     `gorm:"column:phone_encrypted;type:varchar(200)" json:"-"`
	PhoneHash       string     `gorm:"column:phone_hash;type:varchar(64);index" json:"-"`
	Status          string     `gorm:"column:status;type:varchar(20);not null;default:pending" json:"status"` // pending/sent/viewed/signed
	SentAt          *time.Time `gorm:"column:sent_at" json:"sent_at"`
	ViewedAt        *time.Time `gorm:"column:viewed_at" json:"viewed_at"`
	SignedAt        *time.Time `gorm:"column:signed_at" json:"signed_at"`
	ExpiresAt       time.Time  `gorm:"column:expires_at;not null" json:"expires_at"`
}

func (PayrollSlip) TableName() string { return "payroll_slips" }

// ========== 状态常量 ==========

const (
	PayrollStatusDraft      = "draft"
	PayrollStatusCalculated = "calculated"
	PayrollStatusConfirmed  = "confirmed"
	PayrollStatusPaid       = "paid"
)

const (
	SlipStatusPending = "pending"
	SlipStatusSent    = "sent"
	SlipStatusViewed  = "viewed"
	SlipStatusSigned  = "signed"
)
