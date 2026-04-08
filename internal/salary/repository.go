package salary

import (
	"fmt"

	"github.com/wencai/easyhr/internal/common/middleware"
	"gorm.io/gorm"
)

// ========== SalaryTemplateRepository 薪资模板仓储 ==========

// SalaryTemplateRepository 薪资模板数据访问层
type SalaryTemplateRepository struct {
	db *gorm.DB
}

// NewSalaryTemplateRepository 创建薪资模板 Repository
func NewSalaryTemplateRepository(db *gorm.DB) *SalaryTemplateRepository {
	return &SalaryTemplateRepository{db: db}
}

// SeedGlobalTemplateItems 创建全局预置薪资项（OrgID=0）
func (r *SalaryTemplateRepository) SeedGlobalTemplateItems() error {
	presets := []SalaryTemplateItem{
		{Name: "基本工资", Type: "income", SortOrder: 1, IsRequired: true, IsEnabled: true},
		{Name: "绩效工资", Type: "income", SortOrder: 2, IsRequired: false, IsEnabled: true},
		{Name: "岗位补贴", Type: "income", SortOrder: 3, IsRequired: false, IsEnabled: true},
		{Name: "餐补", Type: "income", SortOrder: 4, IsRequired: false, IsEnabled: true},
		{Name: "交通补", Type: "income", SortOrder: 5, IsRequired: false, IsEnabled: true},
		{Name: "通讯补", Type: "income", SortOrder: 6, IsRequired: false, IsEnabled: true},
		{Name: "其他补贴", Type: "income", SortOrder: 7, IsRequired: false, IsEnabled: true},
		{Name: "事假扣款", Type: "deduction", SortOrder: 8, IsRequired: false, IsEnabled: true},
		{Name: "病假扣款", Type: "deduction", SortOrder: 9, IsRequired: false, IsEnabled: true},
		{Name: "其他扣款", Type: "deduction", SortOrder: 10, IsRequired: false, IsEnabled: true},
	}

	for i := range presets {
		presets[i].OrgID = 0 // 全局预置
		// 仅创建不存在的项
		var count int64
		r.db.Model(&SalaryTemplateItem{}).
			Where("org_id = 0 AND name = ?", presets[i].Name).
			Count(&count)
		if count == 0 {
			if err := r.db.Create(&presets[i]).Error; err != nil {
				return fmt.Errorf("seed template item %s: %w", presets[i].Name, err)
			}
		}
	}
	return nil
}

// GetGlobalItems 获取全局预置薪资项
func (r *SalaryTemplateRepository) GetGlobalItems() ([]SalaryTemplateItem, error) {
	var items []SalaryTemplateItem
	err := r.db.Where("org_id = 0").Order("sort_order").Find(&items).Error
	return items, err
}

// GetOrgOverrides 获取企业级覆盖配置
func (r *SalaryTemplateRepository) GetOrgOverrides(orgID int64) ([]SalaryTemplateItem, error) {
	var items []SalaryTemplateItem
	err := r.db.Scopes(middleware.TenantScope(orgID)).Find(&items).Error
	return items, err
}

// UpsertOrgOverride 创建或更新企业级覆盖
func (r *SalaryTemplateRepository) UpsertOrgOverride(orgID, userID, templateItemID int64, isEnabled bool) error {
	var existing SalaryTemplateItem
	err := r.db.Scopes(middleware.TenantScope(orgID)).
		Where("template_item_id = ? OR name = (SELECT name FROM salary_template_items WHERE id = ?)", templateItemID, templateItemID).
		First(&existing).Error

	if err == gorm.ErrRecordNotFound {
		// 获取全局模板名称
		var globalItem SalaryTemplateItem
		if err := r.db.Where("id = ? AND org_id = 0", templateItemID).First(&globalItem).Error; err != nil {
			return fmt.Errorf("global template item not found: %w", err)
		}
		override := SalaryTemplateItem{
			// BaseModel 中 OrgID 会被 GORM 自动填充
		}
		override.OrgID = orgID
		override.CreatedBy = userID
		override.UpdatedBy = userID
		override.Name = globalItem.Name
		override.Type = globalItem.Type
		override.SortOrder = globalItem.SortOrder
		override.IsRequired = globalItem.IsRequired
		override.IsEnabled = isEnabled
		return r.db.Create(&override).Error
	}
	if err != nil {
		return err
	}

	return r.db.Model(&existing).Updates(map[string]interface{}{
		"is_enabled": isEnabled,
		"updated_by": userID,
	}).Error
}

// ========== Repository 主仓储 ==========

// Repository 工资核算数据访问层
type Repository struct {
	db *gorm.DB
}

// NewRepository 创建工资核算 Repository
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// ========== SalaryItem CRUD ==========

// FindSalaryItemsByEmployee 查询员工某月薪资项
func (r *Repository) FindSalaryItemsByEmployee(orgID, employeeID int64, month string) ([]SalaryItem, error) {
	var items []SalaryItem
	err := r.db.Scopes(middleware.TenantScope(orgID)).
		Where("employee_id = ? AND effective_month = ?", employeeID, month).
		Find(&items).Error
	return items, err
}

// UpsertSalaryItem 创建或更新员工薪资项
func (r *Repository) UpsertSalaryItem(orgID, userID, employeeID, templateItemID int64, month string, amount float64) error {
	var existing SalaryItem
	err := r.db.Scopes(middleware.TenantScope(orgID)).
		Where("employee_id = ? AND template_item_id = ? AND effective_month = ?", employeeID, templateItemID, month).
		First(&existing).Error

	if err == gorm.ErrRecordNotFound {
		item := SalaryItem{}
		item.OrgID = orgID
		item.CreatedBy = userID
		item.UpdatedBy = userID
		item.EmployeeID = employeeID
		item.TemplateItemID = templateItemID
		item.Amount = amount
		item.EffectiveMonth = month
		return r.db.Create(&item).Error
	}
	if err != nil {
		return err
	}

	return r.db.Model(&existing).Updates(map[string]interface{}{
		"amount":      amount,
		"updated_by": userID,
	}).Error
}

// ========== PayrollRecord CRUD ==========

// CreatePayrollRecord 创建工资核算记录
func (r *Repository) CreatePayrollRecord(record *PayrollRecord) error {
	return r.db.Create(record).Error
}

// FindPayrollRecordByID 根据 ID 查询工资核算记录
func (r *Repository) FindPayrollRecordByID(orgID, id int64) (*PayrollRecord, error) {
	var record PayrollRecord
	err := r.db.Scopes(middleware.TenantScope(orgID)).Where("id = ?", id).First(&record).Error
	if err != nil {
		return nil, err
	}
	return &record, nil
}

// UpdatePayrollRecord 更新工资核算记录
func (r *Repository) UpdatePayrollRecord(orgID int64, record *PayrollRecord) error {
	result := r.db.Model(&PayrollRecord{}).Scopes(middleware.TenantScope(orgID)).
		Where("id = ?", record.ID).Updates(record)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// ========== PayrollItem CRUD ==========

// CreatePayrollItems 批量创建工资核算明细
func (r *Repository) CreatePayrollItems(items []PayrollItem) error {
	if len(items) == 0 {
		return nil
	}
	return r.db.Create(&items).Error
}

// FindPayrollItemsByRecord 查询工资核算明细
func (r *Repository) FindPayrollItemsByRecord(orgID, recordID int64) ([]PayrollItem, error) {
	var items []PayrollItem
	err := r.db.Scopes(middleware.TenantScope(orgID)).
		Where("payroll_record_id = ?", recordID).
		Find(&items).Error
	return items, err
}

// DeletePayrollItemsByRecord 删除工资核算明细（重算时使用）
func (r *Repository) DeletePayrollItemsByRecord(orgID, recordID int64) error {
	return r.db.Scopes(middleware.TenantScope(orgID)).
		Where("payroll_record_id = ?", recordID).
		Delete(&PayrollItem{}).Error
}

// ========== PayrollSlip CRUD ==========

// CreateSlip 创建工资单
func (r *Repository) CreateSlip(slip *PayrollSlip) error {
	return r.db.Create(slip).Error
}

// FindSlipByToken 通过 token 查询工资单
func (r *Repository) FindSlipByToken(token string) (*PayrollSlip, error) {
	var slip PayrollSlip
	err := r.db.Where("token = ?", token).First(&slip).Error
	if err != nil {
		return nil, err
	}
	return &slip, nil
}

// UpdateSlip 更新工资单
func (r *Repository) UpdateSlip(orgID, id int64, updates map[string]interface{}) error {
	result := r.db.Model(&PayrollSlip{}).Scopes(middleware.TenantScope(orgID)).
		Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
