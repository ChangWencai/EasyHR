package salary

import (
	"fmt"

	"gorm.io/gorm"
)

// PerformanceRepository 绩效系数数据访问层
type PerformanceRepository struct {
	db *gorm.DB
}

// NewPerformanceRepository 创建绩效系数 Repository
func NewPerformanceRepository(db *gorm.DB) *PerformanceRepository {
	return &PerformanceRepository{db: db}
}

// Upsert 创建或更新绩效系数（UNIQUE INDEX 防止冲突）
func (r *PerformanceRepository) Upsert(orgID, employeeID, userID int64, year, month int, coefficient float64) error {
	var existing PerformanceCoefficient
	err := r.db.Where("org_id = ? AND employee_id = ? AND year = ? AND month = ?",
		orgID, employeeID, year, month).First(&existing).Error

	if err == gorm.ErrRecordNotFound {
		// INSERT
		record := PerformanceCoefficient{
			OrgID:       orgID,
			EmployeeID:  employeeID,
			Year:        year,
			Month:       month,
			Coefficient: coefficient,
			CreatedBy:   userID,
			UpdatedBy:   userID,
		}
		return r.db.Create(&record).Error
	}
	if err != nil {
		return fmt.Errorf("query performance coefficient: %w", err)
	}

	// UPDATE
	return r.db.Model(&existing).Updates(map[string]interface{}{
		"coefficient": coefficient,
		"updated_by":  userID,
	}).Error
}

// ListByMonth 查询某月所有员工的绩效系数
func (r *PerformanceRepository) ListByMonth(orgID int64, year, month int) ([]PerformanceCoefficient, error) {
	var records []PerformanceCoefficient
	err := r.db.Where("org_id = ? AND year = ? AND month = ?",
		orgID, year, month).Find(&records).Error
	return records, err
}

// FindByEmployee 查询员工某月的绩效系数
func (r *PerformanceRepository) FindByEmployee(orgID, employeeID int64, year, month int) (*PerformanceCoefficient, error) {
	var record PerformanceCoefficient
	err := r.db.Where("org_id = ? AND employee_id = ? AND year = ? AND month = ?",
		orgID, employeeID, year, month).First(&record).Error
	if err != nil {
		return nil, err
	}
	return &record, nil
}
