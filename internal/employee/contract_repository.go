package employee

import (
	"fmt"

	"github.com/wencai/easyhr/internal/common/middleware"
	"gorm.io/gorm"
)

// ContractRepository 合同数据访问层
type ContractRepository struct {
	db *gorm.DB
}

// NewContractRepository 创建合同 Repository
func NewContractRepository(db *gorm.DB) *ContractRepository {
	return &ContractRepository{db: db}
}

// Create 创建合同记录
func (r *ContractRepository) Create(c *Contract) error {
	return r.db.Create(c).Error
}

// FindByID 根据 ID 查找合同（带租户隔离）
func (r *ContractRepository) FindByID(orgID, id int64) (*Contract, error) {
	var c Contract
	err := r.db.Scopes(middleware.TenantScope(orgID)).Where("id = ?", id).First(&c).Error
	if err != nil {
		return nil, err
	}
	return &c, nil
}

// Update 更新合同信息（部分更新）
func (r *ContractRepository) Update(orgID, id int64, updates map[string]interface{}) error {
	result := r.db.Model(&Contract{}).Scopes(middleware.TenantScope(orgID)).Where("id = ?", id).Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// Delete 删除合同（硬删除，仅限草稿/待签状态）
func (r *ContractRepository) Delete(orgID, id int64) error {
	result := r.db.Scopes(middleware.TenantScope(orgID)).Where("id = ?", id).Delete(&Contract{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// ListByEmployee 按员工查询合同列表（分页）
func (r *ContractRepository) ListByEmployee(orgID, employeeID int64, page, pageSize int) ([]Contract, int64, error) {
	var contracts []Contract
	var total int64

	q := r.db.Model(&Contract{}).Scopes(middleware.TenantScope(orgID)).Where("employee_id = ?", employeeID)

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count contracts: %w", err)
	}

	offset := (page - 1) * pageSize
	if err := q.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&contracts).Error; err != nil {
		return nil, 0, fmt.Errorf("list contracts by employee: %w", err)
	}

	return contracts, total, nil
}

// List 查询企业所有合同（支持状态筛选 + 分页）
func (r *ContractRepository) List(orgID int64, status string, page, pageSize int) ([]Contract, int64, error) {
	var contracts []Contract
	var total int64

	q := r.db.Model(&Contract{}).Scopes(middleware.TenantScope(orgID))

	if status != "" {
		q = q.Where("status = ?", status)
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count contracts: %w", err)
	}

	offset := (page - 1) * pageSize
	if err := q.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&contracts).Error; err != nil {
		return nil, 0, fmt.Errorf("list contracts: %w", err)
	}

	return contracts, total, nil
}

// UpsertSignCode 创建或更新签署验证码
func (r *ContractRepository) UpsertSignCode(signCode *ContractSignCode) error {
	return r.db.Where("contract_id = ? AND phone = ?", signCode.ContractID, signCode.Phone).
		Assign(ContractSignCode{
			Code:      signCode.Code,
			ExpiresAt: signCode.ExpiresAt,
			Verified:  false,
			SignToken: "",
		}).
		FirstOrCreate(signCode).Error
}

// FindLatestSignCode 查询最新验证码记录
func (r *ContractRepository) FindLatestSignCode(contractID int64, phone string) (*ContractSignCode, error) {
	var signCode ContractSignCode
	err := r.db.Where("contract_id = ? AND phone = ?", contractID, phone).
		Order("created_at DESC").
		First(&signCode).Error
	if err != nil {
		return nil, err
	}
	return &signCode, nil
}

// UpdateSignCode 更新签署验证码（设置 verified 和 sign_token）
func (r *ContractRepository) UpdateSignCode(signCode *ContractSignCode) error {
	return r.db.Model(&ContractSignCode{}).Where("id = ?", signCode.ID).Updates(map[string]interface{}{
		"verified":   signCode.Verified,
		"sign_token": signCode.SignToken,
		"expires_at": signCode.ExpiresAt,
	}).Error
}

// FindBySignToken 根据 SignToken 查找验证码记录
func (r *ContractRepository) FindBySignToken(signToken string) (*ContractSignCode, error) {
	var signCode ContractSignCode
	err := r.db.Where("sign_token = ?", signToken).First(&signCode).Error
	if err != nil {
		return nil, err
	}
	return &signCode, nil
}
