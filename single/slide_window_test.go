package single

import (
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/assert/v2"
	"github.com/stretchr/testify/require"
	"net/http"
	"testing"
	"time"
)

func TestSideWindow_Allow(t *testing.T) {
	testCases := []struct {
		name     string
		interval time.Duration
		maxCount int64
		before   func(*testing.T, *SlideWindowLimiter)
		wantErr  error
		wantRes  bool
	}{
		// 测试快路径
		{
			name:     "fast path",
			interval: time.Second,
			maxCount: 10,
			before:   func(t *testing.T, limiter *SlideWindowLimiter) {},
			wantErr:  nil,
			wantRes:  true,
		},

		// 测试删除队头不在时间窗口内的数据
		{
			name:     "delete front",
			interval: time.Second,
			maxCount: 10,
			before: func(t *testing.T, limiter *SlideWindowLimiter) {
				ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Second))
				defer cancel()
				for i := 0; i < 9; i++ {
					res, err := limiter.Allow(ctx)
					require.NoError(t, err)
					require.True(t, res)
				}
			},
			wantErr: nil,
			wantRes: true,
		},
		// 慢路径
		{
			name:     "slow front",
			interval: 100 * time.Millisecond,
			maxCount: 10,
			before: func(t *testing.T, limiter *SlideWindowLimiter) {
				ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Second))
				defer cancel()
				for i := 0; i < 10; i++ {
					time.Sleep(100 * time.Millisecond)
					res, err := limiter.Allow(ctx)
					require.NoError(t, err)
					require.True(t, res)
				}
			},
			wantErr: nil,
			wantRes: true,
		},
		// 超过限制
		{
			name:     "slow front",
			interval: 100 * time.Millisecond,
			maxCount: 10,
			before: func(t *testing.T, limiter *SlideWindowLimiter) {
				ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Second))
				defer cancel()
				for i := 0; i < 10; i++ {
					res, err := limiter.Allow(ctx)
					require.NoError(t, err)
					require.True(t, res)
				}
			},
			wantErr: errors.New("达到了性能瓶颈"),
			wantRes: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			limiter := NewSlideWindowLimiter(tc.interval, tc.maxCount)
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			tc.before(t, limiter)
			res, err := limiter.Allow(ctx)
			assert.Equal(t, err, tc.wantErr)
			assert.Equal(t, res, tc.wantRes)
		})
	}
}

func ExampleSlideWindowLimiter_Allow() {
	r := gin.Default()
	var limit = NewSlideWindowLimiter(10*time.Second, 10)
	r.Use(func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		ok, err := limit.Allow(ctx)
		if err != nil {
			c.AbortWithStatus(http.StatusTooManyRequests)
			return
		}
		if !ok {
			c.AbortWithStatus(http.StatusTooManyRequests)
			return
		}
		c.Next()
	})
	r.GET("/profile", func(c *gin.Context) {
		c.Writer.WriteHeader(http.StatusOK)
		_, _ = c.Writer.Write([]byte("请求成功"))
	})
	if err := r.Run(":8083"); err != nil {
		panic(err)
	}
	//output
}
