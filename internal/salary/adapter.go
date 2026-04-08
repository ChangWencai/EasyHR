package salary

import (
	"time"

	"github.com/wencai/easyhr/internal/tax"
)

// TaxProvider 个税计算接口
type TaxProvider interface {
	CalculateTax(orgID, employeeID int64, year, month int, grossIncome float64) (*tax.TaxResult, error)
}

// SIDeductionProvider 社保扣款接口
type SIDeductionProvider interface {
	GetPersonalDeduction(orgID, employeeID int64, month string) (float64, error)
}

// EmployeeProvider 员工信息接口
type EmployeeProvider interface {
	GetActiveEmployees(orgID int64) ([]EmployeeInfo, error)
	GetEmployeeByID(orgID, employeeID int64) (*EmployeeInfo, error)
}

// EmployeeInfo 员工简要信息（跨模块传输）
type EmployeeInfo struct {
	ID         int64
	Name       string
	Phone      string
	HireDate   time.Time
	BaseSalary float64 // 从 Contract.Salary 获取
}

// BaseAdjustmentProvider 社保基数调整建议接口
type BaseAdjustmentProvider interface {
	SuggestBaseAdjustment(orgID, employeeID int64, newSalary float64)
}
