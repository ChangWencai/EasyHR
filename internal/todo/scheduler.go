package todo

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/go-co-op/gocron/v2"
	"github.com/redis/go-redis/v9"
)

// ContractServiceWrapper wraps the contract renewal check to avoid circular imports.
type ContractServiceWrapper interface {
	CheckContractRenewalReminders(ctx context.Context) error
	CheckPendingSignReminders(ctx context.Context) error
}

// Scheduler 定时任务调度器
type Scheduler struct {
	repo            *Repository
	rdb             *redis.Client
	contractService ContractServiceWrapper
}

// NewScheduler 创建调度器（接受 todoRepo 和可选的 redis.Client）
func NewScheduler(repo *Repository, rdb *redis.Client, contractSvc ContractServiceWrapper) *Scheduler {
	return &Scheduler{repo: repo, rdb: rdb, contractService: contractSvc}
}

// cstZone 定义中国时区（+08:00）
var cstZone = time.FixedZone("CST", 8*3600)

// Start 启动调度器
func (s *Scheduler) Start() (gocron.Scheduler, error) {
	opts := []gocron.SchedulerOption{
		gocron.WithLocation(cstZone),
	}

	// Redis 可用时启用分布式锁
	if s.rdb != nil {
		locker := newTodoRedisLocker(s.rdb)
		opts = append(opts, gocron.WithDistributedLocker(locker))
	}

	sched, err := gocron.NewScheduler(opts...)
	if err != nil {
		return nil, fmt.Errorf("create scheduler: %w", err)
	}

	// 每日 02:00 CST -- 扫描 urgency_status
	_, err = sched.NewJob(
		gocron.DailyJob(1, gocron.NewAtTimes(gocron.NewAtTime(2, 0, 0))),
		gocron.NewTask(func() {
			ctx := context.Background()
			count, err := s.repo.ScanUrgencyStatus(ctx)
			if err != nil {
				log.Printf("[todo-scheduler] ScanUrgencyStatus failed: %v", err)
			} else {
				log.Printf("[todo-scheduler] ScanUrgencyStatus updated %d items", count)
			}
		}),
		gocron.WithName("todo-urgency-scan"),
	)
	if err != nil {
		return nil, fmt.Errorf("create urgency scan job: %w", err)
	}

	// 每日 08:00 CST -- 更新轮播图激活状态
	_, err = sched.NewJob(
		gocron.DailyJob(1, gocron.NewAtTimes(gocron.NewAtTime(8, 0, 0))),
		gocron.NewTask(func() {
			ctx := context.Background()
			count, err := s.repo.UpdateCarouselActivation(ctx)
			if err != nil {
				log.Printf("[todo-scheduler] UpdateCarouselActivation failed: %v", err)
			} else {
				log.Printf("[todo-scheduler] UpdateCarouselActivation updated %d items", count)
			}
		}),
		gocron.WithName("todo-carousel-activation"),
	)
	if err != nil {
		return nil, fmt.Errorf("create carousel activation job: %w", err)
	}

	// 每月1日 00:01 CST -- 生成当月个税申报和社保缴费待办
	_, err = sched.NewJob(
		gocron.CronJob("1 0 1 * *", true), // 每月1日00:01
		gocron.NewTask(func() {
			ctx := context.Background()
			if err := s.repo.GenerateMonthlyTodos(ctx); err != nil {
				log.Printf("[todo-scheduler] GenerateMonthlyTodos failed: %v", err)
			} else {
				log.Printf("[todo-scheduler] GenerateMonthlyTodos completed")
			}
		}),
		gocron.WithName("todo-monthly-generate"),
	)
	if err != nil {
		return nil, fmt.Errorf("create monthly generate job: %w", err)
	}

	// 每月5日 00:01 CST -- 生成当月社保增减员待办
	_, err = sched.NewJob(
		gocron.CronJob("1 0 5 * *", true), // 每月5日00:01
		gocron.NewTask(func() {
			ctx := context.Background()
			cst := time.FixedZone("CST", 8*3600)
			today := time.Now().In(cst)
			deadline := time.Date(today.Year(), today.Month(), 20, 23, 59, 59, 0, cst)

			var orgIDs []int64
			s.repo.db.Model(&struct{ ID int64 }{}).Table("organizations").Pluck("id", &orgIDs)

			for _, orgID := range orgIDs {
				todo := &TodoItem{}
				todo.OrgID = orgID
				todo.Title = fmt.Sprintf("%d月社保公积金增减员，请于20日前完成", today.Month())
				todo.Type = TodoTypeSIChange
				todo.Deadline = &deadline
				todo.IsTimeLimited = true
				todo.Status = TodoStatusPending
				todo.UrgencyStatus = UrgencyNormal
				todo.SourceType = "socialinsurance"
				_ = s.repo.CreateTodo(ctx, todo)
			}
			log.Printf("[todo-scheduler] GenerateSIChangeTodos completed")
		}),
		gocron.WithName("todo-si-change-generate"),
	)
	if err != nil {
		return nil, fmt.Errorf("create si-change generate job: %w", err)
	}

	// 每年6月15日 00:01 CST -- 生成年度基数调整待办
	_, err = sched.NewJob(
		gocron.CronJob("1 0 15 6 *", true), // 每年6月15日00:01
		gocron.NewTask(func() {
			ctx := context.Background()
			if err := s.repo.GenerateAnnualBaseTodos(ctx); err != nil {
				log.Printf("[todo-scheduler] GenerateAnnualBaseTodos failed: %v", err)
			} else {
				log.Printf("[todo-scheduler] GenerateAnnualBaseTodos completed")
			}
		}),
		gocron.WithName("todo-annual-base-generate"),
	)
	if err != nil {
		return nil, fmt.Errorf("create annual base generate job: %w", err)
	}

	// 每日 02:05 CST -- 合同续签检查（实际调用 contractService）
	_, err = sched.NewJob(
		gocron.DailyJob(1, gocron.NewAtTimes(gocron.NewAtTime(2, 5, 0))),
		gocron.NewTask(func() {
			ctx := context.Background()
			if s.contractService != nil {
				if err := s.contractService.CheckContractRenewalReminders(ctx); err != nil {
					log.Printf("[todo-scheduler] CheckContractRenewalReminders failed: %v", err)
				} else {
					log.Printf("[todo-scheduler] CheckContractRenewalReminders completed")
				}
			} else {
				log.Printf("[todo-scheduler] contractService not available, skipping renewal check")
			}
		}),
		gocron.WithName("todo-contract-renewal-check"),
	)
	if err != nil {
		return nil, fmt.Errorf("create contract renewal check job: %w", err)
	}

	// 每日 09:00 CST -- 检查合同3天未签提醒
	_, err = sched.NewJob(
		gocron.DailyJob(1, gocron.NewAtTimes(gocron.NewAtTime(9, 0, 0))),
		gocron.NewTask(func() {
			ctx := context.Background()
			if s.contractService != nil {
				if err := s.contractService.CheckPendingSignReminders(ctx); err != nil {
					log.Printf("[todo-scheduler] CheckPendingSignReminders failed: %v", err)
				} else {
					log.Printf("[todo-scheduler] CheckPendingSignReminders completed")
				}
			} else {
				log.Printf("[todo-scheduler] contractService not available, skipping pending sign check")
			}
		}),
		gocron.WithName("contract-pending-sign-reminder"),
	)
	if err != nil {
		return nil, fmt.Errorf("create pending sign reminder job: %w", err)
	}

	sched.Start()
	return sched, nil
}

// todoRedisLocker 基于 Redis 的分布式锁
type todoRedisLocker struct {
	rdb *redis.Client
}

func newTodoRedisLocker(rdb *redis.Client) *todoRedisLocker {
	return &todoRedisLocker{rdb: rdb}
}

func (l *todoRedisLocker) Lock(ctx context.Context, key string) (gocron.Lock, error) {
	lockKey := "easyhr:todo:lock:" + key
	ok, err := l.rdb.SetNX(ctx, lockKey, "locked", 60*time.Second).Result()
	if err != nil {
		return nil, fmt.Errorf("redis lock: %w", err)
	}
	if !ok {
		return nil, fmt.Errorf("job already locked: %s", key)
	}
	return &todoRedisLock{rdb: l.rdb, key: lockKey}, nil
}

type todoRedisLock struct {
	rdb *redis.Client
	key string
}

func (l *todoRedisLock) Unlock(ctx context.Context) error {
	return l.rdb.Del(ctx, l.key).Err()
}
