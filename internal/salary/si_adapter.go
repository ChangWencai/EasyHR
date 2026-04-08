package salary

import (
	"fmt"

	"github.com/wencai/easyhr/internal/socialinsurance"
)

// SIAdapter 社保扣款适配器
type SIAdapter struct {
	siSvc *socialinsurance.Service
}

// NewSIAdapter 创建社保扣款适配器
func NewSIAdapter(siSvc *socialinsurance.Service) *SIAdapter {
	return &SIAdapter{siSvc: siSvc}
}

// GetPersonalDeduction 获取社保个人扣款总额
func (a *SIAdapter) GetPersonalDeduction(orgID, employeeID int64, month string) (float64, error) {
	deduction, err := a.siSvc.GetSocialInsuranceDeduction(orgID, employeeID, month)
	if err != nil {
		return 0, fmt.Errorf("社保扣款查询失败: %w", err)
	}
	return deduction.TotalPersonal, nil
}

// SuggestBaseAdjustment 社保基数调整建议
func (a *SIAdapter) SuggestBaseAdjustment(orgID, employeeID int64, newSalary float64) {
	a.siSvc.SuggestBaseAdjustment(orgID, employeeID, newSalary)
}
