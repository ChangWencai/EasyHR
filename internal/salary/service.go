package salary

import (
	"fmt"

	"gorm.io/gorm"
)

// Service 工资核算业务逻辑层
type Service struct {
	repo              *Repository
	templateRepo      *SalaryTemplateRepository
	taxProvider       TaxProvider
	siProvider        SIDeductionProvider
	empProvider       EmployeeProvider
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
		repo:              repo,
		templateRepo:      templateRepo,
		taxProvider:       taxProvider,
		siProvider:        siProvider,
		empProvider:       empProvider,
		baseAdjustProvider: baseAdjustProvider,
	}
}

// SeedTemplateItems 初始化预置薪资项模板
func (s *Service) SeedTemplateItems() error {
	return s.templateRepo.SeedGlobalTemplateItems()
}

// GetTemplate 获取企业薪资模板（含启用状态）
func (s *Service) GetTemplate(orgID int64) (*TemplateResponse, error) {
	// 获取全局预置项
	globalItems, err := s.templateRepo.GetGlobalItems()
	if err != nil {
		return nil, fmt.Errorf("获取全局模板失败: %w", err)
	}

	// 获取企业级覆盖
	overrides, err := s.templateRepo.GetOrgOverrides(orgID)
	if err != nil {
		return nil, fmt.Errorf("获取企业配置失败: %w", err)
	}

	// 构建覆盖映射：name -> isEnabled
	overrideMap := make(map[string]bool)
	for _, o := range overrides {
		overrideMap[o.Name] = o.IsEnabled
	}

	// 合并：全局模板 + 企业覆盖
	items := make([]TemplateItemResponse, 0, len(globalItems))
	for _, g := range globalItems {
		isEnabled := g.IsEnabled // 默认使用全局状态
		if overridden, ok := overrideMap[g.Name]; ok {
			isEnabled = overridden // 企业覆盖优先
		}
		items = append(items, TemplateItemResponse{
			ID:        g.ID,
			Name:      g.Name,
			Type:      g.Type,
			SortOrder: g.SortOrder,
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
	// 获取薪资项金额
	salaryItems, err := s.repo.FindSalaryItemsByEmployee(orgID, employeeID, month)
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("查询员工薪资项失败: %w", err)
	}

	// 获取模板信息用于显示名称
	template, err := s.GetTemplate(orgID)
	if err != nil {
		return nil, err
	}

	// 构建模板映射：id -> template
	templateMap := make(map[int64]TemplateItemResponse)
	for _, t := range template.Items {
		templateMap[t.ID] = t
	}

	// 构建薪资项映射：template_item_id -> amount
	itemMap := make(map[int64]float64)
	for _, si := range salaryItems {
		itemMap[si.TemplateItemID] = si.Amount
	}

	// 构建响应：所有启用的模板项 + 对应金额
	items := make([]EmployeeItemResponse, 0)
	for _, t := range template.Items {
		if !t.IsEnabled {
			continue
		}
		amount := itemMap[t.ID] // 未设置时默认为 0
		items = append(items, EmployeeItemResponse{
			TemplateItemID: t.ID,
			ItemName:       t.Name,
			ItemType:       t.Type,
			Amount:         amount,
		})
	}

	return &EmployeeItemsResponse{
		EmployeeID: employeeID,
		Month:      month,
		Items:      items,
	}, nil
}

// SetEmployeeItems 设置员工各项金额
func (s *Service) SetEmployeeItems(orgID, userID, employeeID int64, month string, items []SalaryItemInput) error {
	// 验证月份格式
	if len(month) != 7 || month[4] != '-' {
		return fmt.Errorf("月份格式错误，应为 YYYY-MM")
	}

	// 获取企业启用的模板项
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
