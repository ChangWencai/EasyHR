package tax

import (
	"github.com/wencai/easyhr/internal/socialinsurance"
)

// SocialInsuranceAdapter 社保扣款适配器
// 实现 SIDeductionProvider 接口，解耦个税模块和社保模块
type SocialInsuranceAdapter struct {
	siSvc *socialinsurance.Service
}

// NewSocialInsuranceAdapter 创建社保扣款适配器
func NewSocialInsuranceAdapter(siSvc *socialinsurance.Service) *SocialInsuranceAdapter {
	return &SocialInsuranceAdapter{siSvc: siSvc}
}

// GetPersonalDeduction 获取员工当月社保个人扣款总额
func (a *SocialInsuranceAdapter) GetPersonalDeduction(orgID, employeeID int64, month string) (float64, error) {
	deduction, err := a.siSvc.GetSocialInsuranceDeduction(orgID, employeeID, month)
	if err != nil {
		return 0, err
	}
	return deduction.TotalPersonal, nil
}
