package todo

import (
	"time"

	"github.com/wencai/easyhr/internal/common/model"
)

// TodoItem 待办事项模型（完整字段，对应 todo_items 表）
type TodoItem struct {
	model.BaseModel
	Title         string     `gorm:"column:title;type:varchar(500);not null;comment:事项标题" json:"title"`
	Type          string     `gorm:"column:type;type:varchar(50);not null;comment:事项类型" json:"type"`
	Content       string     `gorm:"column:content;type:text;comment:事项内容" json:"content"`
	EmployeeID    *int64     `gorm:"column:employee_id;index;comment:关联员工ID" json:"employee_id"`
	EmployeeName  string     `gorm:"column:employee_name;type:varchar(100);comment:员工姓名" json:"employee_name"`
	CreatorName   string     `gorm:"column:creator_name;type:varchar(100);comment:创建人姓名" json:"creator_name"`
	Deadline      *time.Time `gorm:"column:deadline;type:date;comment:截止日期" json:"deadline"`
	IsTimeLimited bool       `gorm:"column:is_time_limited;default:false;comment:是否限时任务" json:"is_time_limited"`
	UrgencyStatus string     `gorm:"column:urgency_status;type:varchar(20);default:'normal';comment:紧迫状态(normal/overdue/expired)" json:"urgency_status"`
	Status        string     `gorm:"column:status;type:varchar(20);not null;default:'pending';comment:状态(pending/completed/terminated)" json:"status"`
	SourceType    string     `gorm:"column:source_type;type:varchar(50);comment:来源类型(contract/tax/si/salary/attendance)" json:"source_type"`
	SourceID      *int64     `gorm:"column:source_id;comment:来源记录ID" json:"source_id"`
	IsPinned      bool       `gorm:"column:is_pinned;default:false;comment:是否置顶" json:"is_pinned"`
	SortOrder     int        `gorm:"column:sort_order;default:0;comment:排序值(越大越靠前)" json:"sort_order"`
}

// TableName 指定表名
func (TodoItem) TableName() string { return "todo_items" }

// TodoItemStatus 常量
const (
	TodoStatusPending    = "pending"
	TodoStatusCompleted  = "completed"
	TodoStatusTerminated = "terminated"
)

// UrgencyStatus 常量
const (
	UrgencyNormal  = "normal"
	UrgencyOverdue = "overdue"
	UrgencyExpired = "expired"
)

// TodoType 常量（7种限时任务 + 通用类型）
const (
	TodoTypeContractNew    = "contract_new"
	TodoTypeContractRenew  = "contract_renew"
	TodoTypeTaxDeclaration = "tax_declaration"
	TodoTypeSIPayment      = "si_payment"
	TodoTypeSIChange       = "si_change"
	TodoTypeSIAnnualBase   = "si_annual_base"
	TodoTypeFundAnnualBase = "fund_annual_base"
)

// CarouselItem 轮播图配置模型
type CarouselItem struct {
	model.BaseModel
	ImageURL  string     `gorm:"column:image_url;type:varchar(500);not null;comment:图片OSS地址" json:"image_url"`
	LinkURL   string     `gorm:"column:link_url;type:varchar(500);comment:跳转链接" json:"link_url"`
	SortOrder int        `gorm:"column:sort_order;default:0;comment:排序(越大越靠前)" json:"sort_order"`
	Active    bool       `gorm:"column:active;default:true;comment:是否启用" json:"active"`
	StartAt   *time.Time `gorm:"column:start_at;comment:生效开始时间" json:"start_at"`
	EndAt     *time.Time `gorm:"column:end_at;comment:生效结束时间" json:"end_at"`
}

// TableName 指定表名
func (CarouselItem) TableName() string { return "carousel_items" }

// TodoInvite 协办邀请模型（Token独立验证，无需登录）
type TodoInvite struct {
	model.BaseModel
	TodoID    int64      `gorm:"column:todo_id;index;not null" json:"todo_id"`
	Token     string     `gorm:"column:token;type:varchar(64);uniqueIndex;not null" json:"token"`
	Status    string     `gorm:"column:status;type:varchar(20);not null;default:'pending'" json:"status"`
	ExpiresAt time.Time  `gorm:"column:expires_at;not null" json:"expires_at"`
	UsedAt    *time.Time `gorm:"column:used_at" json:"used_at"`
}

// TableName 指定表名
func (TodoInvite) TableName() string { return "todo_invites" }

// TodoInviteStatus 常量
const (
	InviteStatusPending  = "pending"
	InviteStatusUsed     = "used"
	InviteStatusExpired  = "expired"
)

// InviteExpiryDuration 邀请有效期：7天
const InviteExpiryDuration = 7 * 24 * time.Hour
