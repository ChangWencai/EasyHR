package tax

import (
	"fmt"

	"github.com/wencai/easyhr/internal/employee"
)

// EmployeeAdapter 员工信息适配器
// 实现 EmployeeInfoProvider 接口，解耦个税模块和员工模块
type EmployeeAdapter struct {
	contractRepo *employee.ContractRepository
	empRepo      *employee.Repository
}

// NewEmployeeAdapter 创建员工信息适配器
func NewEmployeeAdapter(contractRepo *employee.ContractRepository, empRepo *employee.Repository) *EmployeeAdapter {
	return &EmployeeAdapter{
		contractRepo: contractRepo,
		empRepo:      empRepo,
	}
}

// GetActiveSalary 获取员工当前有效合同的薪资
func (a *EmployeeAdapter) GetActiveSalary(orgID, employeeID int64) (float64, error) {
	// 查询该员工的 active 合同列表
	contracts, _, err := a.contractRepo.ListByEmployee(orgID, employeeID, 1, 10)
	if err != nil {
		return 0, fmt.Errorf("adapter: query employee contracts: %w", err)
	}

	// 找到 active 状态的合同
	for _, c := range contracts {
		if c.Status == employee.ContractStatusActive {
			return c.Salary, nil
		}
	}

	return 0, fmt.Errorf("adapter: employee %d has no active contract", employeeID)
}

// GetEmployeeHireMonth 获取员工入职月份，格式 "2025-03"
func (a *EmployeeAdapter) GetEmployeeHireMonth(orgID, employeeID int64) (string, error) {
	emp, err := a.empRepo.FindByID(orgID, employeeID)
	if err != nil {
		return "", fmt.Errorf("adapter: find employee: %w", err)
	}
	return emp.HireDate.Format("2006-01"), nil
}
