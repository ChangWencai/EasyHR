package socialinsurance

import (
	"fmt"

	"github.com/wencai/easyhr/internal/common/middleware"
	"gorm.io/gorm"
)

// ReminderRepository 提醒记录数据访问层
type ReminderRepository struct {
	db *gorm.DB
}

// NewReminderRepository 创建提醒 Repository
func NewReminderRepository(db *gorm.DB) *ReminderRepository {
	return &ReminderRepository{db: db}
}

// Create 创建提醒记录
func (r *ReminderRepository) Create(reminder *Reminder) error {
	return r.db.Create(reminder).Error
}

// FindByTypeAndRecordID 按类型和记录ID查找提醒（去重检查）
func (r *ReminderRepository) FindByTypeAndRecordID(orgID int64, reminderType string, recordID int64) (*Reminder, error) {
	var reminder Reminder
	err := r.db.Scopes(middleware.TenantScope(orgID)).
		Where("type = ? AND record_id = ? AND is_dismissed = false", reminderType, recordID).
		First(&reminder).Error
	if err != nil {
		return nil, err
	}
	return &reminder, nil
}

// ListUnread 查询未读提醒列表（分页）
func (r *ReminderRepository) ListUnread(orgID int64, reminderType string, page, pageSize int) ([]Reminder, int64, error) {
	var reminders []Reminder
	var total int64

	q := r.db.Model(&Reminder{}).Scopes(middleware.TenantScope(orgID)).
		Where("is_dismissed = false")

	if reminderType != "" {
		q = q.Where("type = ?", reminderType)
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count reminders: %w", err)
	}

	offset := (page - 1) * pageSize
	if err := q.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&reminders).Error; err != nil {
		return nil, 0, fmt.Errorf("list reminders: %w", err)
	}

	return reminders, total, nil
}

// MarkAsRead 标记为已读
func (r *ReminderRepository) MarkAsRead(orgID, id int64) error {
	result := r.db.Model(&Reminder{}).Scopes(middleware.TenantScope(orgID)).
		Where("id = ?", id).
		Update("is_read", true)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("提醒记录不存在")
	}
	return nil
}

// Dismiss 关闭提醒
func (r *ReminderRepository) Dismiss(orgID, id int64) error {
	result := r.db.Model(&Reminder{}).Scopes(middleware.TenantScope(orgID)).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"is_dismissed": true,
			"is_read":      true,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("提醒记录不存在")
	}
	return nil
}

// FindActiveRecordsGroupedByOrg 查询所有企业的 active 参保记录（用于定时任务扫描）
func (r *ReminderRepository) FindActiveRecordsGroupedByOrg() ([]SocialInsuranceRecord, error) {
	var records []SocialInsuranceRecord
	err := r.db.Where("status = ?", SIStatusActive).Find(&records).Error
	return records, err
}
