package position

import (
	"github.com/wencai/easyhr/internal/common/middleware"
	"gorm.io/gorm"
)

// Repository 岗位数据访问层
type Repository struct {
	db *gorm.DB
}

// NewRepository 创建岗位 Repository
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// Create 创建岗位
func (r *Repository) Create(pos *Position) error {
	return r.db.Create(pos).Error
}

// FindByID 根据 ID 查找岗位（带租户隔离）
func (r *Repository) FindByID(orgID, id int64) (*Position, error) {
	var pos Position
	err := r.db.Scopes(middleware.TenantScope(orgID)).Where("id = ?", id).First(&pos).Error
	if err != nil {
		return nil, err
	}
	return &pos, nil
}

// Update 更新岗位信息（部分更新）
func (r *Repository) Update(orgID, id int64, updates map[string]interface{}) error {
	result := r.db.Model(&Position{}).Scopes(middleware.TenantScope(orgID)).Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// Delete 软删除岗位
func (r *Repository) Delete(orgID, id int64) error {
	result := r.db.Scopes(middleware.TenantScope(orgID)).Where("id = ?", id).Delete(&Position{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// ListByOrg 获取全部岗位（按排序字段升序）
func (r *Repository) ListByOrg(orgID int64) ([]Position, error) {
	var positions []Position
	err := r.db.Scopes(middleware.TenantScope(orgID)).Order("sort_order ASC, id ASC").Find(&positions).Error
	return positions, err
}

// ExistsByNameAndDept 检查同部门同名岗位是否已存在
// NULL department_id 使用 IS NULL 显式比较（PostgreSQL NULL 语义）
func (r *Repository) ExistsByNameAndDept(orgID int64, deptID *int64, name string) (bool, error) {
	var count int64
	query := r.db.Model(&Position{}).Where("org_id = ? AND name = ?", orgID, name)
	if deptID == nil {
		query = query.Where("department_id IS NULL")
	} else {
		query = query.Where("department_id = ?", *deptID)
	}
	if err := query.Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// ListByDepartment 获取指定部门的岗位
func (r *Repository) ListByDepartment(orgID, deptID int64) ([]Position, error) {
	var positions []Position
	err := r.db.Scopes(middleware.TenantScope(orgID)).
		Where("department_id = ?", deptID).
		Order("sort_order ASC, id ASC").
		Find(&positions).Error
	return positions, err
}

// CountByPositionID 统计使用指定岗位的员工数量（用于删除前校验）
func (r *Repository) CountByPositionID(orgID, positionID int64) (int64, error) {
	var count int64
	err := r.db.Table("employees").
		Where("org_id = ? AND position_id = ?", orgID, positionID).
		Count(&count).Error
	return count, err
}
