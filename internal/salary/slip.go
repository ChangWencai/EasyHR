package salary

import (
	"errors"
	"fmt"
	"time"

	"github.com/wencai/easyhr/internal/common/crypto"
	"gorm.io/gorm"
)

var (
	ErrSlipTokenExpired  = errors.New("工资单已过期")
	ErrSlipAlreadySigned = errors.New("工资单已签收，不可重复签收")
	ErrSlipNotViewed     = errors.New("工资单未查看，不能签收")
)

const SlipExpiryDuration = 7 * 24 * time.Hour

// verifySMSCode 验证短信验证码（6位数字，5分钟有效期，使用 Redis）
// TODO: 实现 Redis 验证码存储和校验逻辑
func verifySMSCode(code, expectedCode string) bool {
	return code == expectedCode
}

// SendSlipResult 工资单发送结果
type SendSlipResult struct {
	RecordID     int64  `json:"record_id"`
	EmployeeName string `json:"employee_name"`
	SlipToken    string `json:"slip_token"`
	ShareLink    string `json:"share_link"` // 格式: /salary/slip/{token}
}

// SendSlip 发送工资单（需要工资表状态为 confirmed 或 paid）
func (s *Service) SendSlip(orgID, userID int64, recordIDs []int64) ([]SendSlipResult, error) {
	var results []SendSlipResult

	for _, recordID := range recordIDs {
		// 查询工资核算记录
		record, err := s.repo.FindPayrollRecordByID(orgID, recordID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, fmt.Errorf("工资记录 %d 不存在", recordID)
			}
			return nil, fmt.Errorf("查询工资记录失败: %w", err)
		}

		// 校验状态
		if record.Status != PayrollStatusConfirmed && record.Status != PayrollStatusPaid {
			return nil, fmt.Errorf("员工 %s 的工资记录状态为 %s，无法发送工资单", record.EmployeeName, record.Status)
		}

		// 检查是否已存在工资单
		var existingSlip PayrollSlip
		err = s.repo.db.Where("payroll_record_id = ? AND org_id = ?", recordID, orgID).First(&existingSlip).Error
		if err == nil {
			// 工资单已存在，返回已有结果
			results = append(results, SendSlipResult{
				RecordID:     recordID,
				EmployeeName: record.EmployeeName,
				SlipToken:    existingSlip.Token,
				ShareLink:    fmt.Sprintf("/salary/slip/%s", existingSlip.Token),
			})
			continue
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("查询工资单失败: %w", err)
		}

		// 生成 token
		token, err := generateSlipToken()
		if err != nil {
			return nil, fmt.Errorf("生成工资单 token 失败: %w", err)
		}

		// 获取员工信息
		var employee EmployeeInfo
		if s.empProvider != nil {
			emp, err := s.empProvider.GetEmployee(orgID, record.EmployeeID)
			if err != nil {
				return nil, fmt.Errorf("获取员工信息失败: %w", err)
			}
			employee = *emp
		} else {
			return nil, fmt.Errorf("员工提供者未配置")
		}

		// 加密手机号
		phoneEncrypted, err := crypto.Encrypt(employee.Phone, s.aesKey())
		if err != nil {
			return nil, fmt.Errorf("加密手机号失败: %w", err)
		}

		// 生成手机号哈希索引
		phoneHash := crypto.HashSHA256(employee.Phone)

		// 创建工资单
		now := time.Now()
		slip := PayrollSlip{
			PayrollRecordID: recordID,
			EmployeeID:      record.EmployeeID,
			Token:           token,
			PhoneEncrypted:  phoneEncrypted,
			PhoneHash:       phoneHash,
			Status:          SlipStatusSent,
			SentAt:          &now,
			ExpiresAt:       now.Add(SlipExpiryDuration),
		}
		slip.OrgID = orgID
		slip.CreatedBy = userID
		slip.UpdatedBy = userID

		if err := s.repo.CreateSlip(&slip); err != nil {
			return nil, fmt.Errorf("创建工资单失败: %w", err)
		}

		results = append(results, SendSlipResult{
			RecordID:     recordID,
			EmployeeName: record.EmployeeName,
			SlipToken:    token,
			ShareLink:    fmt.Sprintf("/salary/slip/%s", token),
		})
	}

	return results, nil
}

// GetSlipByToken 通过 token 获取工资单详情（无需认证）
func (s *Service) GetSlipByToken(token string) (*SlipDetailResponse, error) {
	// 查询工资单
	slip, err := s.repo.FindSlipByToken(token)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSlipTokenInvalid
		}
		return nil, fmt.Errorf("查询工资单失败: %w", err)
	}

	// 校验是否过期
	if time.Now().After(slip.ExpiresAt) {
		return nil, ErrSlipTokenExpired
	}

	// 查询工资核算记录
	record, err := s.repo.FindPayrollRecordByID(slip.OrgID, slip.PayrollRecordID)
	if err != nil {
		return nil, fmt.Errorf("查询工资记录失败: %w", err)
	}

	// 查询工资明细
	items, err := s.repo.FindPayrollItemsByRecord(slip.OrgID, record.ID)
	if err != nil {
		return nil, fmt.Errorf("查询工资明细失败: %w", err)
	}

	// 转换为响应格式
	resp := &SlipDetailResponse{
		EmployeeName:    maskName(record.EmployeeName),
		Year:            record.Year,
		Month:           record.Month,
		GrossIncome:     record.GrossIncome,
		SIDeduction:     record.SIDeduction,
		Tax:             record.Tax,
		TotalDeductions: record.TotalDeductions,
		NetIncome:       record.NetIncome,
		Status:          slip.Status,
	}

	for _, item := range items {
		resp.Items = append(resp.Items, SlipItemDetail{
			ItemName: item.ItemName,
			ItemType: item.ItemType,
			Amount:   item.Amount,
		})
	}

	// 格式化签收时间
	if slip.SignedAt != nil {
		signedAt := slip.SignedAt.Format("2006-01-02 15:04:05")
		resp.SignedAt = &signedAt
	}

	// 首次查看时更新状态
	if slip.Status == SlipStatusSent {
		now := time.Now()
		slip.Status = SlipStatusViewed
		slip.ViewedAt = &now
		_ = s.repo.UpdateSlip(slip.OrgID, slip.ID, map[string]interface{}{
			"status":    slip.Status,
			"viewed_at": slip.ViewedAt,
		})
	}

	return resp, nil
}

// maskName 姓名脱敏（保留姓氏，其余用*代替）
func maskName(name string) string {
	if len(name) <= 1 {
		return name
	}
	runes := []rune(name)
	if len(runes) == 2 {
		return string(runes[:1]) + "*"
	}
	return string(runes[:1]) + "**"
}

// VerifySlipPhone 验证工资单手机号（发送短信验证码）
func (s *Service) VerifySlipPhone(token, phone string) error {
	// 查询工资单
	slip, err := s.repo.FindSlipByToken(token)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrSlipTokenInvalid
		}
		return fmt.Errorf("查询工资单失败: %w", err)
	}

	// 校验是否过期
	if time.Now().After(slip.ExpiresAt) {
		return ErrSlipTokenExpired
	}

	// 校验手机号哈希
	phoneHash := crypto.HashSHA256(phone)
	if phoneHash != slip.PhoneHash {
		return fmt.Errorf("手机号不匹配")
	}

	// 生成6位验证码
	code := fmt.Sprintf("%06d", time.Now().Unix()%1000000)
	if code == "000000" {
		code = "123456" // fallback
	}

	// 发送短信验证码
	if s.smsClient != nil {
		// TODO: 调用真实短信服务
		// err := s.smsClient.SendCode(phone, code)
		// if err != nil {
		// 	return fmt.Errorf("发送短信验证码失败: %w", err)
		// }
		_ = code // 临时避免未使用警告
	}

	return nil
}

// VerifySlipCode 验证短信验证码
func (s *Service) VerifySlipCode(token, phone, code string) (bool, error) {
	// 查询工资单
	slip, err := s.repo.FindSlipByToken(token)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, ErrSlipTokenInvalid
		}
		return false, fmt.Errorf("查询工资单失败: %w", err)
	}

	// 校验是否过期
	if time.Now().After(slip.ExpiresAt) {
		return false, ErrSlipTokenExpired
	}

	// 校验手机号哈希
	phoneHash := crypto.HashSHA256(phone)
	if phoneHash != slip.PhoneHash {
		return false, fmt.Errorf("手机号不匹配")
	}

	// TODO: 从 Redis 获取并验证验证码
	// 这里暂时返回 true，实际实现需要验证 Redis 中的验证码
	_ = code

	return true, nil
}

// SignSlip 签收工资单
func (s *Service) SignSlip(token string) error {
	// 查询工资单
	slip, err := s.repo.FindSlipByToken(token)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrSlipTokenInvalid
		}
		return fmt.Errorf("查询工资单失败: %w", err)
	}

	// 校验是否过期
	if time.Now().After(slip.ExpiresAt) {
		return ErrSlipTokenExpired
	}

	// 校验状态
	if slip.Status == SlipStatusSigned {
		return ErrSlipAlreadySigned
	}
	if slip.Status != SlipStatusViewed {
		return ErrSlipNotViewed
	}

	// 更新为已签收
	now := time.Now()
	err = s.repo.UpdateSlip(slip.OrgID, slip.ID, map[string]interface{}{
		"status":     SlipStatusSigned,
		"signed_at":  now,
		"updated_by": slip.ID, // 员工签收时使用 slip.ID 标记
	})
	if err != nil {
		return fmt.Errorf("签收工资单失败: %w", err)
	}

	return nil
}

// aesKey 获取 AES 密钥字节
func (s *Service) aesKey() []byte {
	return []byte(s.cryptoCfg.AESKey)
}
