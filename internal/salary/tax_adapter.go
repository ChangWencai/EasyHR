package salary

import (
	"fmt"

	"github.com/wencai/easyhr/internal/tax"
)

// TaxAdapter 个税计算适配器
type TaxAdapter struct {
	taxSvc TaxCalculatorService
}

// TaxCalculatorService 个税服务接口（用于解耦）
type TaxCalculatorService interface {
	CalculateTax(orgID, employeeID int64, year, month int, grossIncome float64) (*tax.TaxResult, error)
}

// NewTaxAdapter 创建个税计算适配器
func NewTaxAdapter(taxSvc TaxCalculatorService) *TaxAdapter {
	return &TaxAdapter{taxSvc: taxSvc}
}

// CalculateTax 计算个税
func (a *TaxAdapter) CalculateTax(orgID, employeeID int64, year, month int, grossIncome float64) (*tax.TaxResult, error) {
	result, err := a.taxSvc.CalculateTax(orgID, employeeID, year, month, grossIncome)
	if err != nil {
		return nil, fmt.Errorf("个税计算失败: %w", err)
	}
	return result, nil
}
