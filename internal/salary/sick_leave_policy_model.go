package salary

import "time"

// SickLeavePolicy 病假系数策略（城市 x 工龄档位 x 系数）
type SickLeavePolicy struct {
	ID            int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	City          string    `gorm:"column:city;type:varchar(50);not null;comment:城市" json:"city"`
	TenureBucket  string    `gorm:"column:tenure_bucket;type:varchar(20);not null;comment:工龄档位（within_6months/over_6months）" json:"tenure_bucket"`
	Coefficient   float64   `gorm:"column:coefficient;type:decimal(4,2);not null;comment:病假系数" json:"coefficient"`
	CreatedAt     time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
}

func (SickLeavePolicy) TableName() string { return "sick_leave_policies" }

// SalarySlipSendLog 工资条发送日志（Plan 07-03 使用）
type SalarySlipSendLog struct {
	ID              int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	OrgID           int64      `gorm:"column:org_id;index;not null" json:"org_id"`
	PayrollRecordID int64      `gorm:"column:payroll_record_id;index;not null" json:"payroll_record_id"`
	EmployeeID      int64      `gorm:"column:employee_id;index;not null" json:"employee_id"`
	Channel         string     `gorm:"column:channel;type:varchar(20);not null;comment:渠道（miniapp/sms/h5）" json:"channel"`
	Status          string     `gorm:"column:status;type:varchar(20);not null;default:pending;comment:状态（pending/sending/sent/failed）" json:"status"`
	ErrorMessage    string     `gorm:"column:error_message;type:text;comment:错误信息" json:"error_message"`
	SentAt          *time.Time `gorm:"column:sent_at;comment:发送时间" json:"sent_at"`
	CreatedAt       time.Time  `gorm:"column:created_at;autoCreateTime" json:"created_at"`
}

func (SalarySlipSendLog) TableName() string { return "salary_slip_send_logs" }

// 工龄档位常量
const (
	TenureBucketWithin6Months = "within_6months"
	TenureBucketOver6Months   = "over_6months"
)
