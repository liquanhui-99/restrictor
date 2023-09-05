package single

import (
	"context"
	"sync"
	"time"
)

// LeakeyBucketLimiter 漏桶算法实现的限流器
type LeakeyBucketLimiter struct {
	// ticker控制请求的通过频率
	t *time.Ticker
	// closeCh 控制关闭
	close chan struct{}
	// once 控制关闭一次
	once *sync.Once
}

// NewLeakeyBucketLimiter 初始化漏桶限流器, interval流量限流的间隔，即多久可以通过一次请求
func NewLeakeyBucketLimiter(interval time.Duration) *LeakeyBucketLimiter {
	t := time.NewTicker(interval)
	return &LeakeyBucketLimiter{
		t:     t,
		once:  &sync.Once{},
		close: make(chan struct{}),
	}
}

// Allow 是否允许通过限流器继续请求
func (l LeakeyBucketLimiter) Allow(ctx context.Context) (bool, error) {
	select {
	case <-ctx.Done():
		return false, ctx.Err()
	case <-l.t.C:
		return true, nil
	case <-l.close:
		return true, nil
	}
}

// Close 关闭限流器
func (l LeakeyBucketLimiter) Close() {
	l.once.Do(func() {
		l.t.Stop()
		close(l.close)
	})
}
