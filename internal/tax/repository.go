package tax

import (
	"fmt"

	"github.com/wencai/easyhr/internal/common/middleware"
	"gorm.io/gorm"
)

// Repository 个税数据访问层
type Repository struct {
	db *gorm.DB
}

// NewRepository 创建个税 Repository
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// ========== 税率表方法（全局数据, OrgID=0）==========

// SeedTaxBrackets 插入7级税率种子数据
// 先检查是否已有该年份数据，避免重复插入
func (r *Repository) SeedTaxBrackets(year int) error {
	var count int64
	r.db.Model(&TaxBracket{}).Where("org_id = 0 AND effective_year = ?", year).Count(&count)
	if count > 0 {
		return nil // 已有数据，跳过
	}

	brackets := []TaxBracket{
		{Level: 1, LowerBound: 0, UpperBound: 36000, Rate: 0.03, QuickDeduction: 0, EffectiveYear: year},
		{Level: 2, LowerBound: 36000, UpperBound: 144000, Rate: 0.10, QuickDeduction: 2520, EffectiveYear: year},
		{Level: 3, LowerBound: 144000, UpperBound: 300000, Rate: 0.20, QuickDeduction: 16920, EffectiveYear: year},
		{Level: 4, LowerBound: 300000, UpperBound: 420000, Rate: 0.25, QuickDeduction: 31920, EffectiveYear: year},
		{Level: 5, LowerBound: 420000, UpperBound: 660000, Rate: 0.30, QuickDeduction: 52920, EffectiveYear: year},
		{Level: 6, LowerBound: 660000, UpperBound: 960000, Rate: 0.35, QuickDeduction: 85920, EffectiveYear: year},
		{Level: 7, LowerBound: 960000, UpperBound: 0, Rate: 0.45, QuickDeduction: 181920, EffectiveYear: year},
	}

	for i := range brackets {
		brackets[i].OrgID = 0
	}

	return r.db.Create(&brackets).Error
}

// FindTaxBrackets 查询指定年份全部税率，按 level ASC 排序
func (r *Repository) FindTaxBrackets(year int) ([]TaxBracket, error) {
	var brackets []TaxBracket
	err := r.db.Where("org_id = 0 AND effective_year = ?", year).
		Order("level ASC").
		Find(&brackets).Error
	if err != nil {
		return nil, fmt.Errorf("find tax brackets: %w", err)
	}
	if len(brackets) == 0 {
		return nil, ErrBracketNotFound
	}
	return brackets, nil
}

// FindTaxBracketForAmount 纯函数，从税率列表中找到 amount 所在区间
// 边界值规则: lower_bound <= amount < upper_bound (top bracket: amount >= lower_bound)
func FindTaxBracketForAmount(brackets []TaxBracket, amount float64) *TaxBracket {
	for i := range brackets {
		b := &brackets[i]
		// 顶级（UpperBound=0 表示无上限）
		if b.UpperBound == 0 {
			if amount >= b.LowerBound {
				return b
			}
			continue
		}
		// 普通级别: lower_bound <= amount < upper_bound
		if amount >= b.LowerBound && amount < b.UpperBound {
			return b
		}
	}
	// 理论上不应到达此处（7级税率覆盖所有正数），但安全起见返回最高级
	if len(brackets) > 0 {
		return &brackets[len(brackets)-1]
	}
	return nil
}

// ========== 专项附加扣除方法（租户隔离）==========

// CreateDeduction 创建专项附加扣除记录
func (r *Repository) CreateDeduction(orgID int64, deduction *SpecialDeduction) error {
	return r.db.Scopes(middleware.TenantScope(orgID)).Create(deduction).Error
}

// FindDeductionByID 根据 ID 查找扣除记录
func (r *Repository) FindDeductionByID(orgID, id int64) (*SpecialDeduction, error) {
	var deduction SpecialDeduction
	err := r.db.Scopes(middleware.TenantScope(orgID)).
		Where("id = ?", id).First(&deduction).Error
	if err != nil {
		return nil, err
	}
	return &deduction, nil
}

// FindDeductionByEmployeeAndType 查找同类型有效扣除（去重检查）
func (r *Repository) FindDeductionByEmployeeAndType(orgID, employeeID int64, deductionType string) (*SpecialDeduction, error) {
	var deduction SpecialDeduction
	err := r.db.Scopes(middleware.TenantScope(orgID)).
		Where("employee_id = ? AND deduction_type = ? AND effective_end IS NULL", employeeID, deductionType).
		First(&deduction).Error
	if err != nil {
		return nil, err
	}
	return &deduction, nil
}

// ListDeductionsByEmployee 查询当月生效的扣除项（分页）
func (r *Repository) ListDeductionsByEmployee(orgID, employeeID int64, month string, page, pageSize int) ([]SpecialDeduction, int64, error) {
	var deductions []SpecialDeduction
	var total int64

	q := r.db.Model(&SpecialDeduction{}).Scopes(middleware.TenantScope(orgID)).
		Where("employee_id = ? AND effective_start <= ? AND (effective_end IS NULL OR effective_end >= ?)",
			employeeID, month, month)

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count deductions: %w", err)
	}

	offset := (page - 1) * pageSize
	if err := q.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&deductions).Error; err != nil {
		return nil, 0, fmt.Errorf("list deductions: %w", err)
	}

	return deductions, total, nil
}

// ListAllActiveDeductionsByEmployee 获取当月全部有效扣除（计算引擎用，不分页）
func (r *Repository) ListAllActiveDeductionsByEmployee(orgID, employeeID int64, month string) ([]SpecialDeduction, error) {
	var deductions []SpecialDeduction
	err := r.db.Scopes(middleware.TenantScope(orgID)).
		Where("employee_id = ? AND effective_start <= ? AND (effective_end IS NULL OR effective_end >= ?)",
			employeeID, month, month).
		Find(&deductions).Error
	if err != nil {
		return nil, fmt.Errorf("list active deductions: %w", err)
	}
	return deductions, nil
}

// UpdateDeduction 更新扣除记录
func (r *Repository) UpdateDeduction(orgID, id int64, updates map[string]interface{}) error {
	result := r.db.Model(&SpecialDeduction{}).Scopes(middleware.TenantScope(orgID)).
		Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrDeductionNotFound
	}
	return nil
}

// DeleteDeduction 软删除扣除记录
func (r *Repository) DeleteDeduction(orgID, id int64) error {
	result := r.db.Scopes(middleware.TenantScope(orgID)).
		Where("id = ?", id).Delete(&SpecialDeduction{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrDeductionNotFound
	}
	return nil
}

// ========== TaxRecord 方法（租户隔离）==========

// CreateTaxRecord 创建个税计算记录
func (r *Repository) CreateTaxRecord(record *TaxRecord) error {
	return r.db.Create(record).Error
}

// FindTaxRecordByEmployeeMonth 查找指定月份记录
func (r *Repository) FindTaxRecordByEmployeeMonth(orgID, employeeID int64, year, month int) (*TaxRecord, error) {
	var record TaxRecord
	err := r.db.Scopes(middleware.TenantScope(orgID)).
		Where("employee_id = ? AND year = ? AND month = ?", employeeID, year, month).
		First(&record).Error
	if err != nil {
		return nil, err
	}
	return &record, nil
}

// FindTaxRecordsByEmployeeYear 查询指定年份全部记录（用于累计计算）
func (r *Repository) FindTaxRecordsByEmployeeYear(orgID, employeeID int64, year int) ([]TaxRecord, error) {
	var records []TaxRecord
	err := r.db.Scopes(middleware.TenantScope(orgID)).
		Where("employee_id = ? AND year = ?", employeeID, year).
		Order("month ASC").
		Find(&records).Error
	if err != nil {
		return nil, fmt.Errorf("find tax records by employee year: %w", err)
	}
	return records, nil
}

// ListTaxRecords 分页查询个税记录
func (r *Repository) ListTaxRecords(orgID int64, params TaxRecordListQuery) ([]TaxRecord, int64, error) {
	var records []TaxRecord
	var total int64

	q := r.db.Model(&TaxRecord{}).Scopes(middleware.TenantScope(orgID))

	if params.EmployeeID > 0 {
		q = q.Where("employee_id = ?", params.EmployeeID)
	}
	if params.Year > 0 {
		q = q.Where("year = ?", params.Year)
	}
	if params.Month > 0 {
		q = q.Where("month = ?", params.Month)
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count tax records: %w", err)
	}

	page := params.Page
	pageSize := params.PageSize
	if page == 0 {
		page = 1
	}
	if pageSize == 0 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize
	if err := q.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&records).Error; err != nil {
		return nil, 0, fmt.Errorf("list tax records: %w", err)
	}

	return records, total, nil
}

// FindTaxRecordByID 根据 ID 查找个税记录
func (r *Repository) FindTaxRecordByID(orgID, id int64) (*TaxRecord, error) {
	var record TaxRecord
	err := r.db.Scopes(middleware.TenantScope(orgID)).
		Where("id = ?", id).First(&record).Error
	if err != nil {
		return nil, err
	}
	return &record, nil
}

// ========== TaxDeclaration 方法（租户隔离）==========

// CreateDeclaration 创建申报记录
func (r *Repository) CreateDeclaration(decl *TaxDeclaration) error {
	return r.db.Create(decl).Error
}

// FindDeclarationByMonth 查找指定月份的申报记录
func (r *Repository) FindDeclarationByMonth(orgID int64, year, month int) (*TaxDeclaration, error) {
	var decl TaxDeclaration
	err := r.db.Scopes(middleware.TenantScope(orgID)).
		Where("year = ? AND month = ?", year, month).
		First(&decl).Error
	if err != nil {
		return nil, err
	}
	return &decl, nil
}

// UpdateDeclaration 更新申报记录
func (r *Repository) UpdateDeclaration(orgID, id int64, updates map[string]interface{}) error {
	result := r.db.Model(&TaxDeclaration{}).Scopes(middleware.TenantScope(orgID)).
		Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrDeclarationNotFound
	}
	return nil
}

// ListDeclarations 分页查询申报记录
func (r *Repository) ListDeclarations(orgID int64, year int, page, pageSize int) ([]TaxDeclaration, int64, error) {
	var declarations []TaxDeclaration
	var total int64

	q := r.db.Model(&TaxDeclaration{}).Scopes(middleware.TenantScope(orgID))

	if year > 0 {
		q = q.Where("year = ?", year)
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count declarations: %w", err)
	}

	offset := (page - 1) * pageSize
	if err := q.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&declarations).Error; err != nil {
		return nil, 0, fmt.Errorf("list declarations: %w", err)
	}

	return declarations, total, nil
}

// ========== TaxReminder 方法（租户隔离）==========

// CreateReminder 创建提醒记录
func (r *Repository) CreateReminder(reminder *TaxReminder) error {
	return r.db.Create(reminder).Error
}

// ListReminders 分页查询提醒列表
func (r *Repository) ListReminders(orgID int64, page, pageSize int) ([]TaxReminder, int64, error) {
	var reminders []TaxReminder
	var total int64

	q := r.db.Model(&TaxReminder{}).Scopes(middleware.TenantScope(orgID)).
		Where("is_dismissed = false")

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count reminders: %w", err)
	}

	offset := (page - 1) * pageSize
	if err := q.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&reminders).Error; err != nil {
		return nil, 0, fmt.Errorf("list reminders: %w", err)
	}

	return reminders, total, nil
}

// DismissReminder 关闭提醒
func (r *Repository) DismissReminder(orgID, id int64) error {
	result := r.db.Model(&TaxReminder{}).Scopes(middleware.TenantScope(orgID)).
		Where("id = ?", id).Updates(map[string]interface{}{"is_dismissed": true})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrDeclarationNotFound
	}
	return nil
}

// FindReminderByMonth 查找指定月份的提醒记录（去重用）
func (r *Repository) FindReminderByMonth(orgID int64, year, month int) (*TaxReminder, error) {
	var reminder TaxReminder
	err := r.db.Scopes(middleware.TenantScope(orgID)).
		Where("type = ? AND year = ? AND month = ?", ReminderTypeDeclarationDue, year, month).
		First(&reminder).Error
	if err != nil {
		return nil, err
	}
	return &reminder, nil
}

// FindAllOrgIDs 查询所有有税务申报记录的企业ID
func (r *Repository) FindAllOrgIDs() ([]int64, error) {
	var orgIDs []int64
	err := r.db.Model(&TaxDeclaration{}).
		Distinct("org_id").
		Pluck("org_id", &orgIDs).Error
	if err != nil {
		return nil, fmt.Errorf("find all org IDs: %w", err)
	}
	return orgIDs, nil
}

// CountTaxRecordsByOrgMonth 统计指定企业月份的个税记录数和总税额
func (r *Repository) CountTaxRecordsByOrgMonth(orgID int64, year, month int) (count int64, totalTax float64, err error) {
	err = r.db.Model(&TaxRecord{}).Scopes(middleware.TenantScope(orgID)).
		Where("year = ? AND month = ?", year, month).
		Count(&count).Error
	if err != nil {
		return 0, 0, fmt.Errorf("count tax records: %w", err)
	}

	if count == 0 {
		return 0, 0, nil
	}

	err = r.db.Model(&TaxRecord{}).Scopes(middleware.TenantScope(orgID)).
		Where("year = ? AND month = ?", year, month).
		Select("COALESCE(SUM(monthly_tax), 0)").
		Row().Scan(&totalTax)
	if err != nil {
		return count, 0, fmt.Errorf("sum tax records: %w", err)
	}

	return count, totalTax, nil
}

// FindAllTaxRecordsByOrgMonth 查询指定企业月份的全部个税记录（导出用，不分页，限1000条）
func (r *Repository) FindAllTaxRecordsByOrgMonth(orgID int64, year, month int) ([]TaxRecord, error) {
	var records []TaxRecord
	err := r.db.Scopes(middleware.TenantScope(orgID)).
		Where("year = ? AND month = ?", year, month).
		Order("employee_id ASC").
		Limit(1000).
		Find(&records).Error
	if err != nil {
		return nil, fmt.Errorf("find all tax records by org month: %w", err)
	}
	return records, nil
}

// GetOrgName 获取企业名称
func (r *Repository) GetOrgName(orgID int64) (string, error) {
	var name string
	err := r.db.Table("organizations").Where("id = ?", orgID).Select("name").Row().Scan(&name)
	if err != nil {
		return "", fmt.Errorf("get org name: %w", err)
	}
	return name, nil
}
