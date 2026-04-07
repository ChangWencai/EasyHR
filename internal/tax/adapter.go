package tax

// EmployeeInfoProvider 员工信息接口（由 employee adapter 在 main.go 实现, per D-09）
type EmployeeInfoProvider interface {
	// GetActiveSalary 获取员工当前有效合同的薪资
	GetActiveSalary(orgID, employeeID int64) (float64, error)
	// GetEmployeeHireMonth 获取员工入职月份，格式 "2025-03"
	GetEmployeeHireMonth(orgID, employeeID int64) (string, error)
}

// SIDeductionProvider 社保扣款接口（由 socialinsurance adapter 在 main.go 实现, per D-10）
type SIDeductionProvider interface {
	// GetPersonalDeduction 获取员工当月社保个人扣款总额
	GetPersonalDeduction(orgID, employeeID int64, month string) (float64, error)
}
