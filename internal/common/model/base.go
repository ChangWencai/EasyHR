package model

import (
	"time"

	"gorm.io/gorm"
)

type BaseModel struct {
	ID        int64          `gorm:"primaryKey;autoIncrement;comment:主键ID" json:"id"`
	OrgID     int64          `gorm:"column:org_id;index;not null;comment:所属企业ID，外键到organizations.id" json:"org_id"`
	CreatedBy int64          `gorm:"column:created_by;comment:创建人ID" json:"created_by"`
	CreatedAt time.Time      `gorm:"column:created_at;autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedBy int64          `gorm:"column:updated_by;comment:更新人ID" json:"updated_by"`
	UpdatedAt time.Time      `gorm:"column:updated_at;autoUpdateTime;comment:更新时间" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index;comment:软删除时间戳" json:"-"`
}
