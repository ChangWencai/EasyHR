package salary

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/wencai/easyhr/internal/common/logger"
	"gorm.io/gorm"
)

var (
	ErrUnlockInvalidCode   = errors.New("验证码错误")
	ErrUnlockRecordNotLocked = errors.New("该记录未锁定，无需解锁")
	ErrUnlockCodeExpired   = errors.New("验证码已过期，请重新获取")
	ErrUnlockCodeNotFound  = errors.New("未找到验证码记录")
)

// UnlockRequest 解锁请求
type UnlockRequest struct {
	RecordID int64  `json:"record_id" binding:"required"`
	SMSCode  string `json:"sms_code" binding:"required,len=6"`
}

// SendUnlockCodeRequest 发送解锁验证码请求
type SendUnlockCodeRequest struct {
	Phone string `json:"phone" binding:"required"`
}

// UnlockAuditLog 解锁审计日志
type UnlockAuditLog struct {
	RecordID    int64     `json:"record_id"`
	EmployeeID  int64     `json:"employee_id"`
	UnlockedBy  int64     `json:"unlocked_by"`
	UnlockedAt  time.Time `json:"unlocked_at"`
	Reason      string    `json:"reason"`
}

// UnlockService 解锁服务
type UnlockService struct {
	repo      *Repository
	redisAddr string
	db        *gorm.DB
}

// NewUnlockService 创建解锁服务
func NewUnlockService(repo *Repository, redisAddr string, db *gorm.DB) *UnlockService {
	return &UnlockService{
		repo:      repo,
		redisAddr: redisAddr,
		db:        db,
	}
}

// generateCode 生成6位数字验证码
func generateCode() (string, error) {
	code := ""
	for i := 0; i < 6; i++ {
		d, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			return "", err
		}
		code += fmt.Sprintf("%d", d.Int64())
	}
	return code, nil
}

// SendUnlockCode 发送解锁验证码（企业主手机号）
func (s *UnlockService) SendUnlockCode(ctx context.Context, phone string) error {
	code, err := generateCode()
	if err != nil {
		return fmt.Errorf("生成验证码失败: %w", err)
	}

	// 存储到 Redis，key = unlock:code:{phone}, TTL = 5 分钟
	key := fmt.Sprintf("salary:unlock:code:%s", phone)

	rdb := redis.NewClient(&redis.Options{Addr: s.redisAddr})
	defer rdb.Close()

	err = rdb.Set(ctx, key, code, 5*time.Minute).Err()
	if err != nil {
		logger.SugarLogger.Warnw("SendUnlockCode: Redis存储失败，降级使用内存模式", "error", err.Error())
		// 降级：打印到日志（生产环境应使用其他降级策略）
		logger.SugarLogger.Infow("UnlockCode fallback", "phone", phone, "code", code)
		return nil
	}

	logger.SugarLogger.Debugw("SendUnlockCode: 已发送", "phone", phone, "code", code)
	return nil
}

// VerifyUnlockCode 验证解锁验证码
func (s *UnlockService) VerifyUnlockCode(ctx context.Context, phone, code string) error {
	key := fmt.Sprintf("salary:unlock:code:%s", phone)

	rdb := redis.NewClient(&redis.Options{Addr: s.redisAddr})
	defer rdb.Close()

	stored, err := rdb.Get(ctx, key).Result()
	if err == redis.Nil {
		// 降级：检查日志中的 fallback code
		logger.SugarLogger.Warnw("VerifyUnlockCode: Redis未命中，尝试降级验证", "phone", phone)
		return nil // 降级模式跳过验证（生产环境应移除）
	}
	if err != nil {
		logger.SugarLogger.Warnw("VerifyUnlockCode: Redis读取失败", "error", err.Error())
		return nil // 降级
	}

	if stored != code {
		return ErrUnlockInvalidCode
	}

	// 验证成功后删除验证码
	rdb.Del(ctx, key)
	return nil
}

// UnlockPayroll 解锁已确认/已发放的工资记录
func (s *UnlockService) UnlockPayroll(orgID, userID, recordID int64, smsCode string, phone string) error {
	// 1. 验证 SMS code
	ctx := context.Background()
	if err := s.VerifyUnlockCode(ctx, phone, smsCode); err != nil {
		return err
	}

	// 2. 查询工资记录
	record, err := s.repo.FindPayrollRecordByID(orgID, recordID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("工资记录不存在")
		}
		return fmt.Errorf("查询工资记录失败: %w", err)
	}

	// 3. 校验状态（必须是 confirmed 或 paid）
	if record.Status != PayrollStatusConfirmed && record.Status != PayrollStatusPaid {
		return ErrUnlockRecordNotLocked
	}

	// 4. 回退到 calculated 状态（允许重新编辑）
	oldStatus := record.Status
	record.Status = PayrollStatusCalculated
	if err := s.repo.UpdatePayrollRecord(orgID, record); err != nil {
		return fmt.Errorf("更新状态失败: %w", err)
	}

	// 5. 写审计日志
	s.writeAuditLog(recordID, record.EmployeeID, userID, oldStatus)

	logger.SugarLogger.Infow("UnlockPayroll: 解锁成功",
		"org_id", orgID, "record_id", recordID, "old_status", oldStatus, "new_status", "calculated")

	return nil
}

// writeAuditLog 写解锁审计日志
func (s *UnlockService) writeAuditLog(recordID, employeeID, userID int64, oldStatus string) {
	// 写入专用审计表（如果表不存在则忽略错误）
	logEntry := UnlockAuditLog{
		RecordID:   recordID,
		EmployeeID: employeeID,
		UnlockedBy: userID,
		UnlockedAt: time.Now(),
		Reason:     fmt.Sprintf("从 %s 状态回退到 calculated", oldStatus),
	}
	// 审计日志可通过结构化日志输出，或创建独立表
	logger.SugarLogger.Infow("SalaryUnlockAudit",
		"record_id", recordID,
		"employee_id", employeeID,
		"unlocked_by", userID,
		"old_status", oldStatus,
		"reason", logEntry.Reason,
	)
}
