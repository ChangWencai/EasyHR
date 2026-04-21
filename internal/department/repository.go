package department

import (
	"github.com/wencai/easyhr/internal/common/middleware"
	"gorm.io/gorm"
)

// Repository 部门数据访问层
type Repository struct {
	db *gorm.DB
}

// NewRepository 创建部门 Repository
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// Create 创建部门
func (r *Repository) Create(dept *Department) error {
	return r.db.Create(dept).Error
}

// FindByID 根据 ID 查找部门（带租户隔离）
func (r *Repository) FindByID(orgID, id int64) (*Department, error) {
	var dept Department
	err := r.db.Scopes(middleware.TenantScope(orgID)).Where("id = ?", id).First(&dept).Error
	if err != nil {
		return nil, err
	}
	return &dept, nil
}

// Update 更新部门信息（部分更新）
func (r *Repository) Update(orgID, id int64, updates map[string]interface{}) error {
	result := r.db.Model(&Department{}).Scopes(middleware.TenantScope(orgID)).Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// Delete 软删除部门（删除前需检查子部门和员工）
func (r *Repository) Delete(orgID, id int64) error {
	result := r.db.Scopes(middleware.TenantScope(orgID)).Where("id = ?", id).Delete(&Department{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// ListAll 获取全部部门（用于构建树）
func (r *Repository) ListAll(orgID int64) ([]Department, error) {
	var departments []Department
	err := r.db.Scopes(middleware.TenantScope(orgID)).Order("sort_order ASC, id ASC").Find(&departments).Error
	if err != nil {
		return nil, err
	}
	return departments, nil
}

// CountChildren 统计子部门数量
func (r *Repository) CountChildren(orgID, parentID int64) (int64, error) {
	var count int64
	err := r.db.Model(&Department{}).Scopes(middleware.TenantScope(orgID)).
		Where("parent_id = ?", parentID).Count(&count).Error
	return count, err
}

// List 返回部门列表（分页）
func (r *Repository) List(orgID int64, page, pageSize int) ([]Department, int64, error) {
	var departments []Department
	var total int64

	q := r.db.Model(&Department{}).Scopes(middleware.TenantScope(orgID))

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := q.Offset(offset).Limit(pageSize).Order("sort_order ASC, id ASC").Find(&departments).Error; err != nil {
		return nil, 0, err
	}

	return departments, total, nil
}

// FindAllByIDs 批量查询部门（用于目标部门下拉列表）
func (r *Repository) FindAllByIDs(orgID int64, ids []int64) ([]Department, error) {
	var departments []Department
	err := r.db.Scopes(middleware.TenantScope(orgID)).Where("id IN ?", ids).Find(&departments).Error
	return departments, err
}
