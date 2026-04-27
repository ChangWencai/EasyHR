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
	repo         *Repository
	empQuerier   EmployeeQuerier
	reminderRepo *ReminderRepository
	cityRepo     *city.Repository
}

// NewService 创建社保 Service
func NewService(repo *Repository, empQuerier EmployeeQuerier, reminderRepo *ReminderRepository, cityRepo *city.Repository) *Service {
	return &Service{repo: repo, empQuerier: empQuerier, reminderRepo: reminderRepo, cityRepo: cityRepo}
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
func (s *Service) ListPolicies(cityID int64, page, pageSize int) ([]PolicyResponse, int64, error) {
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
func (s *Service) CalculateInsuranceAmounts(cityID int64, salary float64, year int) (*CalculateResponse, error) {
	// 1. 查询适用政策
	policy, err := s.repo.FindByCityAndYear(cityID, year)
	if err != nil {
		return nil, fmt.Errorf("未找到该城市社保政策: %w", err)
	}

	// 2. 获取城市名称
	cityName := s.getCityName(cityID)

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
			CityCode:      req.CityID,
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
			CityID:        r.CityCode,
			CityName:      s.getCityName(r.CityCode),
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
			CityID:        r.CityCode,
			CityName:      s.getCityName(r.CityCode),
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
	cityName := s.getCityName(policy.CityCode)
	return &PolicyResponse{
		ID:            policy.ID,
		CityID:        policy.CityCode,
		CityName:      cityName,
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
func (s *Service) getCityName(cityID int64) string {
	if s.cityRepo == nil {
		return "未知城市"
	}
	return s.cityRepo.GetNameByCode(cityID)
}

// ========== 提醒相关方法 ==========

// CheckPaymentDueReminders 每日扫描缴费到期提醒（D-09/D-10/D-11）
// 每月缴费截止日默认15日，到期前3天生成提醒
func (s *Service) CheckPaymentDueReminders() {
	if s.reminderRepo == nil {
		return
	}

	records, err := s.reminderRepo.FindActiveRecordsGroupedByOrg()
	if err != nil {
		return
	}

	// 按 org_id 分组
	orgRecords := make(map[int64][]SocialInsuranceRecord)
	for _, r := range records {
		orgRecords[r.OrgID] = append(orgRecords[r.OrgID], r)
	}

	now := time.Now()
	cstZone := time.FixedZone("CST", 8*3600)
	nowCST := now.In(cstZone)

	// 缴费截止日为每月15日
	paymentDay := 15

	// 计算当月截止日
	dueDate := time.Date(nowCST.Year(), nowCST.Month(), paymentDay, 23, 59, 59, 0, cstZone)
	// 如果已过当月截止日，使用下月截止日
	if nowCST.Day() > paymentDay {
		nextMonth := nowCST.AddDate(0, 1, 0)
		dueDate = time.Date(nextMonth.Year(), nextMonth.Month(), paymentDay, 23, 59, 59, 0, cstZone)
	}

	daysUntilDue := int(dueDate.Sub(nowCST).Hours() / 24)

	// 只在到期前3天内生成提醒（0 <= days <= 3）
	if daysUntilDue < 0 || daysUntilDue > 3 {
		return
	}

	for orgID, orgRecs := range orgRecords {
		// 按企业汇总，同一企业同一月只生成一条汇总提醒
		// 使用一个虚拟 record_id=0 表示汇总提醒
		monthKey := nowCST.Format("2006-01")
		existing, _ := s.reminderRepo.FindByTypeAndRecordID(orgID, ReminderTypePaymentDue, 0)
		if existing != nil {
			// 已有当月提醒，检查是否是同一个月
			if existing.Title != "" {
				continue // 已生成过，跳过去重
			}
		}

		// 去重：检查是否已有该企业该月的汇总提醒
		var count int64
		s.reminderRepo.db.Model(&Reminder{}).
			Where("org_id = ? AND type = ? AND title LIKE ?", orgID, ReminderTypePaymentDue, "%"+monthKey+"%").
			Count(&count)
		if count > 0 {
			continue
		}

		dueDateOnly := time.Date(dueDate.Year(), dueDate.Month(), dueDate.Day(), 0, 0, 0, 0, time.UTC)
		reminder := &Reminder{
			Type:        ReminderTypePaymentDue,
			Title:       fmt.Sprintf("社保缴费提醒：%d名员工社保将于%s到期，请及时缴费", len(orgRecs), dueDate.Format("2006-01-02")),
			RecordID:    0, // 汇总提醒
			DueDate:     &dueDateOnly,
			IsRead:      false,
			IsDismissed: false,
		}
		reminder.OrgID = orgID

		_ = s.reminderRepo.Create(reminder)
	}
}

// CreateStopReminder 创建停缴提醒（D-07 离职触发）
func (s *Service) CreateStopReminder(orgID, employeeID int64) {
	if s.reminderRepo == nil {
		return
	}

	// 查询员工是否有 active 参保记录
	record, err := s.repo.FindActiveRecordByEmployee(orgID, employeeID)
	if err != nil {
		// 无 active 记录，静默跳过
		return
	}

	// 去重检查
	existing, _ := s.reminderRepo.FindByTypeAndRecordID(orgID, ReminderTypeStop, record.ID)
	if existing != nil {
		return
	}

	reminder := &Reminder{
		Type:        ReminderTypeStop,
		Title:       fmt.Sprintf("社保停缴提醒：%s已离职，请及时办理社保停缴", record.EmployeeName),
		EmployeeID:  employeeID,
		RecordID:    record.ID,
		IsRead:      false,
		IsDismissed: false,
	}
	reminder.OrgID = orgID

	_ = s.reminderRepo.Create(reminder)
}

// OnEmployeeResigned 实现 employee.SocialInsuranceEventHandler 接口
func (s *Service) OnEmployeeResigned(orgID, employeeID int64) {
	s.CreateStopReminder(orgID, employeeID)
}

// SuggestBaseAdjustment 薪资变动时建议基数调整（D-13/SOCL-06 预留接口）
func (s *Service) SuggestBaseAdjustment(orgID, employeeID int64, newSalary float64) {
	if s.reminderRepo == nil {
		return
	}

	// 查询员工 active 参保记录
	record, err := s.repo.FindActiveRecordByEmployee(orgID, employeeID)
	if err != nil {
		return
	}

	// 对比 newSalary 与 BaseAmount，偏差超过10%时建议调整
	if record.BaseAmount == 0 {
		return
	}
	diff := (newSalary - record.BaseAmount) / record.BaseAmount
	if diff < 0 {
		diff = -diff
	}
	if diff <= 0.1 {
		return
	}

	// 去重检查
	existing, _ := s.reminderRepo.FindByTypeAndRecordID(orgID, ReminderTypeBaseAdjust, record.ID)
	if existing != nil {
		return
	}

	reminder := &Reminder{
		Type:       ReminderTypeBaseAdjust,
		Title:      fmt.Sprintf("社保基数调整建议：%s薪资变动，建议调整社保基数", record.EmployeeName),
		EmployeeID: employeeID,
		RecordID:   record.ID,
		IsRead:     false,
	}
	reminder.OrgID = orgID

	_ = s.reminderRepo.Create(reminder)
}

// ListReminders 查询提醒列表
func (s *Service) ListReminders(orgID int64, reminderType string, page, pageSize int) ([]Reminder, int64, error) {
	if page == 0 {
		page = 1
	}
	if pageSize == 0 {
		pageSize = 20
	}
	return s.reminderRepo.ListUnread(orgID, reminderType, page, pageSize)
}

// DismissReminder 关闭提醒
func (s *Service) DismissReminder(orgID, id int64) error {
	return s.reminderRepo.Dismiss(orgID, id)
}

// ========== 导出方法 ==========

// ExportPaymentDetailExcel 导出缴费明细 Excel（SOCL-05）
func (s *Service) ExportPaymentDetailExcel(orgID int64, params RecordListQueryParams) ([]byte, error) {
	if params.Page == 0 {
		params.Page = 1
	}
	if params.PageSize == 0 {
		params.PageSize = 1000 // 导出时使用较大页面
	}

	records, _, err := s.repo.ListRecords(orgID, params.Status, params.EmployeeName, params.Page, params.PageSize)
	if err != nil {
		return nil, fmt.Errorf("查询参保记录失败: %w", err)
	}

	return generatePaymentDetailExcel(records, s)
}

// GenerateEnrollmentPDF 生成参保材料 PDF（SOCL-02）
func (s *Service) GenerateEnrollmentPDF(orgID, recordID int64) ([]byte, error) {
	// 查询参保记录
	record, err := s.repo.FindRecordByID(orgID, recordID)
	if err != nil {
		return nil, fmt.Errorf("参保记录不存在")
	}

	// 解析险种明细
	var details []InsuranceAmountDetail
	if record.Details != nil {
		_ = json.Unmarshal(record.Details, &details)
	}

	data := &EnrollmentPDFData{
		EmployeeName: record.EmployeeName,
		CityName:     s.getCityName(record.CityCode),
		BaseAmount:   record.BaseAmount,
		StartMonth:   record.StartMonth,
		Items:        details,
		TotalCompany: record.TotalCompany,
		TotalPersonal: record.TotalPersonal,
	}

	return generateEnrollmentPDF(data)
}
