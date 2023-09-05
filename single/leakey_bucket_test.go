package single

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/assert/v2"
	"github.com/stretchr/testify/require"
	"net/http"
	"testing"
	"time"
)

func TestLeakeyBucketLimiter_Allow(t *testing.T) {
	testCases := []struct {
		name     string
		interval time.Duration
		after    func(*LeakeyBucketLimiter)
		ctx      func() (context.Context, context.CancelFunc)
		wantErr  error
		wantRes  bool
	}{
		// 超时
		{
			name:     "Deadline",
			interval: 5 * time.Millisecond,
			after: func(limiter *LeakeyBucketLimiter) {
				limiter.Close()
			},
			ctx: func() (context.Context, context.CancelFunc) {
				ctx, cancel := context.WithTimeout(context.Background(), 2*time.Millisecond)
				return ctx, cancel
			},
			wantErr: context.DeadlineExceeded,
			wantRes: false,
		},
		// 成功
		{
			name:     "success",
			interval: 2 * time.Millisecond,
			after: func(limiter *LeakeyBucketLimiter) {
				limiter.Close()
			},
			ctx: func() (context.Context, context.CancelFunc) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				return ctx, cancel
			},
			wantErr: nil,
			wantRes: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			limiter := NewLeakeyBucketLimiter(tc.interval)
			c, cancel := tc.ctx()
			defer cancel()
			res, err := limiter.Allow(c)
			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.wantRes, res)
			tc.after(limiter)
		})
	}
}

func TestLeakeyBucketLimiter_Close(t *testing.T) {
	limiter := NewLeakeyBucketLimiter(2 * time.Millisecond)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	res, err := limiter.Allow(ctx)
	require.NoError(t, err)
	require.True(t, res)
	limiter.Close()
	limiter.Close()
}

func ExampleLeakeyBucketLimiter_Allow() {
	r := gin.Default()
	var limit = NewLeakeyBucketLimiter(10 * time.Second)
	defer limit.Close()
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
