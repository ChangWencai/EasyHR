package employee

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/wencai/easyhr/internal/common/config"
	"github.com/wencai/easyhr/internal/common/crypto"
	"github.com/wencai/easyhr/internal/common/middleware"
	"gorm.io/gorm"
)

var (
	ErrInvitationNotFound  = errors.New("邀请不存在")
	ErrInvitationUsed      = errors.New("邀请已使用")
	ErrInvitationExpired   = errors.New("邀请已过期")
	ErrInvitationCancelled = errors.New("邀请已取消")
	ErrEmployeeNotPending  = errors.New("员工状态不是待入职")
)

// InvitationService 邀请业务逻辑层
type InvitationService struct {
	invRepo   *InvitationRepository
	empRepo   *Repository
	cryptoCfg config.CryptoConfig
}

// NewInvitationService 创建邀请 Service
func NewInvitationService(invRepo *InvitationRepository, empRepo *Repository, cryptoCfg config.CryptoConfig) *InvitationService {
	return &InvitationService{
		invRepo:   invRepo,
		empRepo:   empRepo,
		cryptoCfg: cryptoCfg,
	}
}

// aesKey 获取 AES 密钥字节
func (s *InvitationService) aesKey() []byte {
	return []byte(s.cryptoCfg.AESKey)
}

// generateToken 使用 crypto/rand 生成 32 字节随机数，返回 64 字符 hex 字符串
func generateToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("生成 token 失败: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}

// CreateInvitation 创建入职邀请
func (s *InvitationService) CreateInvitation(orgID, userID int64, req *CreateInvitationRequest) (*CreateInvitationResponse, error) {
	token, err := generateToken()
	if err != nil {
		return nil, err
	}

	now := time.Now()
	expiresAt := now.Add(InvitationExpiryDuration)

	inv := &Invitation{
		OrgID:     orgID,
		Token:     token,
		Position:  req.Position,
		Status:    InvitationStatusPending,
		CreatedBy: userID,
		ExpiresAt: expiresAt,
	}

	if err := s.invRepo.Create(inv); err != nil {
		return nil, fmt.Errorf("创建邀请失败: %w", err)
	}

	return &CreateInvitationResponse{
		Token:     token,
		InviteURL: "/invite/" + token,
		ExpiresAt: expiresAt.Format("2006-01-02T15:04:05Z07:00"),
	}, nil
}

// GetInvitationDetail 获取邀请详情（公开接口，不含敏感信息）
func (s *InvitationService) GetInvitationDetail(token string) (*InvitationDetailResponse, error) {
	inv, err := s.invRepo.FindByToken(token)
	if err != nil {
		return nil, ErrInvitationNotFound
	}

	// 检查邀请状态
	if inv.Status == InvitationStatusUsed {
		return nil, ErrInvitationUsed
	}
	if inv.Status == InvitationStatusCancelled {
		return nil, ErrInvitationCancelled
	}

	// 检查是否过期
	if time.Now().After(inv.ExpiresAt) {
		return nil, ErrInvitationExpired
	}

	// 查询组织名称
	orgName, err := s.invRepo.FindOrgName(inv.OrgID)
	if err != nil {
		return nil, fmt.Errorf("查询企业信息失败: %w", err)
	}

	return &InvitationDetailResponse{
		OrgName:   orgName,
		Position:  inv.Position,
		Status:    inv.Status,
		ExpiresAt: inv.ExpiresAt.Format("2006-01-02T15:04:05Z07:00"),
	}, nil
}

// SubmitInvitation 员工通过邀请链接提交个人信息
func (s *InvitationService) SubmitInvitation(token string, req *SubmitInvitationRequest) error {
	// 1. 查找邀请
	inv, err := s.invRepo.FindByToken(token)
	if err != nil {
		return ErrInvitationNotFound
	}

	// 2. 验证状态
	if inv.Status == InvitationStatusUsed {
		return ErrInvitationUsed
	}
	if inv.Status == InvitationStatusCancelled {
		return ErrInvitationCancelled
	}

	// 3. 验证过期
	if !time.Now().Before(inv.ExpiresAt) {
		return ErrInvitationExpired
	}

	aesKey := s.aesKey()

	// 4. 检查手机号唯一性
	phoneHash := crypto.HashSHA256(req.Phone)
	existingEmp, err := s.empRepo.FindByPhoneHash(inv.OrgID, phoneHash)
	if err == nil && existingEmp != nil {
		return fmt.Errorf("该手机号已存在")
	}

	// 5. 检查身份证号唯一性
	idCardHash := crypto.HashSHA256(req.IDCard)
	existingEmp, err = s.empRepo.FindByIDCardHash(inv.OrgID, idCardHash)
	if err == nil && existingEmp != nil {
		return fmt.Errorf("该身份证号已存在")
	}

	// 6. 从身份证提取性别和出生日期
	gender, birthDate, err := extractFromIDCard(req.IDCard)
	if err != nil {
		return fmt.Errorf("身份证号解析失败: %w", err)
	}

	// 7. 加密敏感字段
	phoneEncrypted, err := crypto.Encrypt(req.Phone, aesKey)
	if err != nil {
		return fmt.Errorf("加密手机号失败: %w", err)
	}
	idCardEncrypted, err := crypto.Encrypt(req.IDCard, aesKey)
	if err != nil {
		return fmt.Errorf("加密身份证号失败: %w", err)
	}

	// 8. 解析入职日期
	hireDate, err := time.Parse("2006-01-02", req.HireDate)
	if err != nil {
		return fmt.Errorf("入职日期格式错误: %w", err)
	}

	// 9. 在事务中创建 Employee + 更新 Invitation
	emp := &Employee{}
	emp.OrgID = inv.OrgID
	emp.Name = req.Name
	emp.PhoneEncrypted = phoneEncrypted
	emp.PhoneHash = phoneHash
	emp.IDCardEncrypted = idCardEncrypted
	emp.IDCardHash = idCardHash
	emp.Gender = gender
	emp.BirthDate = &birthDate
	emp.Position = req.Position
	emp.HireDate = hireDate
	emp.Status = StatusPending
	emp.CreatedBy = inv.CreatedBy
	emp.UpdatedBy = inv.CreatedBy

	return s.invRepo.DB().Transaction(func(tx *gorm.DB) error {
		// 在事务中校验唯一性并创建 Employee
		if emp.PhoneHash != "" {
			var count int64
			tx.Model(&Employee{}).Scopes(middleware.TenantScope(emp.OrgID)).
				Where("phone_hash = ?", emp.PhoneHash).Count(&count)
			if count > 0 {
				return fmt.Errorf("该手机号已存在")
			}
		}
		if emp.IDCardHash != "" {
			var count int64
			tx.Model(&Employee{}).Scopes(middleware.TenantScope(emp.OrgID)).
				Where("id_card_hash = ?", emp.IDCardHash).Count(&count)
			if count > 0 {
				return fmt.Errorf("该身份证号已存在")
			}
		}

		// 创建 Employee
		if err := tx.Create(emp).Error; err != nil {
			return fmt.Errorf("创建员工失败: %w", err)
		}

		// 更新邀请状态
		now := time.Now()
		if err := tx.Model(&Invitation{}).Where("token = ?", token).Updates(map[string]interface{}{
			"status":      InvitationStatusUsed,
			"employee_id": emp.ID,
			"used_at":     &now,
		}).Error; err != nil {
			return fmt.Errorf("更新邀请状态失败: %w", err)
		}

		return nil
	})
}

// ListInvitations 查询邀请列表
func (s *InvitationService) ListInvitations(orgID int64, query ListInvitationsQuery) ([]InvitationListItem, int64, error) {
	page := query.Page
	pageSize := query.PageSize
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	invitations, total, err := s.invRepo.List(orgID, query.Status, page, pageSize)
	if err != nil {
		return nil, 0, fmt.Errorf("查询邀请列表失败: %w", err)
	}

	items := make([]InvitationListItem, 0, len(invitations))
	for _, inv := range invitations {
		item := InvitationListItem{
			ID:         inv.ID,
			Token:      inv.Token,
			Position:   inv.Position,
			Status:     inv.Status,
			CreatedAt:  inv.CreatedAt,
			ExpiresAt:  inv.ExpiresAt,
			UsedAt:     inv.UsedAt,
			EmployeeID: inv.EmployeeID,
		}

		// 如果邀请已使用且有员工 ID，查询员工姓名
		if inv.EmployeeID != nil {
			emp, err := s.empRepo.FindByID(inv.OrgID, *inv.EmployeeID)
			if err == nil {
				item.EmployeeName = emp.Name
			}
		}

		items = append(items, item)
	}

	return items, total, nil
}

// CancelInvitation 取消邀请（仅 pending 状态可取消）
func (s *InvitationService) CancelInvitation(orgID, id int64) error {
	inv, err := s.invRepo.FindByID(orgID, id)
	if err != nil {
		return ErrInvitationNotFound
	}

	if inv.Status != InvitationStatusPending {
		return fmt.Errorf("仅待使用的邀请可以取消")
	}

	if err := s.invRepo.UpdateStatus(inv.Token, InvitationStatusCancelled, nil); err != nil {
		return fmt.Errorf("取消邀请失败: %w", err)
	}

	return nil
}

// ConfirmOnboarding 确认入职（将员工状态从 pending 更新为 active）
func (s *InvitationService) ConfirmOnboarding(orgID, employeeID int64) error {
	emp, err := s.empRepo.FindByID(orgID, employeeID)
	if err != nil {
		return fmt.Errorf("员工不存在")
	}

	if emp.Status != StatusPending {
		return ErrEmployeeNotPending
	}

	if err := s.empRepo.Update(orgID, employeeID, map[string]interface{}{
		"status": StatusActive,
	}); err != nil {
		return fmt.Errorf("确认入职失败: %w", err)
	}

	return nil
}
