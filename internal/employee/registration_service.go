package employee

import (
	"errors"
	"fmt"
	"time"

	"github.com/wencai/easyhr/internal/common/config"
	"github.com/wencai/easyhr/internal/common/crypto"
	"github.com/wencai/easyhr/internal/common/middleware"
	"gorm.io/gorm"
)

var (
	ErrRegistrationNotFound   = errors.New("登记记录不存在")
	ErrRegistrationExpired    = errors.New("登记已过期")
	ErrRegistrationAlreadyUsed = errors.New("登记已提交")
)

// RegistrationService 员工信息登记业务逻辑层
type RegistrationService struct {
	regRepo    *RegistrationRepository
	empRepo    *Repository
	cryptoCfg  config.CryptoConfig
}

// NewRegistrationService 创建登记 Service
func NewRegistrationService(regRepo *RegistrationRepository, empRepo *Repository, cryptoCfg config.CryptoConfig) *RegistrationService {
	return &RegistrationService{
		regRepo:   regRepo,
		empRepo:   empRepo,
		cryptoCfg: cryptoCfg,
	}
}

// aesKey 获取 AES 密钥字节
func (s *RegistrationService) aesKey() []byte {
	return []byte(s.cryptoCfg.AESKey)
}

// CreateRegistration 创建员工信息登记表
func (s *RegistrationService) CreateRegistration(orgID, userID int64, req *CreateRegistrationRequest) (*RegistrationResponse, error) {
	token, err := generateToken()
	if err != nil {
		return nil, err
	}

	now := time.Now()
	expiresAt := now.Add(RegistrationExpiryDuration)

	var hireDate *time.Time
	if req.HireDate != "" {
		parsed, err := time.Parse("2006-01-02", req.HireDate)
		if err != nil {
			return nil, fmt.Errorf("入职日期格式错误: %w", err)
		}
		hireDate = &parsed
	}

	reg := &Registration{
		OrgID:        orgID,
		EmployeeID:   req.EmployeeID,
		Name:         req.Name,
		DepartmentID: req.DepartmentID,
		Position:     req.Position,
		HireDate:     hireDate,
		Token:        token,
		Status:       RegistrationStatusPending,
		ExpiresAt:    expiresAt,
		CreatedBy:    userID,
	}

	if err := s.regRepo.Create(reg); err != nil {
		return nil, fmt.Errorf("创建登记表失败: %w", err)
	}

	return &RegistrationResponse{
		ID:        reg.ID,
		Token:     token,
		Status:    RegistrationStatusPending,
		ExpiresAt: expiresAt,
		CreatedAt: now,
	}, nil
}

// GetRegistrationDetail 获取登记详情（公开接口，仅返回基础信息）
func (s *RegistrationService) GetRegistrationDetail(token string) (*RegistrationDetailResponse, error) {
	reg, err := s.regRepo.FindByToken(token)
	if err != nil {
		return nil, ErrRegistrationNotFound
	}

	// 检查是否过期
	if reg.Status == RegistrationStatusPending && time.Now().After(reg.ExpiresAt) {
		return nil, ErrRegistrationExpired
	}

	if reg.Status == RegistrationStatusUsed {
		return nil, ErrRegistrationAlreadyUsed
	}

	resp := &RegistrationDetailResponse{
		Name:     reg.Name,
		Position: reg.Position,
		Status:   reg.Status,
	}

	if reg.DepartmentID != nil {
		resp.DepartmentID = reg.DepartmentID
	}

	if reg.HireDate != nil {
		resp.HireDate = reg.HireDate.Format("2006-01-02")
	}

	return resp, nil
}

// SubmitRegistration 员工通过登记链接提交个人信息
func (s *RegistrationService) SubmitRegistration(token string, req *SubmitRegistrationRequest) error {
	// 1. 查找登记记录
	reg, err := s.regRepo.FindByToken(token)
	if err != nil {
		return ErrRegistrationNotFound
	}

	// 2. 验证状态
	if reg.Status == RegistrationStatusUsed {
		return ErrRegistrationAlreadyUsed
	}

	// 3. 验证过期
	if !time.Now().Before(reg.ExpiresAt) {
		return ErrRegistrationExpired
	}

	aesKey := s.aesKey()

	// 4. 加密敏感字段
	var phoneEncrypted, phoneHash string
	if req.Phone != "" {
		phoneHash = crypto.HashSHA256(req.Phone)
		phoneEncrypted, err = crypto.Encrypt(req.Phone, aesKey)
		if err != nil {
			return fmt.Errorf("加密手机号失败: %w", err)
		}
	}

	var idCardEncrypted, idCardHash string
	if req.IDCard != "" {
		idCardHash = crypto.HashSHA256(req.IDCard)
		idCardEncrypted, err = crypto.Encrypt(req.IDCard, aesKey)
		if err != nil {
			return fmt.Errorf("加密身份证号失败: %w", err)
		}
	}

	var bankAccountEncrypted, bankAccountHash string
	if req.BankAccount != "" {
		bankAccountHash = crypto.HashSHA256(req.BankAccount)
		bankAccountEncrypted, err = crypto.Encrypt(req.BankAccount, aesKey)
		if err != nil {
			return fmt.Errorf("加密银行卡号失败: %w", err)
		}
	}

	var emergencyPhoneEncrypted, emergencyPhoneHash string
	if req.EmergencyPhone != "" {
		emergencyPhoneHash = crypto.HashSHA256(req.EmergencyPhone)
		emergencyPhoneEncrypted, err = crypto.Encrypt(req.EmergencyPhone, aesKey)
		if err != nil {
			return fmt.Errorf("加密紧急联系人电话失败: %w", err)
		}
	}

	// 5. 在事务中执行：查找/创建/更新 Employee + 更新 Registration 状态
	return s.regRepo.DB().Transaction(func(tx *gorm.DB) error {
		// 尝试通过手机号或身份证号查找已存在员工
		var existingEmp *Employee
		if phoneHash != "" {
			var emp Employee
			if err := tx.Scopes(middleware.TenantScope(reg.OrgID)).
				Where("phone_hash = ?", phoneHash).First(&emp).Error; err == nil {
				existingEmp = &emp
			}
		}
		if existingEmp == nil && idCardHash != "" {
			var emp Employee
			if err := tx.Scopes(middleware.TenantScope(reg.OrgID)).
				Where("id_card_hash = ?", idCardHash).First(&emp).Error; err == nil {
				existingEmp = &emp
			}
		}

		if existingEmp != nil {
			// 已存在员工：以最新版本覆盖更新
			updates := map[string]interface{}{}
			if req.Address != "" {
				updates["address"] = req.Address
			}
			if phoneEncrypted != "" {
				updates["phone_encrypted"] = phoneEncrypted
				updates["phone_hash"] = phoneHash
			}
			if idCardEncrypted != "" {
				updates["id_card_encrypted"] = idCardEncrypted
				updates["id_card_hash"] = idCardHash

				// 从身份证提取性别和出生日期
				if len(req.IDCard) == 18 {
					gender, birthDate, err := extractFromIDCard(req.IDCard)
					if err == nil {
						updates["gender"] = gender
						updates["birth_date"] = &birthDate
					}
				}
			}
			if bankAccountEncrypted != "" {
				updates["bank_account_encrypted"] = bankAccountEncrypted
				updates["bank_account_hash"] = bankAccountHash
			}
			if req.BankName != "" {
				updates["bank_name"] = req.BankName
			}
			if emergencyPhoneEncrypted != "" {
				updates["emergency_phone_encrypted"] = emergencyPhoneEncrypted
				updates["emergency_phone_hash"] = emergencyPhoneHash
			}
			if req.EmergencyContact != "" {
				updates["emergency_contact"] = req.EmergencyContact
			}

			if len(updates) > 0 {
				updates["updated_by"] = reg.CreatedBy
				if err := tx.Model(&Employee{}).Scopes(middleware.TenantScope(reg.OrgID)).
					Where("id = ?", existingEmp.ID).Updates(updates).Error; err != nil {
					return fmt.Errorf("更新员工信息失败: %w", err)
				}
			}
		} else {
			// 不存在：创建新员工记录
			emp := &Employee{}
			emp.OrgID = reg.OrgID
			emp.Name = reg.Name
			emp.PhoneEncrypted = phoneEncrypted
			emp.PhoneHash = phoneHash
			emp.IDCardEncrypted = idCardEncrypted
			emp.IDCardHash = idCardHash
			emp.Position = reg.Position
			emp.DepartmentID = reg.DepartmentID
			emp.Address = req.Address
			emp.BankName = req.BankName
			emp.BankAccountEncrypted = bankAccountEncrypted
			emp.BankAccountHash = bankAccountHash
			emp.EmergencyContact = req.EmergencyContact
			emp.EmergencyPhoneEncrypted = emergencyPhoneEncrypted
			emp.EmergencyPhoneHash = emergencyPhoneHash
			emp.Status = StatusPending
			emp.CreatedBy = reg.CreatedBy
			emp.UpdatedBy = reg.CreatedBy

			if reg.HireDate != nil {
				emp.HireDate = *reg.HireDate
			}

			// 从身份证提取性别和出生日期
			if len(req.IDCard) == 18 {
				gender, birthDate, err := extractFromIDCard(req.IDCard)
				if err == nil {
					emp.Gender = gender
					emp.BirthDate = &birthDate
				}
			}

			if err := tx.Create(emp).Error; err != nil {
				return fmt.Errorf("创建员工失败: %w", err)
			}
		}

		// 更新 Registration 状态
		now := time.Now()
		if err := tx.Model(&Registration{}).Where("token = ?", token).Updates(map[string]interface{}{
			"status":  RegistrationStatusUsed,
			"used_at": &now,
		}).Error; err != nil {
			return fmt.Errorf("更新登记状态失败: %w", err)
		}

		return nil
	})
}

// ListRegistrations 查询登记列表
func (s *RegistrationService) ListRegistrations(orgID int64, params RegistrationListQueryParams) ([]RegistrationResponse, int64, error) {
	registrations, total, err := s.regRepo.List(orgID, params)
	if err != nil {
		return nil, 0, fmt.Errorf("查询登记列表失败: %w", err)
	}

	items := make([]RegistrationResponse, 0, len(registrations))
	for _, reg := range registrations {
		item := RegistrationResponse{
			ID:        reg.ID,
			EmployeeID: reg.EmployeeID,
			Token:     reg.Token,
			Status:    reg.Status,
			ExpiresAt: reg.ExpiresAt,
			UsedAt:    reg.UsedAt,
			CreatedAt: reg.CreatedAt,
		}

		// 查询员工姓名
		if reg.EmployeeID != nil {
			emp, err := s.empRepo.FindByID(orgID, *reg.EmployeeID)
			if err == nil {
				item.EmployeeName = emp.Name
			}
		}
		if item.EmployeeName == "" {
			item.EmployeeName = reg.Name
		}

		// 查询部门名称
		if reg.DepartmentID != nil {
			deptName, err := s.findDepartmentName(*reg.DepartmentID)
			if err == nil {
				item.DepartmentName = deptName
			}
		}

		items = append(items, item)
	}

	return items, total, nil
}

// DeleteRegistration 删除登记记录（仅 pending 状态可删除）
func (s *RegistrationService) DeleteRegistration(orgID, id int64) error {
	_, err := s.regRepo.FindByID(orgID, id)
	if err != nil {
		return ErrRegistrationNotFound
	}

	if err := s.regRepo.Delete(orgID, id); err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("仅待填写的登记表可以删除")
		}
		return fmt.Errorf("删除登记表失败: %w", err)
	}

	return nil
}

// findDepartmentName 查询部门名称
func (s *RegistrationService) findDepartmentName(departmentID int64) (string, error) {
	type deptResult struct {
		Name string
	}
	var result deptResult
	err := s.regRepo.DB().Table("departments").Where("id = ?", departmentID).Select("name").Scan(&result).Error
	if err != nil {
		return "", fmt.Errorf("查询部门名称失败: %w", err)
	}
	return result.Name, nil
}
