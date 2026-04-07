package socialinsurance

import (
	"time"

	"github.com/wencai/easyhr/internal/common/model"
)

// 提醒类型常量
const (
	ReminderTypePaymentDue = "payment_due"  // 缴费到期提醒
	ReminderTypeStop       = "stop_reminder" // 停缴提醒（离职触发）
	ReminderTypeBaseAdjust = "base_adjust"  // 基数调整建议
)

// Reminder 社保提醒记录
type Reminder struct {
	model.BaseModel
	Type        string     `gorm:"column:type;type:varchar(30);not null;index" json:"type"`
	Title       string     `gorm:"column:title;type:varchar(200);not null" json:"title"`
	EmployeeID  int64      `gorm:"column:employee_id;index" json:"employee_id"`
	RecordID    int64      `gorm:"column:record_id;index" json:"record_id"`
	DueDate     *time.Time `gorm:"column:due_date;type:date" json:"due_date"`
	IsRead      bool       `gorm:"column:is_read;default:false" json:"is_read"`
	IsDismissed bool       `gorm:"column:is_dismissed;default:false" json:"is_dismissed"`
}

// TableName 指定表名
func (Reminder) TableName() string {
	return "social_insurance_reminders"
}
