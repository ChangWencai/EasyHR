package socialinsurance

import (
	"fmt"

	"gorm.io/gorm"
)

// ErrPolicyNotFound 政策未找到
var ErrPolicyNotFound = fmt.Errorf("社保政策不存在")

// Repository 社保政策数据访问层
type Repository struct {
	db *gorm.DB
}

// NewRepository 创建社保政策 Repository
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

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
