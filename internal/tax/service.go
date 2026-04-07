package tax

import (
	"fmt"
	"time"
)

// Service 个税业务逻辑层
type Service struct {
	repo        *Repository
	empProvider EmployeeInfoProvider
	siProvider  SIDeductionProvider
}

// NewService 创建个税 Service
func NewService(repo *Repository, empProvider EmployeeInfoProvider, siProvider SIDeductionProvider) *Service {
	return &Service{
		repo:        repo,
		empProvider: empProvider,
		siProvider:  siProvider,
	}
}

// ========== 专项附加扣除管理 ==========

// CreateDeduction 创建专项附加扣除
func (s *Service) CreateDeduction(orgID, userID int64, req *CreateDeductionRequest) (*DeductionResponse, error) {
	// 1. 校验扣除类型合法性
	if !isValidDeductionType(req.DeductionType) {
		return nil, ErrInvalidDeductionType
	}

	// 2. 检查互斥: housing_loan/housing_rent
	if mutexType, ok := MutualExclusionGroup[req.DeductionType]; ok {
		existing, _ := s.repo.FindDeductionByEmployeeAndType(orgID, req.EmployeeID, mutexType)
		if existing != nil {
			return nil, ErrMutuallyExclusiveDeduction
		}
	}

	// 3. 检查去重: 同员工同类型不能重复
	existing, _ := s.repo.FindDeductionByEmployeeAndType(orgID, req.EmployeeID, req.DeductionType)
	if existing != nil {
		return nil, ErrDuplicateDeduction
	}

	// 4. 计算 MonthlyAmount
	standard, ok := DeductionStandard[req.DeductionType]
	if !ok {
		return nil, ErrInvalidDeductionType
	}
	monthlyAmount := standard * float64(req.Count)

	deduction := &SpecialDeduction{
		EmployeeID:     req.EmployeeID,
		DeductionType:  req.DeductionType,
		MonthlyAmount:  monthlyAmount,
		Count:          req.Count,
		EffectiveStart: req.EffectiveStart,
		EffectiveEnd:   req.EffectiveEnd,
		Remark:         req.Remark,
	}
	deduction.OrgID = orgID
	deduction.CreatedBy = userID
	deduction.UpdatedBy = userID

	if err := s.repo.CreateDeduction(orgID, deduction); err != nil {
		return nil, fmt.Errorf("创建专项附加扣除失败: %w", err)
	}

	return s.toDeductionResponse(deduction), nil
}

// UpdateDeduction 更新专项附加扣除
func (s *Service) UpdateDeduction(orgID, userID, id int64, req *UpdateDeductionRequest) error {
	deduction, err := s.repo.FindDeductionByID(orgID, id)
	if err != nil {
		return ErrDeductionNotFound
	}

	updates := map[string]interface{}{
		"count":       req.Count,
		"remark":      req.Remark,
		"updated_by":  userID,
	}

	if req.EffectiveEnd != nil {
		updates["effective_end"] = *req.EffectiveEnd
	}

	// 重新计算 MonthlyAmount
	standard, ok := DeductionStandard[deduction.DeductionType]
	if ok {
		updates["monthly_amount"] = standard * float64(req.Count)
	}

	return s.repo.UpdateDeduction(orgID, id, updates)
}

// DeleteDeduction 删除专项附加扣除
func (s *Service) DeleteDeduction(orgID, id int64) error {
	return s.repo.DeleteDeduction(orgID, id)
}

// ListDeductions 查询专项附加扣除列表
func (s *Service) ListDeductions(orgID int64, params DeductionListQuery) ([]DeductionResponse, int64, error) {
	page := params.Page
	pageSize := params.PageSize
	if page == 0 {
		page = 1
	}
	if pageSize == 0 {
		pageSize = 20
	}

	// 如果没有指定月份，使用当前月
	month := time.Now().Format("2006-01")

	deductions, total, err := s.repo.ListDeductionsByEmployee(orgID, params.EmployeeID, month, page, pageSize)
	if err != nil {
		return nil, 0, err
	}

	responses := make([]DeductionResponse, 0, len(deductions))
	for _, d := range deductions {
		responses = append(responses, *s.toDeductionResponse(&d))
	}
	return responses, total, nil
}

// GetDeduction 获取单个扣除详情
func (s *Service) GetDeduction(orgID, id int64) (*DeductionResponse, error) {
	deduction, err := s.repo.FindDeductionByID(orgID, id)
	if err != nil {
		return nil, ErrDeductionNotFound
	}
	return s.toDeductionResponse(deduction), nil
}

// ========== 税率表管理 ==========

// ListTaxBrackets 查询税率表列表
func (s *Service) ListTaxBrackets(year, page, pageSize int) ([]TaxBracketResponse, int64, error) {
	brackets, err := s.repo.FindTaxBrackets(year)
	if err != nil {
		return nil, 0, err
	}

	// 税率表数量少，不需要分页
	responses := make([]TaxBracketResponse, 0, len(brackets))
	for _, b := range brackets {
		responses = append(responses, TaxBracketResponse{
			ID:             b.ID,
			Level:          b.Level,
			LowerBound:     b.LowerBound,
			UpperBound:     b.UpperBound,
			Rate:           b.Rate,
			QuickDeduction: b.QuickDeduction,
			EffectiveYear:  b.EffectiveYear,
		})
	}
	return responses, int64(len(responses)), nil
}

// SeedDefaultBrackets 种子税率表
func (s *Service) SeedDefaultBrackets(year int) error {
	return s.repo.SeedTaxBrackets(year)
}

// ========== 个税计算 ==========

// CalculateTax 计算个税（实现 TaxCalculator 接口）
func (s *Service) CalculateTax(orgID, employeeID int64, year, month int, grossIncome float64) (*TaxResult, error) {
	if err := ValidateTaxCalculationParams(year, month, grossIncome); err != nil {
		return nil, err
	}

	// 1. 获取税率表
	brackets, err := s.repo.FindTaxBrackets(year)
	if err != nil {
		return nil, fmt.Errorf("获取税率表失败: %w", err)
	}

	// 2. 获取本年已有 TaxRecord 列表
	records, err := s.repo.FindTaxRecordsByEmployeeYear(orgID, employeeID, year)
	if err != nil {
		records = []TaxRecord{} // 无记录时从零开始
	}

	// 过滤：只取当月之前的记录
	records = GetTaxRecordsForCumulative(records, month)

	// 3. 查询当月专项附加扣除总额
	monthStr := fmt.Sprintf("%d-%02d", year, month)
	activeDeductions, err := s.repo.ListAllActiveDeductionsByEmployee(orgID, employeeID, monthStr)
	if err != nil {
		activeDeductions = []SpecialDeduction{} // 无扣除时为0
	}
	var specialDeduction float64
	for _, d := range activeDeductions {
		specialDeduction += d.MonthlyAmount
	}
	specialDeduction = roundTo2(specialDeduction)

	// 4. 查询当月社保个人扣款
	var siDeduction float64
	if s.siProvider != nil {
		siDeduction, err = s.siProvider.GetPersonalDeduction(orgID, employeeID, monthStr)
		if err != nil {
			siDeduction = 0 // 获取失败时默认为0
		}
	}

	// 5. 调用计算引擎
	result := calculateCumulativeTax(
		brackets,
		records,
		grossIncome,
		BasicDeductionMonthly,
		siDeduction,
		specialDeduction,
	)

	// 6. 创建 TaxRecord 存储完整快照
	taxRecord := &TaxRecord{
		EmployeeID:                 employeeID,
		EmployeeName:               "", // 由调用方或 adapter 填充
		Year:                       year,
		Month:                      month,
		GrossIncome:                result.GrossIncome,
		BasicDeduction:             result.BasicDeduction,
		SIDeduction:                result.SIDeduction,
		SpecialDeduction:           result.SpecialDeduction,
		TotalDeduction:             result.TotalDeduction,
		CumulativeIncome:           result.CumulativeIncome,
		CumulativeBasicDeduction:   roundTo2(float64(len(records)+1) * BasicDeductionMonthly),
		CumulativeSIDeduction:      roundTo2(func() float64 {
			var sum float64
			for _, r := range records {
				sum += r.SIDeduction
			}
			return sum + siDeduction
		}()),
		CumulativeSpecialDeduction: roundTo2(func() float64 {
			var sum float64
			for _, r := range records {
				sum += r.SpecialDeduction
			}
			return sum + specialDeduction
		}()),
		CumulativeTaxableIncome:    result.CumulativeTaxableIncome,
		TaxRate:                    result.TaxRate,
		QuickDeduction:             result.QuickDeduction,
		CumulativeTax:              result.CumulativeTax,
		MonthlyTax:                 result.MonthlyTax,
		Source:                     "contract",
	}
	taxRecord.OrgID = orgID
	taxRecord.CreatedBy = 0  // 系统自动生成
	taxRecord.UpdatedBy = 0

	// 尝试创建记录（如果已有当月记录则忽略错误）
	_ = s.repo.CreateTaxRecord(taxRecord)

	return result, nil
}

// CalculateTaxFromContract 独立查询模式（从 Contract.Salary 获取 grossIncome）
func (s *Service) CalculateTaxFromContract(orgID, employeeID int64, year, month int) (*TaxResult, error) {
	if s.empProvider == nil {
		return nil, fmt.Errorf("未配置员工信息提供者")
	}

	salary, err := s.empProvider.GetActiveSalary(orgID, employeeID)
	if err != nil {
		return nil, fmt.Errorf("获取员工薪资失败: %w", err)
	}

	return s.CalculateTax(orgID, employeeID, year, month, salary)
}

// ========== 个税记录查询 ==========

// GetTaxRecord 获取个税记录详情
func (s *Service) GetTaxRecord(orgID, id int64) (*TaxRecordResponse, error) {
	record, err := s.repo.FindTaxRecordByID(orgID, id)
	if err != nil {
		return nil, ErrTaxRecordNotFound
	}
	return s.toTaxRecordResponse(record), nil
}

// ListTaxRecords 查询个税记录列表
func (s *Service) ListTaxRecords(orgID int64, params TaxRecordListQuery) ([]TaxRecordResponse, int64, error) {
	records, total, err := s.repo.ListTaxRecords(orgID, params)
	if err != nil {
		return nil, 0, err
	}

	responses := make([]TaxRecordResponse, 0, len(records))
	for _, r := range records {
		responses = append(responses, *s.toTaxRecordResponse(&r))
	}
	return responses, total, nil
}

// ========== 申报管理 ==========

// GetOrCreateDeclaration 获取或创建当月申报记录
func (s *Service) GetOrCreateDeclaration(orgID int64, year, month int) (*TaxDeclaration, error) {
	decl, err := s.repo.FindDeclarationByMonth(orgID, year, month)
	if err == nil {
		return decl, nil
	}

	// 创建新记录
	decl = &TaxDeclaration{
		Year:  year,
		Month: month,
		Status: DeclarationStatusPending,
	}
	decl.OrgID = orgID

	if createErr := s.repo.CreateDeclaration(decl); createErr != nil {
		return nil, fmt.Errorf("创建申报记录失败: %w", createErr)
	}
	return decl, nil
}

// ListDeclarations 查询申报列表
func (s *Service) ListDeclarations(orgID int64, year, page, pageSize int) ([]DeclarationResponse, int64, error) {
	if page == 0 {
		page = 1
	}
	if pageSize == 0 {
		pageSize = 20
	}

	declarations, total, err := s.repo.ListDeclarations(orgID, year, page, pageSize)
	if err != nil {
		return nil, 0, err
	}

	responses := make([]DeclarationResponse, 0, len(declarations))
	for _, d := range declarations {
		responses = append(responses, *s.toDeclarationResponse(&d))
	}
	return responses, total, nil
}

// MarkAsDeclared 标记为已申报
func (s *Service) MarkAsDeclared(orgID, userID, id int64) error {
	now := time.Now()
	updates := map[string]interface{}{
		"status":      DeclarationStatusDeclared,
		"declared_at": now,
		"declared_by": userID,
	}
	return s.repo.UpdateDeclaration(orgID, id, updates)
}

// ========== 内部转换方法 ==========

func (s *Service) toDeductionResponse(d *SpecialDeduction) *DeductionResponse {
	return &DeductionResponse{
		ID:             d.ID,
		EmployeeID:     d.EmployeeID,
		EmployeeName:   "", // 需要关联查询，由 API 层填充
		DeductionType:  d.DeductionType,
		MonthlyAmount:  d.MonthlyAmount,
		Count:          d.Count,
		EffectiveStart: d.EffectiveStart,
		EffectiveEnd:   d.EffectiveEnd,
		Remark:         d.Remark,
		CreatedAt:      d.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}

func (s *Service) toTaxRecordResponse(r *TaxRecord) *TaxRecordResponse {
	return &TaxRecordResponse{
		ID:                         r.ID,
		EmployeeID:                 r.EmployeeID,
		EmployeeName:               r.EmployeeName,
		Year:                       r.Year,
		Month:                      r.Month,
		GrossIncome:                r.GrossIncome,
		BasicDeduction:             r.BasicDeduction,
		SIDeduction:                r.SIDeduction,
		SpecialDeduction:           r.SpecialDeduction,
		TotalDeduction:             r.TotalDeduction,
		CumulativeIncome:           r.CumulativeIncome,
		CumulativeBasicDeduction:   r.CumulativeBasicDeduction,
		CumulativeSIDeduction:      r.CumulativeSIDeduction,
		CumulativeSpecialDeduction: r.CumulativeSpecialDeduction,
		CumulativeTaxableIncome:    r.CumulativeTaxableIncome,
		TaxRate:                    r.TaxRate,
		QuickDeduction:             r.QuickDeduction,
		CumulativeTax:              r.CumulativeTax,
		MonthlyTax:                 r.MonthlyTax,
		Source:                     r.Source,
		CreatedAt:                  r.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}

func (s *Service) toDeclarationResponse(d *TaxDeclaration) *DeclarationResponse {
	return &DeclarationResponse{
		ID:             d.ID,
		Year:           d.Year,
		Month:          d.Month,
		Status:         d.Status,
		TotalEmployees: d.TotalEmployees,
		TotalIncome:    d.TotalIncome,
		TotalTax:       d.TotalTax,
		DeclaredAt:     d.DeclaredAt,
		DeclaredBy:     d.DeclaredBy,
		CreatedAt:      d.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}
