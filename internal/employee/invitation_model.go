package employee

import (
	"time"
)

// Invitation 入职邀请模型
// 独立于 BaseModel，使用简化字段（邀请创建后仅更新 status/used_at/employee_id）
type Invitation struct {
	ID         int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	OrgID      int64      `gorm:"column:org_id;index;not null" json:"org_id"`
	Token      string     `gorm:"column:token;type:varchar(64);uniqueIndex;not null" json:"token"`
	Position   string     `gorm:"column:position;type:varchar(100)" json:"position"`
	Status     string     `gorm:"column:status;type:varchar(20);not null;default:pending" json:"status"` // pending/used/expired/cancelled
	CreatedBy  int64      `gorm:"column:created_by;not null" json:"created_by"`
	CreatedAt  time.Time  `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	ExpiresAt  time.Time  `gorm:"column:expires_at;not null" json:"expires_at"`
	UsedAt     *time.Time `gorm:"column:used_at" json:"used_at"`
	EmployeeID *int64     `gorm:"column:employee_id" json:"employee_id"`
}

// TableName 指定表名
func (Invitation) TableName() string {
	return "invitations"
}

// InvitationStatus 邀请状态常量
const (
	InvitationStatusPending   = "pending"   // 待使用
	InvitationStatusUsed      = "used"      // 已使用
	InvitationStatusExpired   = "expired"   // 已过期
	InvitationStatusCancelled = "cancelled" // 已取消
)

// InvitationExpiryDuration 邀请有效期：7天
const InvitationExpiryDuration = 7 * 24 * time.Hour
