package Redis

import (
	"context"
	_ "embed"
	"errors"
	"github.com/redis/go-redis/v9"
	"time"
)

//go:embed lua/fixed_window.lua
var fixedWindow string

// FixedWindowLimiter 基于Redis的分布式限流器
type FixedWindowLimiter struct {
	// Redis客户端
	client redis.Cmdable
	// 窗口内最大的请求数量
	maxCount int64
	// 固定窗口的key过期时间
	expiration time.Duration
}

// NewFixedWindowLimiter 初始化固定窗口限流器，client redis的客户端，maxCount固定窗口内允许的最大请求数量
// expiration 窗口的大小
func NewFixedWindowLimiter(client redis.Cmdable, maxCount int64, expiration time.Duration) *FixedWindowLimiter {
	return &FixedWindowLimiter{
		client:     client,
		maxCount:   maxCount,
		expiration: expiration,
	}
}

// Allow 是否允许通过限流器继续请求，key存储再Redis中的键，可以是单个接口，也可以是服务
func (f FixedWindowLimiter) Allow(ctx context.Context, key string) (bool, error) {
	res, err := f.client.Eval(ctx, fixedWindow, []string{key}, f.maxCount,
		f.expiration.Milliseconds()).Result()
	if err != nil {
		return false, err
	}

	if res.(string) == "true" {
		return false, errors.New("到达性能瓶颈")
	}

	return true, nil
}
