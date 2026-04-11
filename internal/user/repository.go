package user

import (
	"errors"
	"fmt"

	"github.com/wencai/easyhr/internal/common/crypto"
	"github.com/wencai/easyhr/internal/common/middleware"
	"github.com/wencai/easyhr/internal/common/model"
	"gorm.io/gorm"
)

var (
	ErrOwnerCannotDelete = errors.New("owner角色不可删除")
	ErrOwnerCannotDemote = errors.New("owner角色不可降级")
	ErrUserNotFound      = errors.New("用户不存在")
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) FindByPhoneHash(phoneHash string) (*model.User, error) {
	var user model.User
	err := r.db.Where("phone_hash = ?", phoneHash).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *Repository) FindByOrgID(orgID int64) (*model.Organization, error) {
	var org model.Organization
	err := r.db.Where("id = ?", orgID).First(&org).Error
	if err != nil {
		return nil, err
	}
	return &org, nil
}

func (r *Repository) CreateUser(user *model.User) error {
	return r.db.Create(user).Error
}

func (r *Repository) CreateOrg(org *model.Organization) error {
	return r.db.Create(org).Error
}

func (r *Repository) UpdateOrg(orgID int64, updates map[string]interface{}) error {
	return r.db.Model(&model.Organization{}).Where("id = ?", orgID).Updates(updates).Error
}

func (r *Repository) ListUsers(orgID int64, page, pageSize int) ([]model.User, int64, error) {
	var users []model.User
	var total int64
	q := r.db.Model(&model.User{}).Scopes(middleware.TenantScope(orgID))
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * pageSize
	if err := q.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&users).Error; err != nil {
		return nil, 0, err
	}
	return users, total, nil
}

func (r *Repository) FindUserByID(orgID, userID int64) (*model.User, error) {
	var user model.User
	err := r.db.Scopes(middleware.TenantScope(orgID)).Where("id = ?", userID).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *Repository) UpdateUserRole(orgID, userID int64, role string) error {
	var user model.User
	if err := r.db.Scopes(middleware.TenantScope(orgID)).Where("id = ?", userID).First(&user).Error; err != nil {
		return err
	}
	if user.Role == "owner" {
		return ErrOwnerCannotDemote
	}
	return r.db.Model(&user).Update("role", role).Error
}

func (r *Repository) DeleteUser(orgID, userID int64) error {
	var user model.User
	if err := r.db.Scopes(middleware.TenantScope(orgID)).Where("id = ?", userID).First(&user).Error; err != nil {
		return err
	}
	if user.Role == "owner" {
		return ErrOwnerCannotDelete
	}
	return r.db.Delete(&user).Error
}

func (r *Repository) FindByPhone(phone string, aesKey []byte) (*model.User, error) {
	hash := crypto.HashSHA256(phone)
	return r.FindByPhoneHash(hash)
}

// UpdateUserPassword 更新用户密码哈希
func (r *Repository) UpdateUserPassword(userID int64, passwordHash string) error {
	return r.db.Model(&model.User{}).Where("id = ?", userID).Update("password_hash", passwordHash).Error
}

func (r *Repository) CreateOrgAndOwner(org *model.Organization, user *model.User) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(org).Error; err != nil {
			return fmt.Errorf("create org: %w", err)
		}
		user.OrgID = org.ID
		if err := tx.Create(user).Error; err != nil {
			return fmt.Errorf("create user: %w", err)
		}
		return nil
	})
}
