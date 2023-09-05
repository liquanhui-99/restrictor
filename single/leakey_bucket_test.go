package single

import (
	"context"
	"github.com/go-playground/assert/v2"
	"github.com/stretchr/testify/require"
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
