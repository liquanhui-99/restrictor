package single

import (
	"context"
	"errors"
	"sync/atomic"
	"time"
)

// FixedWindowLimiter 固定窗口算法限流器
type FixedWindowLimiter struct {
	// 窗口的起始时间
	timeStamp int64
	// interval 窗口的大小
	interval int64
	// 在这个窗口内允许通过的最大请求数量
	maxCount int64
	// 当前已经通过的请求数量
	currentCount int64
}

// NewFixedWindowLimiter 初始化固定窗口限流器，interval标识窗口的大小，
// maxCount窗口内允许的最大请求数量
func NewFixedWindowLimiter(interval time.Duration, maxCount int64) *FixedWindowLimiter {
	return &FixedWindowLimiter{
		timeStamp: time.Now().UnixNano(),
		interval:  int64(interval),
		maxCount:  maxCount,
	}
}

// Allow 是否允许通过限流器继续请求
func (f *FixedWindowLimiter) Allow(ctx context.Context) (bool, error) {
	now := time.Now().UnixNano()
	tm := atomic.LoadInt64(&f.timeStamp)
	cc := atomic.LoadInt64(&f.currentCount)
	// 窗口时间超过了限制，需要新开一个窗口
	if tm+f.interval < now {
		if atomic.CompareAndSwapInt64(&f.timeStamp, tm, 0) {
			atomic.CompareAndSwapInt64(&f.currentCount, cc, 0)
		}
	}
	// 窗口内的请求数量已经超过最大限度
	cc = atomic.LoadInt64(&f.currentCount)
	if f.currentCount > f.maxCount {
		return false, errors.New("超过最大请求数量限制")
	}

	atomic.AddInt64(&f.currentCount, 1)

	return true, nil
}

func (f *FixedWindowLimiter) Close() {
	//TODO implement me
	panic("implement me")
}
