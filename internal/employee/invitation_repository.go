package employee

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

// InvitationRepository 邀请数据访问层
type InvitationRepository struct {
	db *gorm.DB
}

// NewInvitationRepository 创建邀请 Repository
func NewInvitationRepository(db *gorm.DB) *InvitationRepository {
	return &InvitationRepository{db: db}
}

// DB 暴露数据库连接（供 Service 层事务使用）
func (r *InvitationRepository) DB() *gorm.DB {
	return r.db
}

// Create 创建邀请记录
func (r *InvitationRepository) Create(inv *Invitation) error {
	return r.db.Create(inv).Error
}

// FindByToken 根据 token 查找邀请
func (r *InvitationRepository) FindByToken(token string) (*Invitation, error) {
	var inv Invitation
	err := r.db.Where("token = ?", token).First(&inv).Error
	if err != nil {
		return nil, err
	}
	return &inv, nil
}

// FindByID 根据 ID 查找邀请
func (r *InvitationRepository) FindByID(orgID, id int64) (*Invitation, error) {
	var inv Invitation
	err := r.db.Where("org_id = ? AND id = ?", orgID, id).First(&inv).Error
	if err != nil {
		return nil, err
	}
	return &inv, nil
}

// UpdateStatus 更新邀请状态
func (r *InvitationRepository) UpdateStatus(token string, status string, employeeID *int64) error {
	updates := map[string]interface{}{
		"status": status,
	}
	if employeeID != nil {
		updates["employee_id"] = *employeeID
	}
	if status == InvitationStatusUsed {
		now := time.Now()
		updates["used_at"] = &now
	}

	result := r.db.Model(&Invitation{}).Where("token = ?", token).Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// List 分页查询邀请列表
func (r *InvitationRepository) List(orgID int64, status string, page, pageSize int) ([]Invitation, int64, error) {
	var invitations []Invitation
	var total int64

	q := r.db.Model(&Invitation{}).Where("org_id = ?", orgID)

	if status != "" {
		q = q.Where("status = ?", status)
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count invitations: %w", err)
	}

	offset := (page - 1) * pageSize
	if err := q.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&invitations).Error; err != nil {
		return nil, 0, fmt.Errorf("list invitations: %w", err)
	}

	return invitations, total, nil
}

// FindOrgName 查询组织名称
func (r *InvitationRepository) FindOrgName(orgID int64) (string, error) {
	type orgResult struct {
		Name string
	}
	var result orgResult
	err := r.db.Table("organizations").Where("id = ?", orgID).Select("name").Scan(&result).Error
	if err != nil {
		return "", fmt.Errorf("查询企业名称失败: %w", err)
	}
	return result.Name, nil
}
