package socialinsurance

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-co-op/gocron/v2"
	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
	"github.com/shopspring/decimal"
)

// asynq 任务类型常量（D-SI-02/D-SI-03）
const (
	TypeGenerateMonthlyPayments = "si:generate_monthly" // 生成下月缴费记录
	TypeCheckPaymentStatus      = "si:check_status"     // 状态流转检查
)

// MonthlyPaymentPayload 月度记录生成任务载荷
type MonthlyPaymentPayload struct {
	OrgID int64  `json:"org_id"`
	Month string `json:"month"` // YYYYMM
}

// CheckStatusPayload 状态流转检查任务载荷
type CheckStatusPayload struct {
	OrgID int64  `json:"org_id"`
	Month string `json:"month"` // YYYYMM
}

// NewGenerateMonthlyPaymentsTask 创建月度记录生成任务
func NewGenerateMonthlyPaymentsTask(orgID int64, month string) (*asynq.Task, error) {
	payload, err := json.Marshal(MonthlyPaymentPayload{OrgID: orgID, Month: month})
	if err != nil {
		return nil, fmt.Errorf("marshal payload: %w", err)
	}
	return asynq.NewTask(TypeGenerateMonthlyPayments, payload), nil
}

// NewCheckPaymentStatusTask 创建状态流转检查任务
func NewCheckPaymentStatusTask(orgID int64, month string) (*asynq.Task, error) {
	payload, err := json.Marshal(CheckStatusPayload{OrgID: orgID, Month: month})
	if err != nil {
		return nil, fmt.Errorf("marshal payload: %w", err)
	}
	return asynq.NewTask(TypeCheckPaymentStatus, payload), nil
}

// MonthlyPaymentWorker 月度缴费记录 asynq Worker
type MonthlyPaymentWorker struct {
	paymentRepo *SIMonthlyPaymentRepository
	recordRepo  *Repository
}

// NewMonthlyPaymentWorker 创建月度缴费 Worker
func NewMonthlyPaymentWorker(paymentRepo *SIMonthlyPaymentRepository, recordRepo *Repository) *MonthlyPaymentWorker {
	return &MonthlyPaymentWorker{paymentRepo: paymentRepo, recordRepo: recordRepo}
}

// HandleGenerateMonthlyPayments 处理月度记录生成任务（D-SI-02）
// 每天凌晨 02:00 触发，生成下月 SIMonthlyPayment 记录
func (w *MonthlyPaymentWorker) HandleGenerateMonthlyPayments(ctx context.Context, t *asynq.Task) error {
	var payload MonthlyPaymentPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("unmarshal payload: %w", err)
	}

	// 1. 查询该组织所有 active 参保记录
	records, _, err := w.recordRepo.ListRecords(payload.OrgID, SIStatusActive, "", 1, 1000)
	if err != nil {
		return fmt.Errorf("query active records: %w", err)
	}

	if len(records) == 0 {
		return nil
	}

	// 2. 为每个 active 员工生成下月缴费记录
	payments := make([]SIMonthlyPayment, 0, len(records))
	for _, r := range records {
		payment := SIMonthlyPayment{
			EmployeeID:     uint(r.EmployeeID),
			YearMonth:      payload.Month,
			Status:         PaymentStatusPending,
			PaymentChannel: SIPayChannelSelf,
			CompanyAmount:  decimal.NewFromFloat(r.TotalCompany),
			PersonalAmount: decimal.NewFromFloat(r.TotalPersonal),
			TotalAmount:    decimal.NewFromFloat(r.TotalCompany + r.TotalPersonal),
		}
		payment.OrgID = payload.OrgID
		payments = append(payments, payment)
	}

	// 3. 幂等 UPSERT（D-SI-02：ON CONFLICT DO NOTHING）
	if err := w.paymentRepo.BatchUpsert(ctx, nil, payments); err != nil {
		return fmt.Errorf("batch upsert monthly payments: %w", err)
	}

	return nil
}

// HandleCheckPaymentStatus 处理状态流转检查任务（D-SI-03）
// 每天凌晨 02:05 触发
// >=26日：pending -> overdue
// <26日：已确认支付 -> normal
func (w *MonthlyPaymentWorker) HandleCheckPaymentStatus(ctx context.Context, t *asynq.Task) error {
	var payload CheckStatusPayload
	if err := json.Unmarshal(t.Payload(), &payload); err != nil {
		return fmt.Errorf("unmarshal payload: %w", err)
	}

	cstZone := time.FixedZone("CST", 8*3600)
	today := time.Now().In(cstZone)

	// D-SI-03 决策：每月26日为状态更新分界点（与各地实际截止日无关）
	cutoffDay := 26

	if today.Day() >= cutoffDay {
		// >=26日：所有 pending 且未缴的 -> overdue
		if err := w.paymentRepo.UpdateOverduePayments(ctx, payload.OrgID, payload.Month); err != nil {
			return fmt.Errorf("update overdue payments: %w", err)
		}
	}
	// <26日：不做自动流转，等用户确认或 webhook 回调

	return nil
}

// redisLocker 基于 Redis 的分布式锁实现 gocron.Locker 接口
type redisLocker struct {
	rdb   *redis.Client
	prefix string
}

// redisLock 基于 Redis 的分布式锁实现 gocron.Lock 接口
type redisLock struct {
	rdb   *redis.Client
	key   string
}

// newRedisLocker 创建 Redis 分布式锁
func newRedisLocker(rdb *redis.Client, prefix string) *redisLocker {
	return &redisLocker{rdb: rdb, prefix: prefix}
}

// Lock 实现 gocron.Locker 接口
func (l *redisLocker) Lock(ctx context.Context, key string) (gocron.Lock, error) {
	lockKey := l.prefix + key
	// 尝试获取锁，TTL 60 秒
	ok, err := l.rdb.SetNX(ctx, lockKey, "locked", 60*time.Second).Result()
	if err != nil {
		return nil, fmt.Errorf("redis lock: %w", err)
	}
	if !ok {
		return nil, fmt.Errorf("job already locked: %s", key)
	}
	return &redisLock{rdb: l.rdb, key: lockKey}, nil
}

// Unlock 实现 gocron.Lock 接口
func (l *redisLock) Unlock(ctx context.Context) error {
	return l.rdb.Del(ctx, l.key).Err()
}

// StartScheduler 启动社保缴费提醒定时任务
// rdb 为 nil 时不使用分布式锁（开发环境）
func StartScheduler(rdb *redis.Client, svc *Service) (gocron.Scheduler, error) {
	opts := []gocron.SchedulerOption{
		gocron.WithLocation(time.FixedZone("CST", 8*3600)),
	}

	// Redis 可用时启用分布式锁
	if rdb != nil {
		locker := newRedisLocker(rdb, "easyhr:social:")
		opts = append(opts, gocron.WithDistributedLocker(locker))
	}

	s, err := gocron.NewScheduler(opts...)
	if err != nil {
		return nil, fmt.Errorf("create scheduler: %w", err)
	}

	// 每日 08:00 CST 扫描缴费到期提醒
	_, err = s.NewJob(
		gocron.DailyJob(1, gocron.NewAtTimes(gocron.NewAtTime(8, 0, 0))),
		gocron.NewTask(func() {
			svc.CheckPaymentDueReminders()
		}),
		gocron.WithName("social-insurance-payment-due-check"),
	)
	if err != nil {
		return nil, fmt.Errorf("create payment due job: %w", err)
	}

	s.Start()
	return s, nil
}
