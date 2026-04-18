package employee

import (
	"errors"
	"fmt"
	"time"

	"github.com/wencai/easyhr/internal/common/crypto"
	"github.com/wencai/easyhr/internal/common/middleware"
	"gorm.io/gorm"
)

var (
	ErrPhoneDuplicate  = errors.New("该手机号已存在")
	ErrIDCardDuplicate = errors.New("该身份证号已存在")
	ErrEmployeeNotFound = errors.New("员工不存在")
)

// SearchParams 员工搜索参数
type SearchParams struct {
	Name     string // 姓名模糊搜索
	Position string // 岗位模糊搜索
	Phone    string // 手机号明文（内部转hash精确匹配）
	Status   string // 状态精确筛选
}

// Repository 员工数据访问层
type Repository struct {
	db *gorm.DB
}

// NewRepository 创建员工 Repository
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// Create 创建员工记录（事务内校验唯一性）
func (r *Repository) Create(emp *Employee) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 校验手机号唯一性（同 org_id 内）
		if emp.PhoneHash != "" {
			var count int64
			tx.Model(&Employee{}).Scopes(middleware.TenantScope(emp.OrgID)).
				Where("phone_hash = ?", emp.PhoneHash).Count(&count)
			if count > 0 {
				return ErrPhoneDuplicate
			}
		}
		// 校验身份证号唯一性（同 org_id 内）
		if emp.IDCardHash != "" {
			var count int64
			tx.Model(&Employee{}).Scopes(middleware.TenantScope(emp.OrgID)).
				Where("id_card_hash = ?", emp.IDCardHash).Count(&count)
			if count > 0 {
				return ErrIDCardDuplicate
			}
		}
		return tx.Create(emp).Error
	})
}

// FindByID 根据 ID 查找员工（带租户隔离）
func (r *Repository) FindByID(orgID, id int64) (*Employee, error) {
	var emp Employee
	err := r.db.Scopes(middleware.TenantScope(orgID)).Where("id = ?", id).First(&emp).Error
	if err != nil {
		return nil, err
	}
	return &emp, nil
}

// Update 更新员工信息（部分更新）
func (r *Repository) Update(orgID, id int64, updates map[string]interface{}) error {
	result := r.db.Model(&Employee{}).Scopes(middleware.TenantScope(orgID)).Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// Delete 软删除员工
func (r *Repository) Delete(orgID, id int64) error {
	result := r.db.Scopes(middleware.TenantScope(orgID)).Where("id = ?", id).Delete(&Employee{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// List 搜索+分页查询员工列表
func (r *Repository) List(orgID int64, params SearchParams, page, pageSize int) ([]Employee, int64, error) {
	var employees []Employee
	var total int64

	q := r.db.Model(&Employee{}).Scopes(middleware.TenantScope(orgID))

	// 应用搜索条件
	q = r.applySearchFilters(q, params)

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count employees: %w", err)
	}

	offset := (page - 1) * pageSize
	if err := q.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&employees).Error; err != nil {
		return nil, 0, fmt.Errorf("list employees: %w", err)
	}

	return employees, total, nil
}

// FindByPhoneHash 根据手机号哈希查找员工（租户隔离）
func (r *Repository) FindByPhoneHash(orgID int64, phoneHash string) (*Employee, error) {
	var emp Employee
	err := r.db.Scopes(middleware.TenantScope(orgID)).Where("phone_hash = ?", phoneHash).First(&emp).Error
	if err != nil {
		return nil, err
	}
	return &emp, nil
}

// FindByIDCardHash 根据身份证哈希查找员工（租户隔离）
func (r *Repository) FindByIDCardHash(orgID int64, idCardHash string) (*Employee, error) {
	var emp Employee
	err := r.db.Scopes(middleware.TenantScope(orgID)).Where("id_card_hash = ?", idCardHash).First(&emp).Error
	if err != nil {
		return nil, err
	}
	return &emp, nil
}

// FindAllForExport 获取全部匹配数据（不分页，用于导出）
func (r *Repository) FindAllForExport(orgID int64, params SearchParams) ([]Employee, error) {
	var employees []Employee
	q := r.db.Scopes(middleware.TenantScope(orgID))
	q = r.applySearchFilters(q, params)

	if err := q.Order("created_at DESC").Find(&employees).Error; err != nil {
		return nil, fmt.Errorf("find all for export: %w", err)
	}
	return employees, nil
}

// FindByIDs 根据多个 ID 批量查找员工（带租户隔离）
func (r *Repository) FindByIDs(orgID int64, ids []int64) ([]Employee, error) {
	var employees []Employee
	err := r.db.Scopes(middleware.TenantScope(orgID)).Where("id IN ?", ids).Find(&employees).Error
	if err != nil {
		return nil, fmt.Errorf("find employees by IDs: %w", err)
	}
	return employees, nil
}

// FindByUserID 根据 user_id 查找员工（带租户隔离）
func (r *Repository) FindByUserID(orgID int64, userID int64) (*Employee, error) {
	var emp Employee
	err := r.db.Scopes(middleware.TenantScope(orgID)).Where("user_id = ?", userID).First(&emp).Error
	if err != nil {
		return nil, err
	}
	return &emp, nil
}

// ListAllByOrg 获取指定企业全部在职/试用期员工（用于组织架构树构建）
// 仅选择 ID/Name/Position/DepartmentID 字段，避免加载加密数据
func (r *Repository) ListAllByOrg(orgID int64) ([]Employee, error) {
	var employees []Employee
	err := r.db.Scopes(middleware.TenantScope(orgID)).
		Where("status IN ?", []string{StatusActive, StatusProbation}).
		Select("id", "name", "position", "department_id").
		Find(&employees).Error
	if err != nil {
		return nil, fmt.Errorf("list all employees by org: %w", err)
	}
	return employees, nil
}

// CountByDepartment 统计指定部门下的员工数量
func (r *Repository) CountByDepartment(orgID, departmentID int64) (int64, error) {
	var count int64
	err := r.db.Model(&Employee{}).Scopes(middleware.TenantScope(orgID)).
		Where("department_id = ?", departmentID).
		Count(&count).Error
	return count, err
}

// ListRoster 花名册分页查询（支持部门筛选+综合搜索）
func (r *Repository) ListRoster(orgID int64, params ListQueryParams, page, pageSize int) ([]Employee, int64, error) {
	var employees []Employee
	var total int64

	q := r.db.Model(&Employee{}).Scopes(middleware.TenantScope(orgID))

	// 综合搜索：姓名/岗位模糊搜索（手机号需hash，暂不支持search内手机号搜索）
	if params.Search != "" {
		q = q.Where("name LIKE ? OR position LIKE ?", "%"+params.Search+"%", "%"+params.Search+"%")
	}
	if params.Name != "" {
		q = q.Where("name LIKE ?", "%"+params.Name+"%")
	}
	if params.Position != "" {
		q = q.Where("position LIKE ?", "%"+params.Position+"%")
	}
	if params.Phone != "" {
		phoneHash := crypto.HashSHA256(params.Phone)
		q = q.Where("phone_hash = ?", phoneHash)
	}
	if params.Status != "" {
		q = q.Where("status = ?", params.Status)
	}
	if params.DepartmentID != nil {
		q = q.Where("department_id = ?", *params.DepartmentID)
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count roster: %w", err)
	}

	offset := (page - 1) * pageSize
	if err := q.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&employees).Error; err != nil {
		return nil, 0, fmt.Errorf("list roster: %w", err)
	}

	return employees, total, nil
}

// FindAllForRosterExport 获取全部匹配数据用于花名册导出（不分页）
func (r *Repository) FindAllForRosterExport(orgID int64, params ListQueryParams) ([]Employee, error) {
	var employees []Employee
	q := r.db.Scopes(middleware.TenantScope(orgID))

	if params.Search != "" {
		q = q.Where("name LIKE ? OR position LIKE ?", "%"+params.Search+"%", "%"+params.Search+"%")
	}
	if params.Name != "" {
		q = q.Where("name LIKE ?", "%"+params.Name+"%")
	}
	if params.Position != "" {
		q = q.Where("position LIKE ?", "%"+params.Position+"%")
	}
	if params.Phone != "" {
		phoneHash := crypto.HashSHA256(params.Phone)
		q = q.Where("phone_hash = ?", phoneHash)
	}
	if params.Status != "" {
		q = q.Where("status = ?", params.Status)
	}
	if params.DepartmentID != nil {
		q = q.Where("department_id = ?", *params.DepartmentID)
	}

	if err := q.Order("created_at DESC").Find(&employees).Error; err != nil {
		return nil, fmt.Errorf("find all for roster export: %w", err)
	}
	return employees, nil
}

// GetSalaryAmounts 批量获取员工薪资（取最近生效月份的基本薪资）
func (r *Repository) GetSalaryAmounts(orgID int64, employeeIDs []int64) (map[int64]float64, error) {
	result := make(map[int64]float64)
	if len(employeeIDs) == 0 {
		return result, nil
	}

	// 检查 salary_items 表是否存在
	if !r.db.Migrator().HasTable("salary_items") {
		return result, nil
	}

	type row struct {
		EmployeeID int64
		Amount     float64
	}
	var rows []row

	// 子查询：每个员工最近 effective_month 的 "基本工资" 类薪资项
	err := r.db.Raw(`
		SELECT si.employee_id, si.amount
		FROM salary_items si
		INNER JOIN (
			SELECT employee_id, MAX(effective_month) as max_month
			FROM salary_items
			WHERE org_id = ? AND employee_id IN ?
			GROUP BY employee_id
		) latest ON si.employee_id = latest.employee_id AND si.effective_month = latest.max_month
		WHERE si.org_id = ? AND si.employee_id IN ?
	`, orgID, employeeIDs, orgID, employeeIDs).Scan(&rows).Error

	if err != nil {
		return result, fmt.Errorf("get salary amounts: %w", err)
	}

	for _, row := range rows {
		result[row.EmployeeID] = row.Amount
	}

	return result, nil
}

// GetContractExpiryDays 批量获取员工合同到期天数
// 返回 map[employeeID]*int: 正数=剩余天数, 负数=已过期天数, nil=无固定期限/无合同
func (r *Repository) GetContractExpiryDays(orgID int64, employeeIDs []int64) (map[int64]*int, error) {
	result := make(map[int64]*int)
	if len(employeeIDs) == 0 {
		return result, nil
	}

	// 检查 contracts 表是否存在
	if !r.db.Migrator().HasTable("contracts") {
		return result, nil
	}

	type row struct {
		EmployeeID int64
		EndDate    *time.Time
	}
	var rows []row

	// 取每个员工最近 active/signed 合同
	err := r.db.Raw(`
		SELECT c.employee_id, c.end_date
		FROM contracts c
		INNER JOIN (
			SELECT employee_id, MAX(created_at) as max_created
			FROM contracts
			WHERE org_id = ? AND employee_id IN ? AND status IN ('active', 'signed')
			GROUP BY employee_id
		) latest ON c.employee_id = latest.employee_id AND c.created_at = latest.max_created
		WHERE c.org_id = ? AND c.employee_id IN ? AND c.status IN ('active', 'signed')
	`, orgID, employeeIDs, orgID, employeeIDs).Scan(&rows).Error

	if err != nil {
		return result, fmt.Errorf("get contract expiry days: %w", err)
	}

	now := time.Now()
	for _, row := range rows {
		if row.EndDate == nil {
			// 无固定期限合同
			result[row.EmployeeID] = nil
			continue
		}
		days := int(row.EndDate.Sub(now).Hours() / 24)
		result[row.EmployeeID] = &days
	}

	return result, nil
}

// GetDepartmentNames 批量获取部门名称
func (r *Repository) GetDepartmentNames(orgID int64, deptIDs []int64) (map[int64]string, error) {
	result := make(map[int64]string)
	if len(deptIDs) == 0 {
		return result, nil
	}

	// 检查 departments 表是否存在
	if !r.db.Migrator().HasTable("departments") {
		return result, nil
	}

	type row struct {
		ID   int64
		Name string
	}
	var rows []row

	err := r.db.Raw(`SELECT id, name FROM departments WHERE org_id = ? AND id IN ?`, orgID, deptIDs).Scan(&rows).Error
	if err != nil {
		return result, fmt.Errorf("get department names: %w", err)
	}

	for _, row := range rows {
		result[row.ID] = row.Name
	}

	return result, nil
}

// applySearchFilters 应用搜索过滤条件
func (r *Repository) applySearchFilters(q *gorm.DB, params SearchParams) *gorm.DB {
	if params.Name != "" {
		q = q.Where("name LIKE ?", "%"+params.Name+"%")
	}
	if params.Position != "" {
		q = q.Where("position LIKE ?", "%"+params.Position+"%")
	}
	if params.Phone != "" {
		phoneHash := crypto.HashSHA256(params.Phone)
		q = q.Where("phone_hash = ?", phoneHash)
	}
	if params.Status != "" {
		q = q.Where("status = ?", params.Status)
	}
	return q
}
