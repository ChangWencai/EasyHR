package user

import (
	"errors"
	"fmt"

	"github.com/wencai/easyhr/internal/common/crypto"
	"github.com/wencai/easyhr/internal/common/logger"
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

// FindByID 根据 userID 查询用户（不使用 tenant scope，用于 token 刷新等场景）
func (r *Repository) FindByID(userID int64) (*model.User, error) {
	var user model.User
	err := r.db.Where("id = ?", userID).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
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

func (r *Repository) UpdateUserOrgID(userID, orgID int64) error {
	result := r.db.Model(&model.User{}).Where("id = ?", userID).Update("org_id", orgID)
	if result.Error != nil {
		logger.SugarLogger.Errorw("UpdateUserOrgID: 更新失败", "userID", userID, "orgID", orgID, "error", result.Error.Error())
		return result.Error
	}
	if result.RowsAffected == 0 {
		logger.SugarLogger.Warnw("UpdateUserOrgID: 未找到用户或org_id已相同", "userID", userID, "orgID", orgID, "rowsAffected", result.RowsAffected)
	}
	logger.SugarLogger.Infow("UpdateUserOrgID: 更新完成", "userID", userID, "orgID", orgID, "rowsAffected", result.RowsAffected)
	return nil
}

func (r *Repository) UpdateUserAvatar(userID int64, avatar string) error {
	return r.db.Model(&model.User{}).Where("id = ?", userID).Update("avatar", avatar).Error
}

func (r *Repository) UpdateUserName(userID int64, name string) error {
	return r.db.Model(&model.User{}).Where("id = ?", userID).Update("name", name).Error
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

// FindEmployeeByPhoneHashAndOrg 根据手机号哈希和企业ID查找员工记录
func (r *Repository) FindEmployeeByPhoneHashAndOrg(orgID int64, phoneHash string) (employeeID int64, found bool, err error) {
	var emp struct{ ID int64 }
	err = r.db.Table("employees").
		Select("id").
		Scopes(middleware.TenantScope(orgID)).
		Where("phone_hash = ? AND deleted_at IS NULL", phoneHash).
		Scan(&emp).Error
	if err != nil {
		return 0, false, err
	}
	if emp.ID == 0 {
		return 0, false, nil
	}
	return int64(emp.ID), true, nil
}

// LinkEmployeeToUser 将员工记录关联到用户账号（在创建子账户时调用）
func (r *Repository) LinkEmployeeToUser(orgID, employeeID, userID int64) error {
	result := r.db.Table("employees").
		Scopes(middleware.TenantScope(orgID)).
		Where("id = ? AND deleted_at IS NULL", employeeID).
		Update("user_id", userID)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("员工不存在或无权限")
	}
	return nil
}

// CreateUserAndLinkEmployee 在事务中创建子账户并关联同名员工
func (r *Repository) CreateUserAndLinkEmployee(orgID int64, user *model.User) (createdUser *model.User, err error) {
	err = r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(user).Error; err != nil {
			return fmt.Errorf("create user: %w", err)
		}

		// 查找同名下是否有员工档案，若有则关联 user_id
		var emp struct{ ID int64 }
		if err := tx.Table("employees").
			Select("id").
			Scopes(middleware.TenantScope(orgID)).
			Where("phone_hash = ? AND deleted_at IS NULL", user.PhoneHash).
			Scan(&emp).Error; err != nil {
			logger.SugarLogger.Warnw("FindEmployeeByPhoneHash failed in CreateUserAndLinkEmployee", "error", err.Error())
		} else if emp.ID > 0 {
			result := tx.Table("employees").
				Scopes(middleware.TenantScope(orgID)).
				Where("id = ? AND deleted_at IS NULL", emp.ID).
				Update("user_id", user.ID)
			if result.Error != nil {
				logger.SugarLogger.Warnw("LinkEmployeeToUser failed", "employeeID", emp.ID, "userID", user.ID, "error", result.Error.Error())
			} else if result.RowsAffected > 0 {
				logger.SugarLogger.Infow("Employee linked to user", "employeeID", emp.ID, "userID", user.ID, "phoneHash", user.PhoneHash)
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}
	return user, nil
}
