package distribute

import "context"

// DistributedLimiter 分布式场景下使用的限流器接口
type DistributedLimiter interface {
	// Allow 是否允许通过限流器继续请求
	// key 是存储在Redis中的键，可以是单个接口，也可以是整个服务
	// 返回true则通过，返回false和error为不通过
	Allow(ctx context.Context, key string) (bool, error)
}
