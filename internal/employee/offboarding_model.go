package employee

import (
	"time"

	"github.com/wencai/easyhr/internal/common/model"
	"gorm.io/datatypes"
)

// Offboarding 离职管理模型
// 记录员工离职流程：申请/审批/交接/完成
// ChecklistItems 使用 JSONB 存储模板化交接清单
type Offboarding struct {
	model.BaseModel
	EmployeeID      int64          `gorm:"column:employee_id;index;not null" json:"employee_id"`
	Type            string         `gorm:"column:type;type:varchar(20);not null" json:"type"`               // voluntary/involuntary
	ResignationDate time.Time      `gorm:"column:resignation_date;type:date;not null" json:"resignation_date"`
	Reason          string         `gorm:"column:reason;type:varchar(500)" json:"reason"`
	Status          string         `gorm:"column:status;type:varchar(20);not null;default:pending" json:"status"` // pending/approved/completed
	ChecklistItems  datatypes.JSON `gorm:"column:checklist_items;type:jsonb;not null" json:"checklist_items"`
	CompletedAt     *time.Time     `gorm:"column:completed_at" json:"completed_at"`
	ApprovedBy      *int64         `gorm:"column:approved_by" json:"approved_by"`
	ApprovedAt      *time.Time     `gorm:"column:approved_at" json:"approved_at"`
}

// TableName 指定表名
func (Offboarding) TableName() string {
	return "offboardings"
}

// ChecklistCategory 交接清单分类
type ChecklistCategory struct {
	Category string         `json:"category"`
	Items    []ChecklistItem `json:"items"`
}

// ChecklistItem 交接清单条目
type ChecklistItem struct {
	Name      string `json:"name"`
	Completed bool   `json:"completed"`
}

// Offboarding 状态常量
const (
	OffboardingStatusPending   = "pending"
	OffboardingStatusApproved  = "approved"
	OffboardingStatusCompleted = "completed"
)

// Offboarding 类型常量
const (
	OffboardingTypeVoluntary   = "voluntary"   // 员工主动申请
	OffboardingTypeInvoluntary = "involuntary" // 老板/公司发起
)
