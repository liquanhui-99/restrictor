package expand

import (
	"context"
	"errors"
	"github.com/liquanhui-99/restrictor/single"
	"sync"
	"time"
)

// IpLimiter 基于限流器基础上实现的ip限流器
type IpLimiter struct {
	// 本地缓存ip数据，key是ip地址，val是请求的数量
	ips map[string]int64
	// 组合单体的限流接口
	limiter single.Limiter
	// 加读写锁控制本地缓存
	mu sync.RWMutex
	// 限流ip的间隔，多久重置一次ip限流
	interval time.Duration
	// 关闭限流器
	close chan struct{}
	// 单个ip单位时间内最大的请求数
	maxCount int64
}

// NewIpLimiter 初始化Ip限流器
// limiter是单体限流器的实现
// interval是重置ip缓存的间隔，也是ip计数的周期，例如：1分钟内单个ip只允许有100次请求，那间隔就是time.Minute
// maxCount 间隔内单个ip的最大请求数限制
func NewIpLimiter(limiter single.Limiter, interval time.Duration, maxCount int64) *IpLimiter {
	closeCh := make(chan struct{})
	res := &IpLimiter{
		ips:      map[string]int64{},
		limiter:  limiter,
		close:    closeCh,
		mu:       sync.RWMutex{},
		maxCount: maxCount,
	}
	go func() {
		ticker := time.NewTicker(interval)
		for {
			select {
			case <-closeCh:
				// 退出
				return
			case <-ticker.C:
				// 过了单位时间，重新计算
				res.mu.Lock()
				res.ips = map[string]int64{}
				res.mu.Unlock()
			}
		}
	}()
	return res
}

func (l *IpLimiter) AllowIp(ctx context.Context, ip string) (bool, error) {
	// 快路径
	l.mu.RLock()
	cnt, ok := l.ips[ip]
	l.mu.RUnlock()
	if ok && cnt >= l.maxCount {
		return false, errors.New("单位时间ip达到最大数量限制")
	}

	// 慢路径
	res, err := l.limiter.Allow(ctx)
	if err != nil {
		return false, err
	}

	if !res {
		return false, errors.New("达到性能瓶颈")
	}

	l.mu.RLock()
	cnt, ok = l.ips[ip]
	l.mu.RUnlock()
	if ok && cnt >= l.maxCount {
		return false, errors.New("单位时间ip达到最大数量限制")
	}

	l.mu.Lock()
	defer l.mu.Unlock()
	cnt, ok = l.ips[ip]
	if !ok {
		l.ips[ip] = 1
		return true, nil
	}
	if cnt >= l.maxCount {
		return false, errors.New("单位时间ip达到最大数量限制")
	}
	l.ips[ip] = cnt + 1
	return true, nil

}

func (l *IpLimiter) Close() {
	l.limiter.Close()
	close(l.close)
}
