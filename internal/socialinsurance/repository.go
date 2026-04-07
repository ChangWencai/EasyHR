package socialinsurance

import (
	"fmt"

	"github.com/wencai/easyhr/internal/common/middleware"
	"gorm.io/gorm"
)

// ErrPolicyNotFound 政策未找到
var ErrPolicyNotFound = fmt.Errorf("社保政策不存在")

// ErrRecordNotFound 参保记录未找到
var ErrRecordNotFound = fmt.Errorf("参保记录不存在")

// ErrRecordNotActive 参保记录非参保中状态
var ErrRecordNotActive = fmt.Errorf("参保记录非参保中状态")

// EmployeeInfo 员工信息（从 employee 模块查询）
type EmployeeInfo struct {
	ID     int64
	Name   string
	OrgID  int64
	UserID *int64
}

// EmployeeQuerier 员工查询接口（解耦社保和员工模块）
type EmployeeQuerier interface {
	FindEmployeeByIDs(orgID int64, ids []int64) ([]EmployeeInfo, error)
	FindEmployeeByUserID(orgID int64, userID int64) (*EmployeeInfo, error)
}

// Repository 社保数据访问层
type Repository struct {
	db *gorm.DB
}

// NewRepository 创建社保 Repository
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// ========== 政策 CRUD ==========

// Create 创建政策记录（全局数据，OrgID=0）
func (r *Repository) Create(policy *SocialInsurancePolicy) error {
	policy.OrgID = 0
	return r.db.Create(policy).Error
}

// FindByID 根据 ID 查找政策（全局数据，不使用 TenantScope）
func (r *Repository) FindByID(id int64) (*SocialInsurancePolicy, error) {
	var policy SocialInsurancePolicy
	err := r.db.Where("id = ? AND org_id = 0", id).First(&policy).Error
	if err != nil {
		return nil, err
	}
	return &policy, nil
}

// FindByCityAndYear 根据城市ID和年度查询最新有效政策
// 查询条件: city_id = ? AND effective_year <= ? AND org_id = 0
// 按 effective_year DESC 排序取第一条
func (r *Repository) FindByCityAndYear(cityID int, year int) (*SocialInsurancePolicy, error) {
	var policy SocialInsurancePolicy
	err := r.db.Where("city_id = ? AND effective_year <= ? AND org_id = 0", cityID, year).
		Order("effective_year DESC").
		First(&policy).Error
	if err != nil {
		return nil, err
	}
	return &policy, nil
}

// List 政策列表分页查询（支持按城市筛选）
func (r *Repository) List(cityID int, page, pageSize int) ([]SocialInsurancePolicy, int64, error) {
	var policies []SocialInsurancePolicy
	var total int64

	q := r.db.Model(&SocialInsurancePolicy{}).Where("org_id = 0")

	if cityID > 0 {
		q = q.Where("city_id = ?", cityID)
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count policies: %w", err)
	}

	offset := (page - 1) * pageSize
	if err := q.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&policies).Error; err != nil {
		return nil, 0, fmt.Errorf("list policies: %w", err)
	}

	return policies, total, nil
}

// Update 更新政策
func (r *Repository) Update(id int64, updates map[string]interface{}) error {
	result := r.db.Model(&SocialInsurancePolicy{}).Where("id = ? AND org_id = 0", id).Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrPolicyNotFound
	}
	return nil
}

// Delete 软删除政策
func (r *Repository) Delete(id int64) error {
	result := r.db.Where("id = ? AND org_id = 0", id).Delete(&SocialInsurancePolicy{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrPolicyNotFound
	}
	return nil
}

// ========== 参保记录 CRUD ==========

// CreateRecord 创建参保记录
func (r *Repository) CreateRecord(record *SocialInsuranceRecord) error {
	return r.db.Create(record).Error
}

// FindRecordByID 根据 ID 查找参保记录（带租户隔离）
func (r *Repository) FindRecordByID(orgID, id int64) (*SocialInsuranceRecord, error) {
	var record SocialInsuranceRecord
	err := r.db.Scopes(middleware.TenantScope(orgID)).Where("id = ?", id).First(&record).Error
	if err != nil {
		return nil, err
	}
	return &record, nil
}

// FindActiveRecordByEmployee 查询员工当前 active 参保记录
func (r *Repository) FindActiveRecordByEmployee(orgID, employeeID int64) (*SocialInsuranceRecord, error) {
	var record SocialInsuranceRecord
	err := r.db.Scopes(middleware.TenantScope(orgID)).
		Where("employee_id = ? AND status = ?", employeeID, SIStatusActive).
		First(&record).Error
	if err != nil {
		return nil, err
	}
	return &record, nil
}

// FindRecordsByEmployee 查询员工所有参保记录
func (r *Repository) FindRecordsByEmployee(orgID, employeeID int64) ([]SocialInsuranceRecord, error) {
	var records []SocialInsuranceRecord
	err := r.db.Scopes(middleware.TenantScope(orgID)).
		Where("employee_id = ?", employeeID).
		Order("created_at DESC").
		Find(&records).Error
	return records, err
}

// ListRecords 参保记录列表分页查询（支持按状态和员工姓名筛选）
func (r *Repository) ListRecords(orgID int64, status, employeeName string, page, pageSize int) ([]SocialInsuranceRecord, int64, error) {
	var records []SocialInsuranceRecord
	var total int64

	q := r.db.Model(&SocialInsuranceRecord{}).Scopes(middleware.TenantScope(orgID))

	if status != "" {
		q = q.Where("status = ?", status)
	}
	if employeeName != "" {
		q = q.Where("employee_name LIKE ?", "%"+employeeName+"%")
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count records: %w", err)
	}

	offset := (page - 1) * pageSize
	if err := q.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&records).Error; err != nil {
		return nil, 0, fmt.Errorf("list records: %w", err)
	}

	return records, total, nil
}

// UpdateRecord 更新参保记录
func (r *Repository) UpdateRecord(orgID, id int64, updates map[string]interface{}) error {
	result := r.db.Model(&SocialInsuranceRecord{}).Scopes(middleware.TenantScope(orgID)).
		Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrRecordNotFound
	}
	return nil
}

// ========== 变更历史 CRUD ==========

// CreateChangeHistory 创建变更历史记录
func (r *Repository) CreateChangeHistory(history *ChangeHistory) error {
	return r.db.Create(history).Error
}

// ListChangeHistories 查询变更历史列表（按员工筛选）
func (r *Repository) ListChangeHistories(orgID, employeeID int64, page, pageSize int) ([]ChangeHistory, int64, error) {
	var histories []ChangeHistory
	var total int64

	q := r.db.Model(&ChangeHistory{}).Scopes(middleware.TenantScope(orgID))

	if employeeID > 0 {
		q = q.Where("employee_id = ?", employeeID)
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count histories: %w", err)
	}

	offset := (page - 1) * pageSize
	if err := q.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&histories).Error; err != nil {
		return nil, 0, fmt.Errorf("list histories: %w", err)
	}

	return histories, total, nil
}
