package employee

import (
	"time"
)

// Registration 员工信息登记模型
// 管理员创建登记表，员工通过 Token 链接填写个人信息
type Registration struct {
	ID           int64      `gorm:"primaryKey;autoIncrement;comment:主键ID" json:"id"`
	OrgID        int64      `gorm:"column:org_id;index;not null;comment:所属企业ID" json:"org_id"`
	EmployeeID   *int64     `gorm:"column:employee_id;index;comment:关联员工ID（可选）" json:"employee_id"`
	Name         string     `gorm:"column:name;type:varchar(50);not null;comment:员工姓名" json:"name"`
	DepartmentID *int64     `gorm:"column:department_id;index;comment:所属部门ID" json:"department_id"`
	Position     string     `gorm:"column:position;type:varchar(100);not null;comment:岗位" json:"position"`
	HireDate     *time.Time `gorm:"column:hire_date;type:date;comment:入职日期" json:"hire_date"`
	Token        string     `gorm:"column:token;type:varchar(64);uniqueIndex;not null;comment:登记Token" json:"token"`
	Status       string     `gorm:"column:status;type:varchar(20);not null;default:pending;comment:状态（pending/used/expired）" json:"status"`
	ExpiresAt    time.Time  `gorm:"column:expires_at;not null;comment:过期时间" json:"expires_at"`
	UsedAt       *time.Time `gorm:"column:used_at;comment:使用时间" json:"used_at"`
	CreatedBy    int64      `gorm:"column:created_by;not null;comment:创建人ID" json:"created_by"`
	CreatedAt    time.Time  `gorm:"column:created_at;autoCreateTime;comment:创建时间" json:"created_at"`
}

// TableName 指定表名
func (Registration) TableName() string {
	return "registrations"
}

// Registration 状态常量
const (
	RegistrationStatusPending = "pending" // 待填写
	RegistrationStatusUsed    = "used"    // 已提交
	RegistrationStatusExpired = "expired" // 已过期
)

// RegistrationExpiryDuration 登记有效期：7天
const RegistrationExpiryDuration = 7 * 24 * time.Hour
