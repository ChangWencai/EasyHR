package socialinsurance

import (
	"encoding/json"
	"fmt"
	"math"
	"time"

	"github.com/wencai/easyhr/internal/city"
	"gorm.io/datatypes"
)

// Service 社保业务逻辑层
type Service struct {
	repo        *Repository
	empQuerier  EmployeeQuerier
}

// NewService 创建社保 Service
func NewService(repo *Repository, empQuerier EmployeeQuerier) *Service {
	return &Service{repo: repo, empQuerier: empQuerier}
}

// ========== 政策管理 ==========

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
	details := calculateDetails(policy.Config.Data(), salary)

	var totalCompany, totalPersonal, baseAmount float64
	for _, d := range details {
		totalCompany += d.CompanyAmount
		totalPersonal += d.PersonalAmount
		baseAmount = d.Base
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

// ========== 参保操作 ==========

// EnrollPreview 参保预览
func (s *Service) EnrollPreview(orgID int64, req *EnrollPreviewRequest) (*EnrollPreviewResponse, error) {
	// 查询员工列表
	employees, err := s.empQuerier.FindEmployeeByIDs(orgID, req.EmployeeIDs)
	if err != nil {
		return nil, fmt.Errorf("查询员工信息失败: %w", err)
	}

	currentYear := time.Now().Year()
	items := make([]EnrollPreviewItem, 0, len(employees))

	for _, emp := range employees {
		calcResult, err := s.CalculateInsuranceAmounts(req.CityID, 0, currentYear)
		if err != nil {
			// 无政策时跳过此员工
			continue
		}

		items = append(items, EnrollPreviewItem{
			EmployeeID:    emp.ID,
			EmployeeName:  emp.Name,
			BaseAmount:    calcResult.BaseAmount,
			TotalCompany:  calcResult.TotalCompany,
			TotalPersonal: calcResult.TotalPersonal,
			Items:         calcResult.Items,
		})
	}

	return &EnrollPreviewResponse{Items: items}, nil
}

// BatchEnroll 批量参保
func (s *Service) BatchEnroll(orgID, userID int64, req *BatchEnrollRequest) (*BatchEnrollResult, error) {
	// 查询员工列表
	employees, err := s.empQuerier.FindEmployeeByIDs(orgID, req.EmployeeIDs)
	if err != nil {
		return nil, fmt.Errorf("查询员工信息失败: %w", err)
	}

	// 构建员工ID到员工信息的映射
	empMap := make(map[int64]EmployeeInfo, len(employees))
	for _, emp := range employees {
		empMap[emp.ID] = emp
	}

	currentYear := time.Now().Year()
	result := &BatchEnrollResult{}

	for _, empID := range req.EmployeeIDs {
		emp, ok := empMap[empID]
		if !ok {
			result.FailCount++
			result.Failures = append(result.Failures, EnrollFailure{
				EmployeeID:   empID,
				EmployeeName: "",
				Reason:       "员工不存在",
			})
			continue
		}

		// 检查是否已有 active 参保记录
		existing, _ := s.repo.FindActiveRecordByEmployee(orgID, empID)
		if existing != nil {
			result.FailCount++
			result.Failures = append(result.Failures, EnrollFailure{
				EmployeeID:   empID,
				EmployeeName: emp.Name,
				Reason:       "该员工已有参保中记录",
			})
			continue
		}

		// 查询城市政策
		policy, err := s.repo.FindByCityAndYear(req.CityID, currentYear)
		if err != nil {
			result.FailCount++
			result.Failures = append(result.Failures, EnrollFailure{
				EmployeeID:   empID,
				EmployeeName: emp.Name,
				Reason:       "该城市无社保政策",
			})
			continue
		}

		// 计算各险种金额
		details := calculateDetails(policy.Config.Data(), 0)
		detailsJSON, _ := json.Marshal(details)

		var totalCompany, totalPersonal float64
		for _, d := range details {
			totalCompany += d.CompanyAmount
			totalPersonal += d.PersonalAmount
		}

		// 创建参保记录
		record := &SocialInsuranceRecord{
			EmployeeID:    empID,
			EmployeeName:  emp.Name,
			CityID:        req.CityID,
			PolicyID:      policy.ID,
			BaseAmount:    details[0].Base,
			Status:        SIStatusActive,
			StartMonth:    req.StartMonth,
			Details:       datatypes.JSON(detailsJSON),
			TotalCompany:  math.Round(totalCompany*100) / 100,
			TotalPersonal: math.Round(totalPersonal*100) / 100,
		}
		record.OrgID = orgID
		record.CreatedBy = userID
		record.UpdatedBy = userID

		if err := s.repo.CreateRecord(record); err != nil {
			result.FailCount++
			result.Failures = append(result.Failures, EnrollFailure{
				EmployeeID:   empID,
				EmployeeName: emp.Name,
				Reason:       "创建参保记录失败: " + err.Error(),
			})
			continue
		}

		// 创建变更历史（enroll 类型）
		history := &ChangeHistory{
			RecordID:   record.ID,
			EmployeeID: empID,
			ChangeType: SIChangeEnroll,
			AfterValue: datatypes.JSON(detailsJSON),
			Remark:     "批量参保",
		}
		history.OrgID = orgID
		history.CreatedBy = userID
		history.UpdatedBy = userID

		_ = s.repo.CreateChangeHistory(history)

		result.SuccessCount++
	}

	return result, nil
}

// BatchStopEnrollment 批量停缴
func (s *Service) BatchStopEnrollment(orgID, userID int64, req *BatchStopRequest) (*BatchStopResult, error) {
	result := &BatchStopResult{}

	for _, recordID := range req.RecordIDs {
		// 查询参保记录
		record, err := s.repo.FindRecordByID(orgID, recordID)
		if err != nil {
			result.FailCount++
			result.Failures = append(result.Failures, StopFailure{
				RecordID: recordID,
				Reason:   "参保记录不存在",
			})
			continue
		}

		// 验证状态为 active
		if record.Status != SIStatusActive {
			result.FailCount++
			result.Failures = append(result.Failures, StopFailure{
				RecordID: recordID,
				Reason:   "参保记录非参保中状态，无法停缴",
			})
			continue
		}

		// 保存变更前快照
		beforeJSON, _ := json.Marshal(map[string]interface{}{
			"status":     record.Status,
			"end_month":  record.EndMonth,
		})

		// 更新状态和结束月份
		updates := map[string]interface{}{
			"status":    SIStatusStopped,
			"end_month": req.EndMonth,
			"updated_by": userID,
		}
		if err := s.repo.UpdateRecord(orgID, recordID, updates); err != nil {
			result.FailCount++
			result.Failures = append(result.Failures, StopFailure{
				RecordID: recordID,
				Reason:   "更新参保记录失败: " + err.Error(),
			})
			continue
		}

		afterJSON, _ := json.Marshal(map[string]interface{}{
			"status":    SIStatusStopped,
			"end_month": req.EndMonth,
		})

		// 创建变更历史（stop 类型）
		history := &ChangeHistory{
			RecordID:    recordID,
			EmployeeID:  record.EmployeeID,
			ChangeType:  SIChangeStop,
			BeforeValue: datatypes.JSON(beforeJSON),
			AfterValue:  datatypes.JSON(afterJSON),
			Remark:      "批量停缴",
		}
		history.OrgID = orgID
		history.CreatedBy = userID
		history.UpdatedBy = userID

		_ = s.repo.CreateChangeHistory(history)

		result.SuccessCount++
	}

	return result, nil
}

// ========== 查询操作 ==========

// ListRecords 参保记录列表
func (s *Service) ListRecords(orgID int64, params RecordListQueryParams) ([]RecordResponse, int64, int, int, error) {
	// 默认分页参数
	if params.Page == 0 {
		params.Page = 1
	}
	if params.PageSize == 0 {
		params.PageSize = 20
	}

	records, total, err := s.repo.ListRecords(orgID, params.Status, params.EmployeeName, params.Page, params.PageSize)
	if err != nil {
		return nil, 0, 0, 0, err
	}

	responses := make([]RecordResponse, 0, len(records))
	for _, r := range records {
		var details []InsuranceAmountDetail
		if r.Details != nil {
			_ = json.Unmarshal(r.Details, &details)
		}

		responses = append(responses, RecordResponse{
			ID:            r.ID,
			EmployeeID:    r.EmployeeID,
			EmployeeName:  r.EmployeeName,
			CityID:        r.CityID,
			CityName:      getCityName(r.CityID),
			BaseAmount:    r.BaseAmount,
			Status:        r.Status,
			StartMonth:    r.StartMonth,
			EndMonth:      r.EndMonth,
			Details:       details,
			TotalCompany:  r.TotalCompany,
			TotalPersonal: r.TotalPersonal,
			CreatedAt:     r.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	return responses, total, params.Page, params.PageSize, nil
}

// GetMyRecords MEMBER 查询自己的社保记录
func (s *Service) GetMyRecords(orgID, userID int64) ([]RecordResponse, error) {
	// 通过 user_id 查找员工
	empInfo, err := s.empQuerier.FindEmployeeByUserID(orgID, userID)
	if err != nil {
		return nil, fmt.Errorf("未找到关联的员工记录: %w", err)
	}

	records, err := s.repo.FindRecordsByEmployee(orgID, empInfo.ID)
	if err != nil {
		return nil, err
	}

	responses := make([]RecordResponse, 0, len(records))
	for _, r := range records {
		var details []InsuranceAmountDetail
		if r.Details != nil {
			_ = json.Unmarshal(r.Details, &details)
		}

		responses = append(responses, RecordResponse{
			ID:            r.ID,
			EmployeeID:    r.EmployeeID,
			EmployeeName:  r.EmployeeName,
			CityID:        r.CityID,
			CityName:      getCityName(r.CityID),
			BaseAmount:    r.BaseAmount,
			Status:        r.Status,
			StartMonth:    r.StartMonth,
			EndMonth:      r.EndMonth,
			Details:       details,
			TotalCompany:  r.TotalCompany,
			TotalPersonal: r.TotalPersonal,
			CreatedAt:     r.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	return responses, nil
}

// GetChangeHistory 变更历史查询
func (s *Service) GetChangeHistory(orgID, employeeID int64, page, pageSize int) ([]ChangeHistoryResponse, int64, error) {
	if page == 0 {
		page = 1
	}
	if pageSize == 0 {
		pageSize = 20
	}

	histories, total, err := s.repo.ListChangeHistories(orgID, employeeID, page, pageSize)
	if err != nil {
		return nil, 0, err
	}

	responses := make([]ChangeHistoryResponse, 0, len(histories))
	for _, h := range histories {
		var beforeValue, afterValue interface{}
		if h.BeforeValue != nil {
			_ = json.Unmarshal(h.BeforeValue, &beforeValue)
		}
		if h.AfterValue != nil {
			_ = json.Unmarshal(h.AfterValue, &afterValue)
		}

		responses = append(responses, ChangeHistoryResponse{
			ID:           h.ID,
			RecordID:     h.RecordID,
			EmployeeID:   h.EmployeeID,
			EmployeeName: "",
			ChangeType:   h.ChangeType,
			BeforeValue:  beforeValue,
			AfterValue:   afterValue,
			Remark:       h.Remark,
			CreatedAt:    h.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	return responses, total, nil
}

// GetSocialInsuranceDeduction 社保扣款查询（D-12，供 Phase 5 调用）
func (s *Service) GetSocialInsuranceDeduction(orgID, employeeID int64, month string) (*DeductionResponse, error) {
	// 查询该员工的 active 参保记录
	record, err := s.repo.FindActiveRecordByEmployee(orgID, employeeID)
	if err != nil {
		return nil, fmt.Errorf("未找到该员工的参保记录: %w", err)
	}

	var details []InsuranceAmountDetail
	if record.Details != nil {
		_ = json.Unmarshal(record.Details, &details)
	}

	var totalPersonal float64
	items := make([]DeductionItem, 0, len(details))
	for _, d := range details {
		items = append(items, DeductionItem{
			Name:           d.Name,
			PersonalRate:   d.PersonalRate,
			PersonalAmount: d.PersonalAmount,
		})
		totalPersonal += d.PersonalAmount
	}

	return &DeductionResponse{
		Items:         items,
		TotalPersonal: math.Round(totalPersonal*100) / 100,
	}, nil
}

// ========== 内部方法 ==========

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

// calculateDetails 计算各险种明细
func calculateDetails(config FiveInsurances, salary float64) []InsuranceAmountDetail {
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
	}

	return details
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
