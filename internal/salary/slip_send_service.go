package salary

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/hibiken/asynq"
	"github.com/wencai/easyhr/internal/common/crypto"
	"gorm.io/gorm"
)

// slipSendChannelPriority 渠道优先级：miniapp > sms > h5
var slipSendChannelPriority = []string{"miniapp", "sms", "h5"}

// SlipSendService 工资条发送服务（含 asynq 队列处理）
type SlipSendService struct {
	svc      *Service
	asynqCfg AsynqConfig
}

// AsynqConfig asynq 客户端配置
type AsynqConfig struct {
	RedisAddr string
}

// NewSlipSendService 创建工资条发送服务
func NewSlipSendService(svc *Service, redisAddr string) *SlipSendService {
	return &SlipSendService{
		svc: svc,
		asynqCfg: AsynqConfig{
			RedisAddr: redisAddr,
		},
	}
}

// SendAllSlips 批量发送工资条（入队，立即返回）
func (s *SlipSendService) SendAllSlips(orgID, userID int64, year, month int, employeeIDs []int64, channel string) error {
	task, err := NewSlipSendTask(&SlipSendPayload{
		OrgID:       orgID,
		UserID:      userID,
		Year:        year,
		Month:       month,
		EmployeeIDs: employeeIDs,
		Channel:     channel,
	})
	if err != nil {
		return fmt.Errorf("创建任务失败: %w", err)
	}

	client := asynq.NewClient(asynq.RedisClientOpt{Addr: s.asynqCfg.RedisAddr})
	defer client.Close()

	_, err = client.Enqueue(task, asynq.MaxRetry(3))
	if err != nil {
		return fmt.Errorf("入队失败: %w", err)
	}

	return nil
}

// HandleSlipSendTask asynq Worker Handler（后台处理工资条发送）
func HandleSlipSendTask(ctx context.Context, t *asynq.Task) error {
	var payload SlipSendPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("解析任务载荷失败: %w", err)
	}

	log.Printf("[SlipSend] 开始处理: org=%d, year=%d, month=%d, employees=%v, channel=%s",
		payload.OrgID, payload.Year, payload.Month, payload.EmployeeIDs, payload.Channel)

	// 获取工资条发送服务（通过全局 asynqClient 实例）
	svc, err := GetGlobalSlipSendService()
	if err != nil {
		return fmt.Errorf("获取服务实例失败: %w", err)
	}

	return svc.processSlipSend(ctx, &payload)
}

// processSlipSend 处理工资条发送
func (s *SlipSendService) processSlipSend(ctx context.Context, payload *SlipSendPayload) error {
	orgID, userID := payload.OrgID, payload.UserID
	year, month := payload.Year, payload.Month
	channel := payload.Channel

	// 1. 获取员工列表
	var employees []EmployeeInfo
	var err error
	if len(payload.EmployeeIDs) > 0 {
		// 指定员工
		for _, empID := range payload.EmployeeIDs {
			emp, e := s.svc.empProvider.GetEmployee(orgID, empID)
			if e != nil {
				log.Printf("[SlipSend] 获取员工 %d 失败: %v", empID, e)
				continue
			}
			employees = append(employees, *emp)
		}
	} else {
		// 全员
		employees, err = s.svc.empProvider.GetActiveEmployees(orgID)
		if err != nil {
			return fmt.Errorf("获取员工列表失败: %w", err)
		}
	}

	if len(employees) == 0 {
		return fmt.Errorf("没有需要发送的员工")
	}

	// 2. 逐个处理
	successCount := 0
	failCount := 0

	for _, emp := range employees {
		result := s.sendSlipToEmployee(orgID, userID, &emp, year, month, channel)
		if result.Success {
			successCount++
		} else {
			failCount++
			log.Printf("[SlipSend] 员工 %s 发送失败: %s", emp.Name, result.ErrorMessage)
		}
	}

	log.Printf("[SlipSend] 完成: 成功=%d, 失败=%d", successCount, failCount)
	return nil
}

// SlipSendResult 单个员工发送结果
type SlipSendResult struct {
	Success      bool
	EmployeeID   int64
	EmployeeName string
	Token        string
	ErrorMessage string
}

// sendSlipToEmployee 向单个员工发送工资条
func (s *SlipSendService) sendSlipToEmployee(orgID, userID int64, emp *EmployeeInfo, year, month int, channel string) SlipSendResult {
	// 1. 查找工资记录
	record, err := s.svc.repo.FindPayrollRecordByEmployeeMonth(orgID, emp.ID, year, month)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return SlipSendResult{Success: false, EmployeeID: emp.ID, EmployeeName: emp.Name, ErrorMessage: "该月工资记录不存在"}
		}
		return SlipSendResult{Success: false, EmployeeID: emp.ID, EmployeeName: emp.Name, ErrorMessage: fmt.Sprintf("查询工资记录失败: %v", err)}
	}

	// 2. 校验状态
	if record.Status != PayrollStatusConfirmed && record.Status != PayrollStatusPaid {
		return SlipSendResult{Success: false, EmployeeID: emp.ID, EmployeeName: emp.Name, ErrorMessage: fmt.Sprintf("工资状态为 %s，无法发送", record.Status)}
	}

	// 3. 检查是否已存在工资单
	existingSlip, err := s.findExistingSlip(orgID, record.ID)
	if err == nil && existingSlip != nil {
		// 已存在，写发送日志（使用已有 token）
		s.writeSendLog(orgID, record.ID, emp.ID, channel, SlipLogStatusSent, "")
		return SlipSendResult{Success: true, EmployeeID: emp.ID, EmployeeName: emp.Name, Token: existingSlip.Token}
	}

	// 4. 生成 token
	token, err := generateSlipToken()
	if err != nil {
		return SlipSendResult{Success: false, EmployeeID: emp.ID, EmployeeName: emp.Name, ErrorMessage: "生成 token 失败"}
	}

	// 5. 加密手机号
	phoneEncrypted, err := crypto.Encrypt(emp.Phone, []byte(s.svc.cryptoCfg.AESKey))
	if err != nil {
		return SlipSendResult{Success: false, EmployeeID: emp.ID, EmployeeName: emp.Name, ErrorMessage: fmt.Sprintf("加密手机号失败: %v", err)}
	}
	phoneHash := crypto.HashSHA256(emp.Phone)

	// 6. 创建工资单
	now := time.Now()
	slip := PayrollSlip{
		PayrollRecordID: record.ID,
		EmployeeID:      emp.ID,
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

	if err := s.svc.repo.CreateSlip(&slip); err != nil {
		return SlipSendResult{Success: false, EmployeeID: emp.ID, EmployeeName: emp.Name, ErrorMessage: fmt.Sprintf("创建工资单失败: %v", err)}
	}

	// 7. 写发送日志
	s.writeSendLog(orgID, record.ID, emp.ID, channel, SlipLogStatusSent, "")

	return SlipSendResult{Success: true, EmployeeID: emp.ID, EmployeeName: emp.Name, Token: token}
}

// writeSendLog 写工资条发送日志
func (s *SlipSendService) writeSendLog(orgID, recordID, employeeID int64, channel, status, errorMsg string) {
	log := SalarySlipSendLog{
		OrgID:           orgID,
		PayrollRecordID: recordID,
		EmployeeID:      employeeID,
		Channel:         channel,
		Status:          status,
	}
	if status == SlipLogStatusSent {
		now := time.Now()
		log.SentAt = &now
	}
	if errorMsg != "" {
		log.ErrorMessage = errorMsg
	}
	_ = s.svc.repo.CreateSlipSendLog(&log)
}

// findExistingSlip 查找已存在的工资单
func (s *SlipSendService) findExistingSlip(orgID int64, recordID int64) (*PayrollSlip, error) {
	var slips []PayrollSlip
	err := s.svc.repo.db.Scopes(middlewareTenantScopeLocal(orgID)).
		Where("payroll_record_id = ?", recordID).
		Find(&slips).Error
	if err != nil {
		return nil, err
	}
	if len(slips) == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	return &slips[0], nil
}

// middlewareTenantScopeLocal 本地租户 scope
func middlewareTenantScopeLocal(orgID int64) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("org_id = ?", orgID)
	}
}

// 发送日志状态常量
const (
	SlipLogStatusPending  = "pending"
	SlipLogStatusSending  = "sending"
	SlipLogStatusSent     = "sent"
	SlipLogStatusFailed   = "failed"
)

// ===== 全局 asynq 服务实例注册 =====

var globalSlipSendService *SlipSendService

// RegisterGlobalSlipSendService 注册全局实例（供 asynq worker 使用）
func RegisterGlobalSlipSendService(svc *SlipSendService) {
	globalSlipSendService = svc
}

// GetGlobalSlipSendService 获取全局实例
func GetGlobalSlipSendService() (*SlipSendService, error) {
	if globalSlipSendService == nil {
		return nil, fmt.Errorf("SlipSendService 未注册")
	}
	return globalSlipSendService, nil
}

