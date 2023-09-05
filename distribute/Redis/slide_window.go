package Redis

import (
	"context"
	_ "embed"
	"errors"
	"github.com/redis/go-redis/v9"
	"time"
)

//go:embed lua/slide_window.lua
var slideWindow string

// SlideWindowLimiter 基于Redis实现的滑动窗口的限流器
type SlideWindowLimiter struct {
	// Redis客户端
	client redis.Cmdable
	// 窗口内最大的请求数量
	maxCount int64
	// 固定窗口的key过期时间
	expiration time.Duration
}

// NewSlideWindowLimiter 初始化滑动窗口限流器，client是redis的客户端，maxCount窗口内允许的最大请求数量,
// expiration 滑动窗口的大小
func NewSlideWindowLimiter(client redis.Cmdable, maxCount int64, expiration time.Duration) *SlideWindowLimiter {
	return &SlideWindowLimiter{
		client:     client,
		maxCount:   maxCount,
		expiration: expiration,
	}
}

// Allow 是否允许请求通过限流器，key是存在redis中的键，可以标识单个接口，也可以标识一个服务
func (s SlideWindowLimiter) Allow(ctx context.Context, key string) (bool, error) {
	res, err := s.client.Eval(ctx, slideWindow, []string{key},
		s.expiration.Milliseconds(), s.maxCount, time.Now().UnixMilli()).Result()
	if err != nil {
		return false, err
	}

	if res.(string) == "true" {
		return false, errors.New("达到性能瓶颈")
	}

	return true, nil
}
