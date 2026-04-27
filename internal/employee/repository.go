package employee

import (
	"errors"
	"fmt"
	"time"

	"github.com/wencai/easyhr/internal/common/crypto"
	"github.com/wencai/easyhr/internal/common/middleware"
	"github.com/wencai/easyhr/internal/common/model"
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

// SearchByName 根据姓名模糊搜索员工（用于下拉列表）
func (r *Repository) SearchByName(orgID int64, name string) ([]Employee, error) {
	var employees []Employee
	err := r.db.Scopes(middleware.TenantScope(orgID)).
		Where("name LIKE ? AND status != ?", "%"+name+"%", "offboarded").
		Order("name ASC").
		Limit(20).
		Find(&employees).Error
	if err != nil {
		return nil, fmt.Errorf("search employees by name: %w", err)
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

// FindByPhoneHashGlobal 根据手机号哈希查找员工（跨租户，用于签署等无认证场景）
func (r *Repository) FindByPhoneHashGlobal(phoneHash string) (*Employee, error) {
	var emp Employee
	err := r.db.Where("phone_hash = ?", phoneHash).First(&emp).Error
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
		Select("id", "name", "position", "department_id", "position_id").
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

// UpdateDepartmentID 更新员工所属部门（用于部门转移）
func (r *Repository) UpdateDepartmentID(orgID, empID, deptID int64) error {
	result := r.db.Model(&Employee{}).Scopes(middleware.TenantScope(orgID)).
		Where("id = ?", empID).Updates(map[string]interface{}{"department_id": deptID})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
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

// GetSalaryAmounts 批量获取员工薪资
// 优先从 employees.salary 字段获取（创建员工时填写），如果没有则从 salary_items 表取最近生效月份的基本薪资
func (r *Repository) GetSalaryAmounts(orgID int64, employeeIDs []int64) (map[int64]float64, error) {
	result := make(map[int64]float64)
	if len(employeeIDs) == 0 {
		return result, nil
	}

	// 第一步：从 employees 表获取直接存储的薪资
	type empRow struct {
		ID     int64
		Salary *float64
	}
	var empRows []empRow
	err := r.db.Model(&Employee{}).
		Select("id, salary").
		Where("org_id = ? AND id IN ?", orgID, employeeIDs).
		Find(&empRows).Error
	if err != nil {
		return result, fmt.Errorf("get salary from employees: %w", err)
	}
	for _, row := range empRows {
		if row.Salary != nil && *row.Salary > 0 {
			result[row.ID] = *row.Salary
		}
	}

	// 第二步：如果 salary_items 表存在，从那里补充没有薪资的员工数据
	if !r.db.Migrator().HasTable("salary_items") {
		return result, nil
	}

	// 找出还没有薪资的员工 IDs
	missingIDs := make([]int64, 0)
	for _, id := range employeeIDs {
		if _, ok := result[id]; !ok {
			missingIDs = append(missingIDs, id)
		}
	}
	if len(missingIDs) == 0 {
		return result, nil
	}

	type itemRow struct {
		EmployeeID int64
		Amount     float64
	}
	var rows []itemRow

	// 子查询：每个员工最近 effective_month 的薪资项
	err = r.db.Raw(`
		SELECT si.employee_id, si.amount
		FROM salary_items si
		INNER JOIN (
			SELECT employee_id, MAX(effective_month) as max_month
			FROM salary_items
			WHERE org_id = ? AND employee_id IN ?
			GROUP BY employee_id
		) latest ON si.employee_id = latest.employee_id AND si.effective_month = latest.max_month
		WHERE si.org_id = ? AND si.employee_id IN ?
	`, orgID, missingIDs, orgID, missingIDs).Scan(&rows).Error

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

		// 取每个员工最近 active/signed/pending_sign 合同
	err := r.db.Raw(`
		SELECT c.employee_id, c.end_date
		FROM contracts c
		INNER JOIN (
			SELECT employee_id, MAX(created_at) as max_created
			FROM contracts
			WHERE org_id = ? AND employee_id IN ? AND status IN ('active', 'signed', 'pending_sign')
			GROUP BY employee_id
		) latest ON c.employee_id = latest.employee_id AND c.created_at = latest.max_created
		WHERE c.org_id = ? AND c.employee_id IN ? AND c.status IN ('active', 'signed', 'pending_sign')
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

// positionCountRow 岗位员工数量行
type positionCountRow struct {
	PositionID int64
	Count      int64
}

// deptCountRow 部门员工数量行
type deptCountRow struct {
	DepartmentID int64
	Count       int64
}

// CountByDepartmentIDGrouped 批量获取每个部门的员工数量
func (r *Repository) CountByDepartmentIDGrouped(orgID int64) ([]deptCountRow, error) {
	var results []deptCountRow
	err := r.db.Model(&Employee{}).
		Scopes(middleware.TenantScope(orgID)).
		Select("department_id, COUNT(*) as count").
		Where("department_id IS NOT NULL").
		Group("department_id").
		Scan(&results).Error
	if err != nil {
		return nil, fmt.Errorf("count by department_id: %w", err)
	}
	return results, nil
}

// CountByPositionIDGrouped 批量获取每个岗位的员工数量
func (r *Repository) CountByPositionIDGrouped(orgID int64) ([]positionCountRow, error) {
	var results []positionCountRow
	err := r.db.Model(&Employee{}).
		Scopes(middleware.TenantScope(orgID)).
		Select("position_id, COUNT(*) as count").
		Where("position_id IS NOT NULL").
		Group("position_id").
		Scan(&results).Error
	if err != nil {
		return nil, fmt.Errorf("count by position_id: %w", err)
	}
	return results, nil
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

// CreateEmployeeAndUser 在事务中创建员工和对应的用户账号（member 角色），并关联 user_id
func (r *Repository) CreateEmployeeAndUser(emp *Employee, user *model.User) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 1. 创建员工（含唯一性校验）
		if err := createEmployeeTx(tx, emp); err != nil {
			return err
		}

		// 2. 创建用户（member 角色，无密码，后续由员工设置或管理员设置）
		user.OrgID = emp.OrgID
		if err := tx.Create(user).Error; err != nil {
			return fmt.Errorf("创建关联账号失败: %w", err)
		}

		// 3. 回写 employees.user_id
		if user.ID <= 0 {
			return fmt.Errorf("创建关联账号失败: userID无效")
		}
		if err := tx.Model(&Employee{}).Where("id = ?", emp.ID).Update("user_id", user.ID).Error; err != nil {
			return fmt.Errorf("关联账号失败: %w", err)
		}

		return nil
	})
}

// createEmployeeTx 事务内创建员工（含唯一性校验）
func createEmployeeTx(tx *gorm.DB, emp *Employee) error {
	if emp.PhoneHash != "" {
		var count int64
		tx.Model(&Employee{}).Where("org_id = ? AND phone_hash = ?", emp.OrgID, emp.PhoneHash).Count(&count)
		if count > 0 {
			return ErrPhoneDuplicate
		}
	}
	if emp.IDCardHash != "" {
		var count int64
		tx.Model(&Employee{}).Where("org_id = ? AND id_card_hash = ?", emp.OrgID, emp.IDCardHash).Count(&count)
		if count > 0 {
			return ErrIDCardDuplicate
		}
	}
	return tx.Create(emp).Error
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
