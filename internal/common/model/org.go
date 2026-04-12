package model

import (
	"time"

	"gorm.io/gorm"
)

// Organization 租户（企业）模型
// 注意：不使用 BaseModel，因为 organizations 是顶级实体
type Organization struct {
	ID           int64          `gorm:"primaryKey;autoIncrement;comment:主键ID" json:"id"`
	Name         string         `gorm:"column:name;type:varchar(200);not null;comment:企业名称" json:"name"`
	CreditCode   string         `gorm:"column:credit_code;type:varchar(18);uniqueIndex:idx_org_credit_code,where:deleted_at IS NULL;not null;comment:统一社会信用代码" json:"credit_code"`
	City         string         `gorm:"column:city;type:varchar(50);not null;comment:所在城市" json:"city"`
	ContactName  string         `gorm:"column:contact_name;type:varchar(50);comment:联系人姓名" json:"contact_name"`
	ContactPhone string         `gorm:"column:contact_phone;type:varchar(200);comment:联系人电话" json:"contact_phone"`
	Status       string         `gorm:"column:status;type:varchar(20);default:active;comment:状态（active/inactive）" json:"status"`
	CreatedBy    int64          `gorm:"column:created_by;comment:创建人ID" json:"created_by"`
	CreatedAt    time.Time      `gorm:"column:created_at;autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedBy    int64          `gorm:"column:updated_by;comment:更新人ID" json:"updated_by"`
	UpdatedAt    time.Time      `gorm:"column:updated_at;autoUpdateTime;comment:更新时间" json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"column:deleted_at;index;comment:软删除时间戳" json:"-"`
}

func (Organization) TableName() string {
	return "organizations"
}
