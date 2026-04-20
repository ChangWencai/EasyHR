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
	// 先获取全局模板项
	var globalItem SalaryTemplateItem
	if err := r.db.Where("id = ? AND org_id = 0", templateItemID).First(&globalItem).Error; err != nil {
		return fmt.Errorf("global template item not found: %w", err)
	}
	fmt.Printf("Global item: ID=%d, Name=%s\n", globalItem.ID, globalItem.Name)

	// 查找是否已有企业级覆盖（通过 name 匹配）
	var existing SalaryTemplateItem
	err := r.db.Scopes(middleware.TenantScope(orgID)).
		Where("name = ?", globalItem.Name).
		First(&existing).Error
	fmt.Printf("Find existing err: %v\n", err)

	if err == gorm.ErrRecordNotFound {
		fmt.Printf("Creating new override for orgID=%d, name=%s, enabled=%v\n", orgID, globalItem.Name, isEnabled)
		// 创建新的企业级覆盖
		override := SalaryTemplateItem{}
		override.OrgID = orgID
		override.CreatedBy = userID
		override.UpdatedBy = userID
		override.Name = globalItem.Name
		override.Type = globalItem.Type
		override.SortOrder = globalItem.SortOrder
		override.IsRequired = globalItem.IsRequired
		override.IsEnabled = isEnabled
		if err := r.db.Debug().Create(&override).Error; err != nil {
			return fmt.Errorf("create override: %w", err)
		}
		fmt.Printf("Created override: ID=%d, IsEnabled=%v\n", override.ID, override.IsEnabled)
		return nil
	}
	if err != nil {
		return fmt.Errorf("find override: %w", err)
	}

	fmt.Printf("Updating existing override: ID=%d, current IsEnabled=%v, new IsEnabled=%v\n", existing.ID, existing.IsEnabled, isEnabled)
	// 更新现有覆盖
	result := r.db.Debug().Model(&SalaryTemplateItem{}).
		Scopes(middleware.TenantScope(orgID)).
		Where("id = ?", existing.ID).
		Update("is_enabled", isEnabled)
	fmt.Printf("Update result: RowsAffected=%d, Error=%v\n", result.RowsAffected, result.Error)
	if result.Error != nil {
		return fmt.Errorf("update override: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("no rows affected when updating override (id=%d, enabled=%v)", existing.ID, isEnabled)
	}
	return nil
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

// FindPayrollRecordsByMonth 查询指定月份所有记录
func (r *Repository) FindPayrollRecordsByMonth(orgID int64, year, month int) ([]PayrollRecord, error) {
	var records []PayrollRecord
	err := r.db.Scopes(middleware.TenantScope(orgID)).
		Where("year = ? AND month = ?", year, month).
		Find(&records).Error
	return records, err
}

// FindPayrollRecordByEmployeeMonth 查询员工某月的工资记录
func (r *Repository) FindPayrollRecordByEmployeeMonth(orgID, employeeID int64, year, month int) (*PayrollRecord, error) {
	var record PayrollRecord
	err := r.db.Scopes(middleware.TenantScope(orgID)).
		Where("employee_id = ? AND year = ? AND month = ?", employeeID, year, month).
		First(&record).Error
	if err != nil {
		return nil, err
	}
	return &record, nil
}

// FindPreviousMonthRecords 查询上月记录（用于异常检查）
func (r *Repository) FindPreviousMonthRecords(orgID int64, year, month int) ([]PayrollRecord, error) {
	prevYear, prevMonth := year, month-1
	if prevMonth == 0 {
		prevYear--
		prevMonth = 12
	}
	var records []PayrollRecord
	err := r.db.Scopes(middleware.TenantScope(orgID)).
		Where("year = ? AND month = ? AND status IN ?", prevYear, prevMonth,
			[]string{PayrollStatusConfirmed, PayrollStatusPaid}).
		Find(&records).Error
	return records, err
}

// ListPayrollRecords 分页查询工资核算记录
func (r *Repository) ListPayrollRecords(orgID int64, year, month, page, pageSize int) ([]PayrollRecord, int64, error) {
	var records []PayrollRecord
	var total int64

	q := r.db.Model(&PayrollRecord{}).Scopes(middleware.TenantScope(orgID)).
		Where("year = ? AND month = ?", year, month)

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count payroll records: %w", err)
	}

	offset := (page - 1) * pageSize
	if err := q.Offset(offset).Limit(pageSize).Order("employee_name").Find(&records).Error; err != nil {
		return nil, 0, fmt.Errorf("list payroll records: %w", err)
	}

	return records, total, nil
}

// BatchCreatePayrollItems 批量创建工资明细
func (r *Repository) BatchCreatePayrollItems(orgID int64, items []PayrollItem) error {
	if len(items) == 0 {
		return nil
	}
	for i := range items {
		items[i].OrgID = orgID
	}
	return r.db.Create(&items).Error
}

// ========== SalarySlipSendLog CRUD ==========

// CreateSlipSendLog 创建工资条发送日志
func (r *Repository) CreateSlipSendLog(log *SalarySlipSendLog) error {
	return r.db.Create(log).Error
}

// FindSlipSendLogs 查询工资条发送日志列表（含 payroll_slips 联查确认状态）
func (r *Repository) FindSlipSendLogs(orgID int64, recordIDs []int64) ([]SalarySlipSendLog, error) {
	var logs []SalarySlipSendLog
	q := r.db.Scopes(middleware.TenantScope(orgID)).
		Joins("LEFT JOIN payroll_slips ON payroll_slips.id = salary_slip_send_logs.payroll_record_id").
		Select("salary_slip_send_logs.*, payroll_slips.confirmed_at AS confirmed_at").
		Order("salary_slip_send_logs.created_at DESC")
	if len(recordIDs) > 0 {
		q = q.Where("salary_slip_send_logs.payroll_record_id IN ?", recordIDs)
	}
	err := q.Find(&logs).Error
	return logs, err
}

// FindSlipSendLogByRecord 查询指定工资记录的发送日志
func (r *Repository) FindSlipSendLogByRecord(orgID int64, recordID int64) ([]SalarySlipSendLog, error) {
	var logs []SalarySlipSendLog
	err := r.db.Scopes(middleware.TenantScope(orgID)).
		Where("payroll_record_id = ?", recordID).
		Order("created_at DESC").
		Find(&logs).Error
	return logs, err
}

// UpdateSlipSendLog 更新工资条发送日志
func (r *Repository) UpdateSlipSendLog(orgID int64, id int64, updates map[string]interface{}) error {
	result := r.db.Model(&SalarySlipSendLog{}).
		Scopes(middleware.TenantScope(orgID)).
		Where("id = ?", id).
		Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// FindSlipSendLogByEmployeeMonth 查询员工某月的发送日志
func (r *Repository) FindSlipSendLogByEmployeeMonth(orgID int64, employeeID int64, year, month int) ([]SalarySlipSendLog, error) {
	var logs []SalarySlipSendLog
	// 通过 payroll_record_id 关联查询
	err := r.db.Scopes(middleware.TenantScope(orgID)).
		Joins("JOIN payroll_records ON payroll_records.id = salary_slip_send_logs.payroll_record_id").
		Where("salary_slip_send_logs.employee_id = ? AND payroll_records.year = ? AND payroll_records.month = ?", employeeID, year, month).
		Order("salary_slip_send_logs.created_at DESC").
		Find(&logs).Error
	return logs, err
}

// ListUnconfirmedSlipsByMonth 查询指定月份未确认的工资条（D-13-08）
func (r *Repository) ListUnconfirmedSlipsByMonth(orgID int64, year, month int) ([]PayrollSlip, error) {
	var slips []PayrollSlip
	err := r.db.Table("payroll_slips ps").
		Joins("JOIN payroll_records pr ON pr.id = ps.payroll_record_id").
		Where("ps.org_id = ? AND pr.year = ? AND pr.month = ? AND ps.confirmed_at IS NULL AND ps.deleted_at IS NULL", orgID, year, month).
		Find(&slips).Error
	return slips, err
}

// FindOwnerUserIDByOrg 查询企业的所有者用户ID
func (r *Repository) FindOwnerUserIDByOrg(orgID int64) (int64, error) {
	var userID int64
	err := r.db.Table("users").
		Select("id").
		Where("org_id = ? AND role = 'owner' AND deleted_at IS NULL", orgID).
		Scan(&userID).Error
	if err != nil {
		return 0, err
	}
	return userID, nil
}
