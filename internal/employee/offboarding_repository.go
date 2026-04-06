package employee

import (
	"fmt"

	"github.com/wencai/easyhr/internal/common/middleware"
	"gorm.io/gorm"
)

// OffboardingRepository 离职管理数据访问层
type OffboardingRepository struct {
	db *gorm.DB
}

// NewOffboardingRepository 创建离职 Repository
func NewOffboardingRepository(db *gorm.DB) *OffboardingRepository {
	return &OffboardingRepository{db: db}
}

// Create 创建离职记录
func (r *OffboardingRepository) Create(ob *Offboarding) error {
	return r.db.Create(ob).Error
}

// FindByID 根据 ID 查找离职记录（带租户隔离）
func (r *OffboardingRepository) FindByID(orgID, id int64) (*Offboarding, error) {
	var ob Offboarding
	err := r.db.Scopes(middleware.TenantScope(orgID)).Where("id = ?", id).First(&ob).Error
	if err != nil {
		return nil, err
	}
	return &ob, nil
}

// FindByEmployeeID 根据员工 ID 查找离职记录（带租户隔离）
func (r *OffboardingRepository) FindByEmployeeID(orgID, employeeID int64) (*Offboarding, error) {
	var ob Offboarding
	err := r.db.Scopes(middleware.TenantScope(orgID)).
		Where("employee_id = ?", employeeID).
		Order("created_at DESC").
		First(&ob).Error
	if err != nil {
		return nil, err
	}
	return &ob, nil
}

// Update 更新离职记录（部分更新，带租户隔离）
func (r *OffboardingRepository) Update(orgID, id int64, updates map[string]interface{}) error {
	result := r.db.Model(&Offboarding{}).Scopes(middleware.TenantScope(orgID)).
		Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// List 分页查询离职列表（带租户隔离）
func (r *OffboardingRepository) List(orgID int64, status string, page, pageSize int) ([]Offboarding, int64, error) {
	var offboardings []Offboarding
	var total int64

	q := r.db.Model(&Offboarding{}).Scopes(middleware.TenantScope(orgID))

	if status != "" {
		q = q.Where("status = ?", status)
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count offboardings: %w", err)
	}

	offset := (page - 1) * pageSize
	if err := q.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&offboardings).Error; err != nil {
		return nil, 0, fmt.Errorf("list offboardings: %w", err)
	}

	return offboardings, total, nil
}
