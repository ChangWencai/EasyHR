package socialinsurance

import (
	"fmt"
	"math"

	"github.com/wencai/easyhr/internal/city"
)

// Service 社保政策业务逻辑层
type Service struct {
	repo *Repository
}

// NewService 创建社保政策 Service
func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// CreatePolicy 创建社保政策
func (s *Service) CreatePolicy(policy *SocialInsurancePolicy) error {
	return s.repo.Create(policy)
}

// GetPolicy 获取政策详情
func (s *Service) GetPolicy(id int64) (*PolicyResponse, error) {
	policy, err := s.repo.FindByID(id)
	if err != nil {
		return nil, ErrPolicyNotFound
	}
	return s.toPolicyResponse(policy), nil
}

// ListPolicies 政策列表（关联城市名称）
func (s *Service) ListPolicies(cityID int, page, pageSize int) ([]PolicyResponse, int64, error) {
	policies, total, err := s.repo.List(cityID, page, pageSize)
	if err != nil {
		return nil, 0, err
	}

	responses := make([]PolicyResponse, 0, len(policies))
	for i := range policies {
		responses = append(responses, *s.toPolicyResponse(&policies[i]))
	}
	return responses, total, nil
}

// UpdatePolicy 更新政策
func (s *Service) UpdatePolicy(id int64, req *UpdatePolicyRequest) error {
	updates := map[string]interface{}{
		"config": newJSONType(req.Config),
	}
	return s.repo.Update(id, updates)
}

// DeletePolicy 删除政策
func (s *Service) DeletePolicy(id int64) error {
	return s.repo.Delete(id)
}

// CalculateInsuranceAmounts 根据城市+薪资计算各险种缴费金额
func (s *Service) CalculateInsuranceAmounts(cityID int, salary float64, year int) (*CalculateResponse, error) {
	// 1. 查询适用政策
	policy, err := s.repo.FindByCityAndYear(cityID, year)
	if err != nil {
		return nil, fmt.Errorf("未找到该城市社保政策: %w", err)
	}

	// 2. 获取城市名称
	cityName := getCityName(cityID)

	// 3. 计算各险种
	config := policy.Config.Data()

	items := []struct {
		name string
		item InsuranceItem
	}{
		{"养老保险", config.Pension},
		{"医疗保险", config.Medical},
		{"失业保险", config.Unemployment},
		{"工伤保险", config.WorkInjury},
		{"生育保险", config.Maternity},
		{"住房公积金", config.HousingFund},
	}

	var details []InsuranceAmountDetail
	var totalCompany, totalPersonal float64
	var baseAmount float64

	for _, it := range items {
		// 基数 clamp 到上下限范围
		base := clamp(salary, it.item.BaseLower, it.item.BaseUpper)
		companyAmount := base * it.item.CompanyRate
		personalAmount := base * it.item.PersonalRate

		details = append(details, InsuranceAmountDetail{
			Name:           it.name,
			Base:           base,
			CompanyRate:    it.item.CompanyRate,
			CompanyAmount:  math.Round(companyAmount*100) / 100,
			PersonalRate:   it.item.PersonalRate,
			PersonalAmount: math.Round(personalAmount*100) / 100,
		})

		totalCompany += math.Round(companyAmount*100) / 100
		totalPersonal += math.Round(personalAmount*100) / 100
		baseAmount = base
	}

	return &CalculateResponse{
		CityName:      cityName,
		Salary:        salary,
		BaseAmount:    baseAmount,
		TotalCompany:  math.Round(totalCompany*100) / 100,
		TotalPersonal: math.Round(totalPersonal*100) / 100,
		Items:         details,
	}, nil
}

// GetSocialInsuranceDeduction 预留接口（D-12），Plan 02 实现
func (s *Service) GetSocialInsuranceDeduction(orgID, employeeID int64, month string) (interface{}, error) {
	// TODO: Plan 02 实现
	return nil, fmt.Errorf("not implemented")
}

// toPolicyResponse 转换为响应 DTO
func (s *Service) toPolicyResponse(policy *SocialInsurancePolicy) *PolicyResponse {
	return &PolicyResponse{
		ID:            policy.ID,
		CityID:        policy.CityID,
		CityName:      getCityName(policy.CityID),
		EffectiveYear: policy.EffectiveYear,
		Config:        policy.Config.Data(),
		CreatedAt:     policy.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:     policy.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

// clamp 将值限制在 [lower, upper] 范围内
func clamp(value, lower, upper float64) float64 {
	if value < lower {
		return lower
	}
	if value > upper {
		return upper
	}
	return value
}

// getCityName 根据 cityID 获取城市名称
func getCityName(cityID int) string {
	for _, c := range city.Cities {
		if c.ID == cityID {
			return c.Name
		}
	}
	return "未知城市"
}
