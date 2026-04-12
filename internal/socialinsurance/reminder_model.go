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
	Type        string     `gorm:"column:type;type:varchar(30);not null;index;comment:提醒类型（payment_due/stop_reminder/base_adjust）" json:"type"`
	Title       string     `gorm:"column:title;type:varchar(200);not null;comment:提醒标题" json:"title"`
	EmployeeID  int64      `gorm:"column:employee_id;index;comment:员工ID" json:"employee_id"`
	RecordID    int64      `gorm:"column:record_id;index;comment:关联参保记录ID" json:"record_id"`
	DueDate     *time.Time `gorm:"column:due_date;type:date;comment:到期日期" json:"due_date"`
	IsRead      bool       `gorm:"column:is_read;default:false;comment:是否已读" json:"is_read"`
	IsDismissed bool       `gorm:"column:is_dismissed;default:false;comment:是否已忽略" json:"is_dismissed"`
}

// TableName 指定表名
func (Reminder) TableName() string {
	return "social_insurance_reminders"
}
