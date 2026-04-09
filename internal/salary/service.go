package salary

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// Service 工资核算业务逻辑层
type Service struct {
	repo               *Repository
	templateRepo       *SalaryTemplateRepository
	taxProvider        TaxProvider
	siProvider         SIDeductionProvider
	empProvider        EmployeeProvider
	baseAdjustProvider BaseAdjustmentProvider
}

// NewService 创建工资核算 Service
func NewService(
	repo *Repository,
	templateRepo *SalaryTemplateRepository,
	taxProvider TaxProvider,
	siProvider SIDeductionProvider,
	empProvider EmployeeProvider,
	baseAdjustProvider BaseAdjustmentProvider,
) *Service {
	return &Service{
		repo:               repo,
		templateRepo:       templateRepo,
		taxProvider:        taxProvider,
		siProvider:         siProvider,
		empProvider:        empProvider,
		baseAdjustProvider: baseAdjustProvider,
	}
}

// SeedTemplateItems 初始化预置薪资项模板
func (s *Service) SeedTemplateItems() error {
	return s.templateRepo.SeedGlobalTemplateItems()
}

// GetTemplate 获取企业薪资模板（含启用状态）
func (s *Service) GetTemplate(orgID int64) (*TemplateResponse, error) {
	globalItems, err := s.templateRepo.GetGlobalItems()
	if err != nil {
		return nil, fmt.Errorf("获取全局模板失败: %w", err)
	}

	overrides, err := s.templateRepo.GetOrgOverrides(orgID)
	if err != nil {
		return nil, fmt.Errorf("获取企业配置失败: %w", err)
	}

	overrideMap := make(map[string]bool)
	for _, o := range overrides {
		overrideMap[o.Name] = o.IsEnabled
	}

	items := make([]TemplateItemResponse, 0, len(globalItems))
	for _, g := range globalItems {
		isEnabled := g.IsEnabled
		if overridden, ok := overrideMap[g.Name]; ok {
			isEnabled = overridden
		}
		items = append(items, TemplateItemResponse{
			ID:         g.ID,
			Name:       g.Name,
			Type:       g.Type,
			SortOrder:  g.SortOrder,
			IsRequired: g.IsRequired,
			IsEnabled:  isEnabled,
		})
	}

	return &TemplateResponse{Items: items}, nil
}

// UpdateTemplate 批量更新企业薪资项启用/禁用
func (s *Service) UpdateTemplate(orgID, userID int64, items []TemplateItemUpdate) error {
	for _, item := range items {
		if err := s.templateRepo.UpsertOrgOverride(orgID, userID, item.TemplateItemID, item.IsEnabled); err != nil {
			return fmt.Errorf("更新模板项 %d 失败: %w", item.TemplateItemID, err)
		}
	}
	return nil
}

// GetEmployeeItems 获取员工某月各项金额
func (s *Service) GetEmployeeItems(orgID, employeeID int64, month string) (*EmployeeItemsResponse, error) {
	salaryItems, err := s.repo.FindSalaryItemsByEmployee(orgID, employeeID, month)
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("查询员工薪资项失败: %w", err)
	}

	template, err := s.GetTemplate(orgID)
	if err != nil {
		return nil, err
	}

	templateMap := make(map[int64]TemplateItemResponse)
	for _, t := range template.Items {
		templateMap[t.ID] = t
	}

	itemMap := make(map[int64]float64)
	for _, si := range salaryItems {
		itemMap[si.TemplateItemID] = si.Amount
	}

	respItems := make([]EmployeeItemResponse, 0)
	for _, t := range template.Items {
		if !t.IsEnabled {
			continue
		}
		amount := itemMap[t.ID]
		respItems = append(respItems, EmployeeItemResponse{
			TemplateItemID: t.ID,
			ItemName:       t.Name,
			ItemType:       t.Type,
			Amount:         amount,
		})
	}

	return &EmployeeItemsResponse{
		EmployeeID: employeeID,
		Month:      month,
		Items:      respItems,
	}, nil
}

// SetEmployeeItems 设置员工各项金额
func (s *Service) SetEmployeeItems(orgID, userID, employeeID int64, month string, items []SalaryItemInput) error {
	if len(month) != 7 || month[4] != '-' {
		return fmt.Errorf("月份格式错误，应为 YYYY-MM")
	}

	template, err := s.GetTemplate(orgID)
	if err != nil {
		return err
	}
	enabledMap := make(map[int64]bool)
	for _, t := range template.Items {
		if t.IsEnabled {
			enabledMap[t.ID] = true
		}
	}

	for _, item := range items {
		if item.Amount < 0 {
			return fmt.Errorf("薪资金额不能为负数")
		}
		if !enabledMap[item.TemplateItemID] {
			return fmt.Errorf("薪资项 %d 未启用", item.TemplateItemID)
		}
		if err := s.repo.UpsertSalaryItem(orgID, userID, employeeID, item.TemplateItemID, month, item.Amount); err != nil {
			return fmt.Errorf("设置薪资项失败: %w", err)
		}
	}
	return nil
}

// ========== 工资核算流程 ==========

// CreatePayroll 创建月度工资表
func (s *Service) CreatePayroll(orgID, userID int64, year, month int, copyFromMonth *string) ([]PayrollRecordResponse, error) {
	existing, _ := s.repo.FindPayrollRecordsByMonth(orgID, year, month)
	if len(existing) > 0 {
		return nil, fmt.Errorf("%d年%d月工资表已存在", year, month)
	}

	employees, err := s.empProvider.GetActiveEmployees(orgID)
	if err != nil {
		return nil, fmt.Errorf("获取员工列表失败: %w", err)
	}

	monthStr := fmt.Sprintf("%d-%02d", year, month)
	var records []PayrollRecordResponse

	for _, emp := range employees {
		record := PayrollRecord{
			EmployeeID:   emp.ID,
			EmployeeName: emp.Name,
			Year:         year,
			Month:        month,
			Status:       PayrollStatusDraft,
		}
		record.OrgID = orgID
		record.CreatedBy = userID
		record.UpdatedBy = userID

		if err := s.repo.CreatePayrollRecord(&record); err != nil {
			return nil, fmt.Errorf("创建员工 %s 工资记录失败: %w", emp.Name, err)
		}

		if copyFromMonth != nil {
			s.copyFromPreviousMonth(orgID, userID, emp.ID, *copyFromMonth, monthStr)
		} else if emp.BaseSalary > 0 {
			tmpl, _ := s.GetTemplate(orgID)
			if tmpl != nil {
				for _, t := range tmpl.Items {
					if t.Name == "基本工资" && t.IsEnabled {
						_ = s.repo.UpsertSalaryItem(orgID, userID, emp.ID, t.ID, monthStr, emp.BaseSalary)
						break
					}
				}
			}
		}

		records = append(records, PayrollRecordResponse{
			ID:           record.ID,
			EmployeeID:   record.EmployeeID,
			EmployeeName: record.EmployeeName,
			Year:         record.Year,
			Month:        record.Month,
			Status:       record.Status,
		})
	}

	return records, nil
}

// copyFromPreviousMonth 复制上月薪资项数据
func (s *Service) copyFromPreviousMonth(orgID, userID, employeeID int64, fromMonth, toMonth string) {
	prevItems, err := s.repo.FindSalaryItemsByEmployee(orgID, employeeID, fromMonth)
	if err != nil {
		return
	}
	tmpl, err := s.GetTemplate(orgID)
	if err != nil {
		return
	}

	enabledMap := make(map[int64]bool)
	templateMap := make(map[int64]TemplateItemResponse)
	for _, t := range tmpl.Items {
		if t.IsEnabled {
			enabledMap[t.ID] = true
		}
		templateMap[t.ID] = t
	}

	for _, item := range prevItems {
		if !enabledMap[item.TemplateItemID] {
			continue
		}
		if t, ok := templateMap[item.TemplateItemID]; ok {
			if t.Name == "事假扣款" || t.Name == "病假扣款" {
				continue
			}
		}
		_ = s.repo.UpsertSalaryItem(orgID, userID, employeeID, item.TemplateItemID, toMonth, item.Amount)
	}
}

// CalculatePayroll 一键核算
func (s *Service) CalculatePayroll(orgID, userID int64, year, month int) (*BatchCalculateResponse, error) {
	records, err := s.repo.FindPayrollRecordsByMonth(orgID, year, month)
	if err != nil {
		return nil, fmt.Errorf("查询工资记录失败: %w", err)
	}

	monthStr := fmt.Sprintf("%d-%02d", year, month)
	var totalNetIncome float64

	for i := range records {
		rec := &records[i]
		if rec.Status != PayrollStatusDraft && rec.Status != PayrollStatusCalculated {
			continue
		}

		salaryItems, err := s.repo.FindSalaryItemsByEmployee(orgID, rec.EmployeeID, monthStr)
		if err != nil {
			return nil, fmt.Errorf("获取员工 %s 薪资项失败: %w", rec.EmployeeName, err)
		}

		tmpl, _ := s.GetTemplate(orgID)
		templateMap := make(map[int64]TemplateItemResponse)
		for _, t := range tmpl.Items {
			templateMap[t.ID] = t
		}

		var inputs []PayrollItemInput
		for _, si := range salaryItems {
			if t, ok := templateMap[si.TemplateItemID]; ok {
				inputs = append(inputs, PayrollItemInput{
					ItemName: t.Name,
					ItemType: t.Type,
					Amount:   si.Amount,
				})
			}
		}

		siDeduction := float64(0)
		if s.siProvider != nil {
			if deduction, err := s.siProvider.GetPersonalDeduction(orgID, rec.EmployeeID, monthStr); err == nil {
				siDeduction = deduction
			}
		}

		var grossIncome float64
		for _, inp := range inputs {
			if inp.ItemType == "income" {
				grossIncome += inp.Amount
			}
		}

		var payrollResult *PayrollResult
		if s.taxProvider != nil {
			taxResult, err := s.taxProvider.CalculateTax(orgID, rec.EmployeeID, year, month, grossIncome)
			if err == nil && taxResult != nil {
				payrollResult = calculatePayroll(inputs, siDeduction, taxResult)
			}
		}
		if payrollResult == nil {
			payrollResult = calculatePayroll(inputs, siDeduction, nil)
		}

		rec.GrossIncome = payrollResult.GrossIncome
		rec.SIDeduction = payrollResult.SIDeduction
		rec.Tax = payrollResult.Tax
		rec.TotalDeductions = payrollResult.TotalDeductions
		rec.NetIncome = payrollResult.NetIncome
		rec.Status = PayrollStatusCalculated
		rec.UpdatedBy = userID

		if err := s.repo.UpdatePayrollRecord(orgID, rec); err != nil {
			return nil, fmt.Errorf("更新员工 %s 工资记录失败: %w", rec.EmployeeName, err)
		}

		_ = s.repo.DeletePayrollItemsByRecord(orgID, rec.ID)
		var payrollItems []PayrollItem
		for _, inp := range inputs {
			pi := PayrollItem{
				PayrollRecordID: rec.ID,
				ItemName:        inp.ItemName,
				ItemType:        inp.ItemType,
				Amount:          inp.Amount,
			}
			pi.OrgID = orgID
			pi.CreatedBy = userID
			pi.UpdatedBy = userID
			payrollItems = append(payrollItems, pi)
		}
		if err := s.repo.BatchCreatePayrollItems(orgID, payrollItems); err != nil {
			return nil, fmt.Errorf("保存员工 %s 工资明细失败: %w", rec.EmployeeName, err)
		}

		totalNetIncome += payrollResult.NetIncome
	}

	return &BatchCalculateResponse{
		TotalEmployees: len(records),
		TotalNetIncome: roundTo2Salary(totalNetIncome),
	}, nil
}

// GetPayrollList 查询工资表列表
func (s *Service) GetPayrollList(orgID int64, year, month, page, pageSize int) ([]PayrollRecordResponse, int64, error) {
	records, total, err := s.repo.ListPayrollRecords(orgID, year, month, page, pageSize)
	if err != nil {
		return nil, 0, err
	}

	var result []PayrollRecordResponse
	for _, rec := range records {
		r := s.toPayrollRecordResponse(orgID, &rec)
		result = append(result, r)
	}

	return result, total, nil
}

// GetPayrollDetail 查询单个工资记录详情
func (s *Service) GetPayrollDetail(orgID, recordID int64) (*PayrollRecordResponse, error) {
	rec, err := s.repo.FindPayrollRecordByID(orgID, recordID)
	if err != nil {
		return nil, fmt.Errorf("工资记录不存在: %w", err)
	}

	resp := s.toPayrollRecordResponse(orgID, rec)
	return &resp, nil
}

// toPayrollRecordResponse 转换为响应 DTO
func (s *Service) toPayrollRecordResponse(orgID int64, rec *PayrollRecord) PayrollRecordResponse {
	r := PayrollRecordResponse{
		ID:              rec.ID,
		EmployeeID:      rec.EmployeeID,
		EmployeeName:    rec.EmployeeName,
		Year:            rec.Year,
		Month:           rec.Month,
		Status:          rec.Status,
		GrossIncome:     rec.GrossIncome,
		SIDeduction:     rec.SIDeduction,
		Tax:             rec.Tax,
		TotalDeductions: rec.TotalDeductions,
		NetIncome:       rec.NetIncome,
		PayMethod:       rec.PayMethod,
		PayNote:         rec.PayNote,
	}
	if rec.PayDate != nil {
		pd := rec.PayDate.Format("2006-01-02")
		r.PayDate = &pd
	}

	items, _ := s.repo.FindPayrollItemsByRecord(orgID, rec.ID)
	for _, item := range items {
		r.Items = append(r.Items, PayrollItemResponse{
			ItemName: item.ItemName,
			ItemType: item.ItemType,
			Amount:   item.Amount,
		})
	}

	return r
}

// ConfirmPayroll 确认工资表
func (s *Service) ConfirmPayroll(orgID, userID int64, year, month int) (*ConfirmResponse, error) {
	records, err := s.repo.FindPayrollRecordsByMonth(orgID, year, month)
	if err != nil {
		return nil, fmt.Errorf("查询工资记录失败: %w", err)
	}

	for _, rec := range records {
		if rec.Status != PayrollStatusCalculated {
			return nil, fmt.Errorf("员工 %s 的工资记录状态为 %s，无法确认", rec.EmployeeName, rec.Status)
		}
	}

	prevRecords, _ := s.repo.FindPreviousMonthRecords(orgID, year, month)
	prevNetMap := make(map[int64]float64)
	for _, pr := range prevRecords {
		prevNetMap[pr.EmployeeID] = pr.NetIncome
	}

	var abnormalInputs []AbnormalCheckInput
	for _, rec := range records {
		abnormalInputs = append(abnormalInputs, AbnormalCheckInput{
			EmployeeID:   rec.EmployeeID,
			EmployeeName: rec.EmployeeName,
			NetIncome:    rec.NetIncome,
		})
	}
	abnormalItems := checkAbnormalPayments(abnormalInputs, prevNetMap)

	for i := range records {
		records[i].Status = PayrollStatusConfirmed
		records[i].UpdatedBy = userID
		if err := s.repo.UpdatePayrollRecord(orgID, &records[i]); err != nil {
			return nil, fmt.Errorf("确认员工 %s 失败: %w", records[i].EmployeeName, err)
		}

		if s.baseAdjustProvider != nil {
			s.baseAdjustProvider.SuggestBaseAdjustment(orgID, records[i].EmployeeID, records[i].GrossIncome)
		}
	}

	return &ConfirmResponse{
		ConfirmedCount: len(records),
		AbnormalItems:  abnormalItems,
	}, nil
}

// RecordPayment 发放记录
func (s *Service) RecordPayment(orgID, userID, recordID int64, req *RecordPaymentRequest) error {
	rec, err := s.repo.FindPayrollRecordByID(orgID, recordID)
	if err != nil {
		return fmt.Errorf("工资记录不存在: %w", err)
	}

	if rec.Status != PayrollStatusConfirmed {
		return fmt.Errorf("只有已确认的工资记录才能标记发放，当前状态: %s", rec.Status)
	}

	payDate, err := time.Parse("2006-01-02", req.PayDate)
	if err != nil {
		return fmt.Errorf("发放日期格式错误: %w", err)
	}

	rec.Status = PayrollStatusPaid
	rec.PayMethod = req.PayMethod
	rec.PayDate = &payDate
	rec.PayNote = req.PayNote
	rec.UpdatedBy = userID

	return s.repo.UpdatePayrollRecord(orgID, rec)
}

// ImportAttendance 考勤 Excel 导入
func (s *Service) ImportAttendance(orgID, userID int64, year, month int, file []byte) (*AttendanceImportResult, error) {
	rows, err := parseAttendanceExcel(file)
	if err != nil {
		return nil, fmt.Errorf("解析考勤文件失败: %w", err)
	}

	monthStr := fmt.Sprintf("%d-%02d", year, month)

	employees, err := s.empProvider.GetActiveEmployees(orgID)
	if err != nil {
		return nil, fmt.Errorf("获取员工列表失败: %w", err)
	}

	nameMap := make(map[string]*EmployeeInfo)
	for i := range employees {
		nameMap[employees[i].Name] = &employees[i]
	}

	tmpl, err := s.GetTemplate(orgID)
	if err != nil {
		return nil, err
	}

	var sickTemplateID, personalTemplateID int64
	for _, t := range tmpl.Items {
		if t.Name == "事假扣款" {
			personalTemplateID = t.ID
		}
		if t.Name == "病假扣款" {
			sickTemplateID = t.ID
		}
	}

	result := &AttendanceImportResult{}
	for rowIdx, row := range rows {
		emp, ok := nameMap[row.Name]
		if !ok {
			result.ErrorRows = append(result.ErrorRows, AttendanceErrorRow{
				RowNumber: rowIdx + 2,
				Name:      row.Name,
				Error:     "未找到匹配的在职员工",
			})
			continue
		}

		if row.PersonalLeaveDays > 0 && personalTemplateID > 0 {
			dailyWage := calculateDailyWage(emp.BaseSalary)
			deduction := calculateLeaveDeduction(dailyWage, row.PersonalLeaveDays)
			_ = s.repo.UpsertSalaryItem(orgID, userID, emp.ID, personalTemplateID, monthStr, deduction)
		}
		if row.SickLeaveDays > 0 && sickTemplateID > 0 {
			dailyWage := calculateDailyWage(emp.BaseSalary)
			deduction := calculateLeaveDeduction(dailyWage, row.SickLeaveDays)
			_ = s.repo.UpsertSalaryItem(orgID, userID, emp.ID, sickTemplateID, monthStr, deduction)
		}

		result.MatchedCount++
	}

	return result, nil
}

// generateSlipToken 生成工资单查看 token（64 字符 hex 字符串）
func generateSlipToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("生成工资单 token 失败: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}
