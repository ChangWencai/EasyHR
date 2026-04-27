package salary

import "fmt"

// PerfCreatorAdapter 绩效系数创建适配器
// 将 employee.PerfCreator 接口实现为 salary.PerformanceService 的初始化逻辑
// 解耦 salary 模块和员工模块，员工模块不直接 import salary 包
type PerfCreatorAdapter struct {
	perfSvc *PerformanceService
}

// NewPerfCreatorAdapter 创建绩效系数创建适配器
func NewPerfCreatorAdapter(perfSvc *PerformanceService) *PerfCreatorAdapter {
	return &PerfCreatorAdapter{perfSvc: perfSvc}
}

// InitEmployeePerf 初始化员工绩效系数
func (a *PerfCreatorAdapter) InitEmployeePerf(orgID, userID, empID int64, year, month int, coefficient float64) error {
	// 系数范围 [0, 1]，默认 1.0
	if coefficient < 0 {
		coefficient = 0
	}
	if coefficient > 1 {
		coefficient = 1
	}
	coefficients := []CoefficientInput{{EmployeeID: empID, Coefficient: coefficient}}
	if err := a.perfSvc.SetCoefficient(orgID, userID, year, month, coefficients); err != nil {
		return fmt.Errorf("初始化绩效系数失败: %w", err)
	}
	return nil
}
