package employee

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

// RegistrationRepository 员工信息登记数据访问层
type RegistrationRepository struct {
	db *gorm.DB
}

// NewRegistrationRepository 创建登记 Repository
func NewRegistrationRepository(db *gorm.DB) *RegistrationRepository {
	return &RegistrationRepository{db: db}
}

// DB 暴露数据库连接（供 Service 层事务使用）
func (r *RegistrationRepository) DB() *gorm.DB {
	return r.db
}

// Create 创建登记记录
func (r *RegistrationRepository) Create(reg *Registration) error {
	return r.db.Create(reg).Error
}

// FindByToken 根据 Token 查找登记记录（公开接口，不需要 org_id）
func (r *RegistrationRepository) FindByToken(token string) (*Registration, error) {
	var reg Registration
	err := r.db.Where("token = ?", token).First(&reg).Error
	if err != nil {
		return nil, err
	}
	return &reg, nil
}

// FindByID 根据 ID 查找登记记录（管理接口，使用租户隔离）
func (r *RegistrationRepository) FindByID(orgID, id int64) (*Registration, error) {
	var reg Registration
	err := r.db.Where("org_id = ? AND id = ?", orgID, id).First(&reg).Error
	if err != nil {
		return nil, err
	}
	return &reg, nil
}

// UpdateStatus 更新登记状态
func (r *RegistrationRepository) UpdateStatus(token string, status string, usedAt *time.Time) error {
	updates := map[string]interface{}{
		"status": status,
	}
	if usedAt != nil {
		updates["used_at"] = usedAt
	}
	result := r.db.Model(&Registration{}).Where("token = ?", token).Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// List 分页查询登记列表
func (r *RegistrationRepository) List(orgID int64, params RegistrationListQueryParams) ([]Registration, int64, error) {
	var registrations []Registration
	var total int64

	q := r.db.Model(&Registration{}).Where("org_id = ?", orgID)

	if params.Status != "" {
		q = q.Where("status = ?", params.Status)
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count registrations: %w", err)
	}

	page := params.Page
	pageSize := params.PageSize
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize
	if err := q.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&registrations).Error; err != nil {
		return nil, 0, fmt.Errorf("list registrations: %w", err)
	}

	return registrations, total, nil
}

// Delete 删除登记记录
func (r *RegistrationRepository) Delete(orgID, id int64) error {
	result := r.db.Where("org_id = ? AND id = ? AND status = ?", orgID, id, RegistrationStatusPending).Delete(&Registration{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
