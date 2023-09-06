package single

import (
	"context"
)

// Limiter 单机使用的限流器接口
type Limiter interface {
	// Allow 是否允许通过限流器继续请求，返回true则通过，返回false和error为不通过
	Allow(ctx context.Context) (bool, error)
	// Close 关闭限流器
	Close()
}
