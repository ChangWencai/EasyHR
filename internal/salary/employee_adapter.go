package salary

import (
	"fmt"

	"github.com/wencai/easyhr/internal/employee"
)

// EmployeeAdapter 员工信息适配器
type EmployeeAdapter struct {
	empRepo      *employee.Repository
	contractRepo *employee.ContractRepository
}

// NewEmployeeAdapter 创建员工信息适配器
func NewEmployeeAdapter(empRepo *employee.Repository, contractRepo *employee.ContractRepository) *EmployeeAdapter {
	return &EmployeeAdapter{
		empRepo:      empRepo,
		contractRepo: contractRepo,
	}
}

// GetActiveEmployees 获取在职员工列表
func (a *EmployeeAdapter) GetActiveEmployees(orgID int64) ([]EmployeeInfo, error) {
	// 查询在职员工
	emps, _, err := a.empRepo.List(orgID, employee.SearchParams{Status: employee.StatusActive}, 1, 1000)
	if err != nil {
		return nil, fmt.Errorf("查询在职员工失败: %w", err)
	}

	result := make([]EmployeeInfo, 0, len(emps))
	for _, emp := range emps {
		info := EmployeeInfo{
			ID:   emp.ID,
			Name: emp.Name,
			// Phone 解密需要 crypto，这里留空，由 Service 层按需处理
			HireDate: emp.HireDate,
		}

		// 获取合同薪资
		contracts, _, err := a.contractRepo.ListByEmployee(orgID, emp.ID, 1, 10)
		if err != nil {
			// 合同查询失败不阻断，使用默认值 0
			result = append(result, info)
			continue
		}
		for _, c := range contracts {
			if c.Status == employee.ContractStatusActive {
				info.BaseSalary = c.Salary
				break
			}
		}

		result = append(result, info)
	}

	return result, nil
}

// GetEmployeeByID 获取单个员工信息
func (a *EmployeeAdapter) GetEmployeeByID(orgID, employeeID int64) (*EmployeeInfo, error) {
	emp, err := a.empRepo.FindByID(orgID, employeeID)
	if err != nil {
		return nil, fmt.Errorf("查询员工失败: %w", err)
	}

	info := &EmployeeInfo{
		ID:       emp.ID,
		Name:     emp.Name,
		HireDate: emp.HireDate,
	}

	// 获取合同薪资
	contracts, _, err := a.contractRepo.ListByEmployee(orgID, emp.ID, 1, 10)
	if err == nil {
		for _, c := range contracts {
			if c.Status == employee.ContractStatusActive {
				info.BaseSalary = c.Salary
				break
			}
		}
	}

	return info, nil
}
