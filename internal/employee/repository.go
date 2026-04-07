package employee

import (
	"errors"
	"fmt"

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
