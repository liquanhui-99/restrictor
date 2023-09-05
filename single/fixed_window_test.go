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

func TestFixedWindowLimiter_Allow(t *testing.T) {
	testCases := []struct {
		name     string
		interval time.Duration
		before   func(*testing.T, *FixedWindowLimiter)
		ctx      func() (context.Context, context.CancelFunc)
		wantErr  error
		wantRes  bool
		maxCount int64
	}{
		// 开新的窗口
		{
			name:     "reset",
			interval: 10 * time.Millisecond,
			before: func(t *testing.T, limiter *FixedWindowLimiter) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()
				for i := 0; i < 10; i++ {
					time.Sleep(10 * time.Millisecond)
					res, err := limiter.Allow(ctx)
					require.NoError(t, err)
					require.True(t, res)
				}
			},
			maxCount: 10,
			wantErr:  nil,
			wantRes:  true,
		},
		// 超过最大数量
		{
			name:     "over max count",
			interval: 2 * time.Minute,
			before: func(t *testing.T, limiter *FixedWindowLimiter) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				defer cancel()
				for i := 0; i < 10; i++ {
					res, err := limiter.Allow(ctx)
					require.NoError(t, err)
					require.True(t, res)
				}
			},
			maxCount: 10,
			wantErr:  errors.New("超过最大请求数量限制"),
			wantRes:  false,
		},
		// 超过最大数量
		{
			name:     "over max count",
			interval: 2 * time.Minute,
			before:   func(t *testing.T, limiter *FixedWindowLimiter) {},
			maxCount: 10,
			wantErr:  nil,
			wantRes:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			limiter := NewFixedWindowLimiter(tc.interval, tc.maxCount)
			tc.before(t, limiter)
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			res, err := limiter.Allow(ctx)
			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.wantRes, res)
		})
	}
}

func ExampleFixedWindowLimiter_Allow() {
	r := gin.Default()
	var limit = NewFixedWindowLimiter(10*time.Second, 10)
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
