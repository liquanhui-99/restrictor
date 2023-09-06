package single

import (
	"context"
	"errors"
	"sync"
	"time"
)

// TokenBucketLimiter 令牌桶算法实现的限流器
type TokenBucketLimiter struct {
	// 令牌桶队列
	ch chan struct{}
	// 关闭令牌功能
	close chan struct{}
	// once控制只能关闭关闭一次
	once *sync.Once
}

// NewTokenBucketLimiter 初始化令牌桶，capacity是缓存令牌的channel容量，控制可以通过的最大请求，
// 容量设置需要谨慎，如果开的过大，服务器可能会被瞬间的流量击垮；interval是发送令牌的间隔，多久发送一次令牌
func NewTokenBucketLimiter(capacity int64, interval time.Duration) *TokenBucketLimiter {
	ch := make(chan struct{}, capacity)
	closeCh := make(chan struct{})
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-closeCh:
				return
			case <-ticker.C:
				// 发送令牌
				select {
				case ch <- struct{}{}:
				default:
				}
			}
		}
	}()

	limiter := &TokenBucketLimiter{
		ch:    ch,
		close: closeCh,
		once:  &sync.Once{},
	}

	return limiter
}

// Allow 是否运行继续请求
func (t TokenBucketLimiter) Allow(ctx context.Context) (bool, error) {
	select {
	case <-t.close:
		// 关闭限流器
		return true, nil
	case <-ctx.Done():
		return false, ctx.Err()
	case <-t.ch:
		return true, nil
	default:
		return false, errors.New("达到了性能瓶颈")
	}
}

// Close 关闭限流器
func (t TokenBucketLimiter) Close() {
	t.once.Do(func() {
		close(t.close)
		close(t.ch)
	})
}
