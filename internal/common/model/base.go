package model

import (
	"time"

	"gorm.io/gorm"
)

type BaseModel struct {
	ID        int64          `gorm:"primaryKey;autoIncrement" json:"id"`
	OrgID     int64          `gorm:"column:org_id;index;not null" json:"org_id"`
	CreatedBy int64          `gorm:"column:created_by" json:"created_by"`
	CreatedAt time.Time      `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedBy int64          `gorm:"column:updated_by" json:"updated_by"`
	UpdatedAt time.Time      `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index" json:"-"`
}
