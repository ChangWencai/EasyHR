package audit

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type AuditLog struct {
	ID         int64          `gorm:"primaryKey;autoIncrement" json:"id"`
	OrgID      int64          `gorm:"column:org_id;index;not null" json:"org_id"`
	UserID     int64          `gorm:"column:user_id;index;not null" json:"user_id"`
	Module     string         `gorm:"column:module;type:varchar(50);not null" json:"module"`
	Action     string         `gorm:"column:action;type:varchar(20);not null" json:"action"`
	TargetType string         `gorm:"column:target_type;type:varchar(50);not null" json:"target_type"`
	TargetID   int64          `gorm:"column:target_id;not null" json:"target_id"`
	Detail     datatypes.JSON `gorm:"column:detail;type:jsonb" json:"detail"`
	IPAddress  string         `gorm:"column:ip_address;type:varchar(45)" json:"ip_address"`
	CreatedAt  time.Time      `gorm:"column:created_at;autoCreateTime" json:"created_at"`
}

func (AuditLog) TableName() string {
	return "audit_logs"
}

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) List(orgID int64, module, action string, page, pageSize int) ([]AuditLog, int64, error) {
	var logs []AuditLog
	var total int64
	q := r.db.Model(&AuditLog{}).Where("org_id = ?", orgID)
	if module != "" {
		q = q.Where("module = ?", module)
	}
	if action != "" {
		q = q.Where("action = ?", action)
	}
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * pageSize
	if err := q.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&logs).Error; err != nil {
		return nil, 0, err
	}
	return logs, total, nil
}
