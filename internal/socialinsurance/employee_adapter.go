package socialinsurance

import (
	"fmt"

	"github.com/wencai/easyhr/internal/employee"
)

// EmployeeAdapter 员工查询适配器
// 将 employee.Repository 的接口适配为 socialinsurance.EmployeeQuerier
// 解耦社保模块和员工模块，社保不直接 import employee 包
type EmployeeAdapter struct {
	empRepo *employee.Repository
}

// NewEmployeeAdapter 创建员工查询适配器
func NewEmployeeAdapter(empRepo *employee.Repository) *EmployeeAdapter {
	return &EmployeeAdapter{empRepo: empRepo}
}

// FindEmployeeByIDs 批量查找员工信息
func (a *EmployeeAdapter) FindEmployeeByIDs(orgID int64, ids []int64) ([]EmployeeInfo, error) {
	emps, err := a.empRepo.FindByIDs(orgID, ids)
	if err != nil {
		return nil, fmt.Errorf("adapter: find employees by IDs: %w", err)
	}

	result := make([]EmployeeInfo, 0, len(emps))
	for _, emp := range emps {
		result = append(result, EmployeeInfo{
			ID:     emp.ID,
			Name:   emp.Name,
			OrgID:  emp.OrgID,
			UserID: emp.UserID,
		})
	}
	return result, nil
}

// FindEmployeeByUserID 通过 user_id 查找员工信息
func (a *EmployeeAdapter) FindEmployeeByUserID(orgID int64, userID int64) (*EmployeeInfo, error) {
	emp, err := a.empRepo.FindByUserID(orgID, userID)
	if err != nil {
		return nil, fmt.Errorf("adapter: find employee by user ID: %w", err)
	}

	return &EmployeeInfo{
		ID:     emp.ID,
		Name:   emp.Name,
		OrgID:  emp.OrgID,
		UserID: emp.UserID,
	}, nil
}
