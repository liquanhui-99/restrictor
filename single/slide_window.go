package single

import (
	"container/list"
	"context"
	"errors"
	"sync"
	"time"
)

// SlideWindowLimiter 滑动窗口算法实现的限流器
type SlideWindowLimiter struct {
	// 窗口的大小
	interval int64
	// 缓存之前每一个请求的时间戳
	queue *list.List
	// 窗口内允许的最大请求树
	maxCount int64
	// 加锁保护queue
	mu sync.Mutex
}

// NewSlideWindowLimiter 初始化滑动窗口限流器，interval标识窗口的大小
// maxCount窗口内的最大请求数量
func NewSlideWindowLimiter(interval time.Duration, maxCount int64) *SlideWindowLimiter {
	return &SlideWindowLimiter{
		interval: int64(interval),
		queue:    list.New(),
		maxCount: maxCount,
	}
}

func (s *SlideWindowLimiter) Allow(ctx context.Context) (bool, error) {
	now := time.Now().UnixNano()
	// 快路径，只要队列的长度小于最大的限流数就可以直接通过
	s.mu.Lock()
	if int64(s.queue.Len()) < s.maxCount {
		_ = s.queue.PushBack(now)
		s.mu.Unlock()
		return true, nil
	}

	// 慢路径，队列满了，必须先清理出超过窗口时间的请求，再取判断是否超过
	// 最大请求限制
	boundary := now - s.interval
	first := s.queue.Front()
	// 循环队列，第一个元素不为空且第一个元素的请求时间不在窗口范围之内，
	// 需要清除掉缓存
	for first != nil && boundary > first.Value.(int64) {
		_ = s.queue.Remove(first)
		first = s.queue.Front()
	}

	if int64(s.queue.Len()) >= s.maxCount {
		s.mu.Unlock()
		return false, errors.New("达到了性能瓶颈")
	}

	_ = s.queue.PushBack(now)
	s.mu.Unlock()

	return true, nil
}

func (s *SlideWindowLimiter) Close() {}
