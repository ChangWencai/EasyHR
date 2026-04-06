package testutil

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

func SetupTestRedis() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       1,
	})
}

func CleanupTestRedis(rdb *redis.Client) {
	ctx := context.Background()
	if rdb != nil {
		rdb.FlushDB(ctx)
	}
}

func WaitForRedis(rdb *redis.Client) bool {
	ctx := context.Background()
	for i := 0; i < 10; i++ {
		if err := rdb.Ping(ctx).Err(); err == nil {
			return true
		}
		time.Sleep(200 * time.Millisecond)
	}
	return false
}
