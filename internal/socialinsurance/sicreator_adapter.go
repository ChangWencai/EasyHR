package socialinsurance

import (
	"encoding/json"
	"fmt"
	"math"

	"gorm.io/datatypes"
)

// SICreatorAdapter 社保创建适配器
// 将 employee.SICreator 接口实现为 socialinsurance 的参保创建逻辑
// 解耦社保模块和员工模块，员工模块不直接 import socialinsurance 包
type SICreatorAdapter struct {
	repo     *Repository
	cityRepo interface {
		GetNameByCode(cityCode int64) string
	}
}

// NewSICreatorAdapter 创建社保创建适配器
func NewSICreatorAdapter(repo *Repository, cityRepo interface {
	GetNameByCode(cityCode int64) string
}) *SICreatorAdapter {
	return &SICreatorAdapter{repo: repo, cityRepo: cityRepo}
}

// CreateEmployeeSI 创建员工社保参保记录
func (a *SICreatorAdapter) CreateEmployeeSI(orgID, userID, empID int64, empName string, cityCode int64, baseAmount float64, startMonth string, hfBase float64) error {
	// 查询城市政策
	year := 0
	if len(startMonth) >= 4 {
		fmt.Sscanf(startMonth[:4], "%d", &year)
	}
	if year == 0 {
		year = 2026
	}

	policy, err := a.repo.FindByCityAndYear(cityCode, year)
	if err != nil {
		// 无政策时静默跳过，不阻止员工创建
		return nil
	}

	// 确定社保基数
	if baseAmount <= 0 {
		return nil
	}

	// 计算各险种金额
	details := calculateDetailsForCreator(policy.Config.Data(), baseAmount, hfBase)
	detailsJSON, _ := json.Marshal(details)

	var totalCompany, totalPersonal float64
	for _, d := range details {
		totalCompany += d.CompanyAmount
		totalPersonal += d.PersonalAmount
	}

	record := &SocialInsuranceRecord{
		EmployeeID:    empID,
		EmployeeName:  empName,
		CityCode:      cityCode,
		PolicyID:      policy.ID,
		BaseAmount:    details[0].Base,
		Status:        SIStatusActive,
		StartMonth:    startMonth,
		Details:       datatypes.JSON(detailsJSON),
		TotalCompany:  math.Round(totalCompany*100) / 100,
		TotalPersonal: math.Round(totalPersonal*100) / 100,
	}
	record.OrgID = orgID
	record.CreatedBy = userID
	record.UpdatedBy = userID

	if err := a.repo.CreateRecord(record); err != nil {
		return fmt.Errorf("创建参保记录失败: %w", err)
	}

	// 创建变更历史
	history := &ChangeHistory{
		RecordID:   record.ID,
		EmployeeID: empID,
		ChangeType: SIChangeEnroll,
		AfterValue: datatypes.JSON(detailsJSON),
		Remark:     "入职时一并参保",
	}
	history.OrgID = orgID
	history.CreatedBy = userID
	history.UpdatedBy = userID
	_ = a.repo.CreateChangeHistory(history)

	return nil
}

// calculateDetailsForCreator 计算各险种明细（社保创建适配器专用）
func calculateDetailsForCreator(config FiveInsurances, salary, hfBase float64) []InsuranceAmountDetail {
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

	details := make([]InsuranceAmountDetail, 0, len(items))
	for _, it := range items {
		base := salary
		if it.name == "住房公积金" && hfBase > 0 {
			base = hfBase
		}
		base = clampCreator(base, it.item.BaseLower, it.item.BaseUpper)
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
	}
	return details
}

func clampCreator(value, lower, upper float64) float64 {
	if value < lower {
		return lower
	}
	if value > upper {
		return upper
	}
	return value
}
