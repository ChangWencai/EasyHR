package socialinsurance

import (
	"context"
	"fmt"
	"time"

	"github.com/go-co-op/gocron/v2"
	"github.com/redis/go-redis/v9"
)

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
