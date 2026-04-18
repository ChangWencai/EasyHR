package salary

import (
	"fmt"

	"github.com/wencai/easyhr/internal/common/middleware"
	"gorm.io/gorm"
)

// AdjustmentRepository 调薪数据访问层
type AdjustmentRepository struct {
	db *gorm.DB
}

// NewAdjustmentRepository 创建调薪 Repository
func NewAdjustmentRepository(db *gorm.DB) *AdjustmentRepository {
	return &AdjustmentRepository{db: db}
}

// Create 插入调薪记录（INSERT ONLY）
func (r *AdjustmentRepository) Create(adj *SalaryAdjustment) error {
	return r.db.Create(adj).Error
}

// List 分页查询调薪记录
func (r *AdjustmentRepository) List(orgID int64, effectiveMonth string, page, pageSize int) ([]SalaryAdjustment, int64, error) {
	var records []SalaryAdjustment
	var total int64

	q := r.db.Model(&SalaryAdjustment{}).
		Where("org_id = ?", orgID)

	if effectiveMonth != "" {
		q = q.Where("effective_month = ?", effectiveMonth)
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count adjustments: %w", err)
	}

	offset := (page - 1) * pageSize
	if err := q.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&records).Error; err != nil {
		return nil, 0, fmt.Errorf("list adjustments: %w", err)
	}

	return records, total, nil
}

// FindByEmployeeMonth 查询员工某月的调薪记录
func (r *AdjustmentRepository) FindByEmployeeMonth(orgID, employeeID int64, effectiveMonth string) ([]SalaryAdjustment, error) {
	var records []SalaryAdjustment
	err := r.db.Where("org_id = ? AND employee_id = ? AND effective_month = ?",
		orgID, employeeID, effectiveMonth).
		Find(&records).Error
	return records, err
}

// CountEmployeesByDepartments 统计指定部门下的员工数
func (r *AdjustmentRepository) CountEmployeesByDepartments(orgID int64, departmentIDs []int64) (int64, error) {
	var count int64
	err := r.db.Table("employees").
		Scopes(middleware.TenantScope(orgID)).
		Where("department_id IN ?", departmentIDs).
		Where("status = ?", "active").
		Count(&count).Error
	return count, err
}

// GetEmployeeIDsByDepartments 获取指定部门下的活跃员工 ID 列表
func (r *AdjustmentRepository) GetEmployeeIDsByDepartments(orgID int64, departmentIDs []int64) ([]int64, error) {
	var ids []int64
	err := r.db.Table("employees").
		Scopes(middleware.TenantScope(orgID)).
		Where("department_id IN ?", departmentIDs).
		Where("status = ?", "active").
		Pluck("id", &ids).Error
	return ids, err
}
