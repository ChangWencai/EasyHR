package socialinsurance

import (
	"context"
	"fmt"
	"time"

	"github.com/shopspring/decimal"
	"github.com/wencai/easyhr/internal/common/middleware"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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
// 查询条件: city_code = ? AND effective_year <= ? AND org_id = 0
// 按 effective_year DESC 排序取第一条
func (r *Repository) FindByCityAndYear(cityCode int64, year int) (*SocialInsurancePolicy, error) {
	var policy SocialInsurancePolicy
	err := r.db.Where("city_code = ? AND effective_year <= ? AND org_id = 0", cityCode, year).
		Order("effective_year DESC").
		First(&policy).Error
	if err != nil {
		return nil, err
	}
	return &policy, nil
}

// List 政策列表分页查询（支持按城市筛选）
func (r *Repository) List(cityCode int64, page, pageSize int) ([]SocialInsurancePolicy, int64, error) {
	var policies []SocialInsurancePolicy
	var total int64

	q := r.db.Model(&SocialInsurancePolicy{}).Where("org_id = 0")

	if cityCode > 0 {
		q = q.Where("city_code = ?", cityCode)
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

// ========== 月度缴费记录 CRUD ==========

// SIMonthlyPaymentRepository 月度缴费记录数据访问层
type SIMonthlyPaymentRepository struct {
	db *gorm.DB
}

// NewMonthlyPaymentRepository 创建月度缴费 Repository
func NewMonthlyPaymentRepository(db *gorm.DB) *SIMonthlyPaymentRepository {
	return &SIMonthlyPaymentRepository{db: db}
}

// Create 创建月度缴费记录
func (r *SIMonthlyPaymentRepository) Create(ctx context.Context, tx *gorm.DB, payment *SIMonthlyPayment) error {
	db := r.getDB(tx)
	return db.WithContext(ctx).Create(payment).Error
}

// GetByOrgAndEmployee 根据组织、员工和年月查询月度缴费记录
func (r *SIMonthlyPaymentRepository) GetByOrgAndEmployee(ctx context.Context, orgID int64, employeeID uint, yearMonth string) (*SIMonthlyPayment, error) {
	var payment SIMonthlyPayment
	err := r.db.WithContext(ctx).
		Scopes(middleware.TenantScope(orgID)).
		Where("employee_id = ? AND year_month = ?", employeeID, yearMonth).
		First(&payment).Error
	if err != nil {
		return nil, err
	}
	return &payment, nil
}

// GetByOrgAndYearMonth 查询组织某月的所有缴费记录
func (r *SIMonthlyPaymentRepository) GetByOrgAndYearMonth(ctx context.Context, orgID int64, yearMonth string) ([]SIMonthlyPayment, error) {
	var payments []SIMonthlyPayment
	err := r.db.WithContext(ctx).
		Scopes(middleware.TenantScope(orgID)).
		Where("year_month = ?", yearMonth).
		Find(&payments).Error
	return payments, err
}

// GetOverdueByOrg 查询组织所有欠缴记录（含员工姓名）
func (r *SIMonthlyPaymentRepository) GetOverdueByOrg(ctx context.Context, orgID int64) ([]SIMonthlyPayment, error) {
	var payments []SIMonthlyPayment
	err := r.db.WithContext(ctx).
		Scopes(middleware.TenantScope(orgID)).
		Where("status = ?", PaymentStatusOverdue).
		Order("year_month ASC").
		Find(&payments).Error
	return payments, err
}

// UpdateStatus 更新缴费状态（仅限 status 字段，per D-SI-09 INSERT ONLY 策略）
func (r *SIMonthlyPaymentRepository) UpdateStatus(ctx context.Context, orgID, id int64, status PaymentStatus) error {
	result := r.db.WithContext(ctx).
		Model(&SIMonthlyPayment{}).
		Scopes(middleware.TenantScope(orgID)).
		Where("id = ?", id).
		Update("status", status)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrRecordNotFound
	}
	return nil
}

// BatchUpsert 批量幂等写入月度缴费记录（D-SI-02：ON CONFLICT DO NOTHING）
func (r *SIMonthlyPaymentRepository) BatchUpsert(ctx context.Context, tx *gorm.DB, payments []SIMonthlyPayment) error {
	if len(payments) == 0 {
		return nil
	}
	db := r.getDB(tx)
	// PostgreSQL UPSERT：唯一约束 (org_id, employee_id, year_month) + soft delete
	return db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "org_id"}, {Name: "employee_id"}, {Name: "year_month"}},
		DoNothing: true,
	}).Create(&payments).Error
}

// SumFieldByOrgAndYearMonth 聚合查询某字段总和（供 Dashboard 使用）
func (r *SIMonthlyPaymentRepository) SumFieldByOrgAndYearMonth(ctx context.Context, orgID int64, yearMonth string, field string, statuses []PaymentStatus) (decimal.Decimal, error) {
	allowedFields := map[string]bool{"total_amount": true, "company_amount": true, "personal_amount": true}
	if !allowedFields[field] {
		return decimal.Zero, fmt.Errorf("invalid aggregation field: %s", field)
	}
	var result decimal.Decimal
	err := r.db.WithContext(ctx).
		Model(&SIMonthlyPayment{}).
		Scopes(middleware.TenantScope(orgID)).
		Where("year_month = ? AND status IN ?", yearMonth, statuses).
		Select(fmt.Sprintf("COALESCE(SUM(%s), 0)", field)).
		Scan(&result).Error
	return result, err
}

// UpdateOverduePayments 批量将 pending 记录更新为 overdue（D-SI-03：>=26日未缴）
func (r *SIMonthlyPaymentRepository) UpdateOverduePayments(ctx context.Context, orgID int64, yearMonth string) error {
	return r.db.WithContext(ctx).
		Model(&SIMonthlyPayment{}).
		Scopes(middleware.TenantScope(orgID)).
		Where("year_month = ? AND status = ?", yearMonth, PaymentStatusPending).
		Update("status", PaymentStatusOverdue).Error
}

// DeleteOlderThan 删除超过指定月数的过期记录（D-SI-02：>24个月）
func (r *SIMonthlyPaymentRepository) DeleteOlderThan(ctx context.Context, orgID int64, cutoffYearMonth string) error {
	cutoffTime, err := time.Parse("200601", cutoffYearMonth)
	if err != nil {
		return fmt.Errorf("invalid cutoff year_month: %w", err)
	}
	return r.db.WithContext(ctx).
		Scopes(middleware.TenantScope(orgID)).
		Where("year_month < ?", cutoffTime.Format("200601")).
		Delete(&SIMonthlyPayment{}).Error
}

// getDB 返回可用数据库连接（优先使用事务）
func (r *SIMonthlyPaymentRepository) getDB(tx *gorm.DB) *gorm.DB {
	if tx != nil {
		return tx
	}
	return r.db
}
