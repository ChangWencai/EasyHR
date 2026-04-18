package salary

import (
	"time"
)

// SalaryAdjustment 调薪记录（INSERT ONLY，禁止 UPDATE）
type SalaryAdjustment struct {
	ID             int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	OrgID          int64      `gorm:"column:org_id;index;not null" json:"org_id"`
	EmployeeID     *int64     `gorm:"column:employee_id;index;comment:员工ID（单人调薪）" json:"employee_id"`
	DepartmentID   *int64     `gorm:"column:department_id;index;comment:部门ID（普调）" json:"department_id"`
	Type           string     `gorm:"column:type;type:varchar(20);not null;comment:类型（individual/department）" json:"type"`
	EffectiveMonth string     `gorm:"column:effective_month;type:varchar(7);not null;comment:生效月份（YYYY-MM）" json:"effective_month"`
	AdjustmentType string     `gorm:"column:adjustment_type;type:varchar(20);not null;comment:调整类型（base_salary/allowance/bonus/year_end_bonus/other）" json:"adjustment_type"`
	AdjustBy       string     `gorm:"column:adjust_by;type:varchar(10);not null;comment:调整方式（amount/ratio）" json:"adjust_by"`
	OldValue       float64    `gorm:"column:old_value;type:decimal(12,2);not null;default:0;comment:调整前值" json:"old_value"`
	NewValue       float64    `gorm:"column:new_value;type:decimal(12,2);not null;default:0;comment:调整后值" json:"new_value"`
	Status         string     `gorm:"column:status;type:varchar(20);not null;default:active;comment:状态" json:"status"`
	CreatedBy      int64      `gorm:"column:created_by;comment:创建人ID" json:"created_by"`
	CreatedAt      time.Time  `gorm:"column:created_at;autoCreateTime" json:"created_at"`
}

func (SalaryAdjustment) TableName() string { return "salary_adjustments" }

// 确保 SalaryAdjustment 不使用 BaseModel（INSERT ONLY 不需要 updated_by/deleted_at）
// 但保留 org_id 用于多租户隔离

// AdjustmentType constants
const (
	AdjustmentTypeIndividual  = "individual"
	AdjustmentTypeDepartment  = "department"
	AdjustmentByAmount        = "amount"
	AdjustmentByRatio         = "ratio"
	AdjustmentTargetBaseSalary   = "base_salary"
	AdjustmentTargetAllowance    = "allowance"
	AdjustmentTargetBonus        = "bonus"
	AdjustmentTargetYearEndBonus = "year_end_bonus"
	AdjustmentTargetOther        = "other"
)
