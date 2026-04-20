package salary

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/go-co-op/gocron/v2"
	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// SalaryScheduler 工资模块定时任务调度器（D-13-08 每日9点提醒未确认工资单）
type SalaryScheduler struct {
	db        *gorm.DB
	redisAddr string
}

// NewSalaryScheduler 创建工资模块调度器
func NewSalaryScheduler(db *gorm.DB, redisAddr string) *SalaryScheduler {
	return &SalaryScheduler{db: db, redisAddr: redisAddr}
}

// Start 启动定时任务（D-13-08 每日9点执行）
func (s *SalaryScheduler) Start() (gocron.Scheduler, error) {
	opts := []gocron.SchedulerOption{
		gocron.WithLocation(time.FixedZone("CST", 8*3600)),
	}

	// 尝试使用 Redis 分布式锁
	rdb := redis.NewClient(&redis.Options{Addr: s.redisAddr})
	locker := newSalaryRedisLocker(rdb)
	opts = append(opts, gocron.WithDistributedLocker(locker))

	sched, err := gocron.NewScheduler(opts...)
	if err != nil {
		return nil, fmt.Errorf("create scheduler: %w", err)
	}

	// 每日9:00 执行（D-13-08）
	_, err = sched.NewJob(
		gocron.DailyJob(1, gocron.NewAtTimes(gocron.NewAtTime(9, 0, 0))),
		gocron.NewTask(func() {
			s.enqueueRemindUnconfirmedTasks()
		}),
		gocron.WithName("salary-remind-unconfirmed"),
	)
	if err != nil {
		return nil, fmt.Errorf("add remind-unconfirmed task: %w", err)
	}

	sched.Start()
	log.Printf("[SalaryScheduler] started: salary-remind-unconfirmed at 09:00 CST")
	return sched, nil
}

// enqueueRemindUnconfirmedTasks 查询所有企业，为每个企业创建待确认提醒任务
func (s *SalaryScheduler) enqueueRemindUnconfirmedTasks() {
	// 计算上月年月
	now := time.Now()
	year := now.Year()
	month := int(now.Month())
	if month == 1 {
		year--
		month = 12
	} else {
		month--
	}

	// 查询所有活跃企业
	var orgIDs []int64
	err := s.db.Table("organizations").Where("deleted_at IS NULL").Pluck("id", &orgIDs).Error
	if err != nil {
		log.Printf("[SalaryScheduler] 查询企业列表失败: %v", err)
		return
	}

	client := asynq.NewClient(asynq.RedisClientOpt{Addr: s.redisAddr})
	defer client.Close()

	queued := 0
	for _, orgID := range orgIDs {
		task, err := NewRemindUnconfirmedTask(orgID, year, month)
		if err != nil {
			log.Printf("[SalaryScheduler] 创建任务失败 org=%d: %v", orgID, err)
			continue
		}
		_, err = client.Enqueue(task, asynq.MaxRetry(1))
		if err != nil {
			log.Printf("[SalaryScheduler] 入队失败 org=%d: %v", orgID, err)
			continue
		}
		queued++
	}
	log.Printf("[SalaryScheduler] 提醒任务入队完成: 共 %d 个企业, 年月=%d-%02d", queued, year, month)
}

// salaryRedisLocker 基于 Redis 的分布式锁实现 gocron.Locker 接口
type salaryRedisLocker struct {
	rdb *redis.Client
}

func newSalaryRedisLocker(rdb *redis.Client) *salaryRedisLocker {
	return &salaryRedisLocker{rdb: rdb}
}

func (l *salaryRedisLocker) Lock(ctx context.Context, key string) (gocron.Lock, error) {
	lockKey := fmt.Sprintf("gocron:salary:%s", key)
	ok, err := l.rdb.SetNX(ctx, lockKey, "1", 10*time.Minute).Result()
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("lock not acquired: %s", key)
	}
	return &salaryRedisLock{rdb: l.rdb, key: lockKey}, nil
}

type salaryRedisLock struct {
	rdb *redis.Client
	key string
}

func (l *salaryRedisLock) Unlock(context.Context) error {
	return l.rdb.Del(context.Background(), l.key).Err()
}
