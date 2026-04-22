package email_template

import (
	"github.com/wencai/easyhr/internal/common/middleware"
	"gorm.io/gorm"
)

// Repository 邮箱模板数据访问层
type Repository struct {
	db *gorm.DB
}

// NewRepository 创建 Repository
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// Create 创建模板
func (r *Repository) Create(tpl *EmailTemplate) error {
	return r.db.Create(tpl).Error
}

// FindByID 根据 ID 查询
func (r *Repository) FindByID(orgID, id int64) (*EmailTemplate, error) {
	var tpl EmailTemplate
	err := r.db.Scopes(middleware.TenantScope(orgID)).Where("id = ?", id).First(&tpl).Error
	return &tpl, err
}

// FindByName 根据名称查询（用于唯一性校验）
func (r *Repository) FindByName(orgID int64, name string) (*EmailTemplate, error) {
	var tpl EmailTemplate
	err := r.db.Scopes(middleware.TenantScope(orgID)).Where("name = ?", name).First(&tpl).Error
	return &tpl, err
}

// List 查询列表
func (r *Repository) List(orgID int64, page, pageSize int) ([]EmailTemplate, int64, error) {
	var templates []EmailTemplate
	var total int64

	q := r.db.Model(&EmailTemplate{}).Scopes(middleware.TenantScope(orgID))
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := q.Offset(offset).Limit(pageSize).Order("is_default DESC, id DESC").Find(&templates).Error; err != nil {
		return nil, 0, err
	}

	return templates, total, nil
}

// Update 更新模板
func (r *Repository) Update(orgID, id int64, updates map[string]interface{}) error {
	result := r.db.Model(&EmailTemplate{}).Scopes(middleware.TenantScope(orgID)).Where("id = ?", id).Updates(updates)
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return result.Error
}

// ClearDefault 清除该企业的默认标记
func (r *Repository) ClearDefault(orgID int64) error {
	return r.db.Model(&EmailTemplate{}).Scopes(middleware.TenantScope(orgID)).
		Where("is_default = ?", true).
		Update("is_default", false).Error
}

// Delete 删除模板（软删除）
func (r *Repository) Delete(orgID, id int64) error {
	return r.db.Scopes(middleware.TenantScope(orgID)).Where("id = ?", id).Delete(&EmailTemplate{}).Error
}
