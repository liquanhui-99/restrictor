package single

import (
	"context"
	"errors"
	"github.com/go-playground/assert/v2"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestTokenBucketLimiter_Allow(t *testing.T) {
	testCases := []struct {
		name     string
		capacity int64
		interval time.Duration
		wantErr  error
		wantRes  bool
		before   func()
		after    func(*TokenBucketLimiter)
		ctx      func() (context.Context, context.CancelFunc)
	}{
		// context超时
		{
			name:     "Deadline",
			capacity: 1,
			interval: time.Second,
			wantErr:  context.DeadlineExceeded,
			wantRes:  false,
			before:   func() {},
			after: func(limiter *TokenBucketLimiter) {
				limiter.Close()
			},
			ctx: func() (context.Context, context.CancelFunc) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Nanosecond)
				return ctx, cancel
			},
		},
		// 通过限流器
		{
			name:     "success",
			capacity: 100,
			interval: 10 * time.Millisecond,
			wantErr:  nil,
			wantRes:  true,
			after: func(limiter *TokenBucketLimiter) {
				limiter.Close()
			},
			ctx: func() (context.Context, context.CancelFunc) {
				time.Sleep(time.Second)
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				return ctx, cancel
			},
		},
		// 达到性能瓶颈
		{
			name:     "max count",
			capacity: 0,
			interval: 1 * time.Second,
			wantErr:  errors.New("达到了性能瓶颈"),
			wantRes:  false,
			after: func(limiter *TokenBucketLimiter) {
				limiter.Close()
			},
			ctx: func() (context.Context, context.CancelFunc) {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				return ctx, cancel
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			limiter := NewTokenBucketLimiter(tc.capacity, tc.interval)
			c, cancel := tc.ctx()
			defer cancel()
			ok, err := limiter.Allow(c)
			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.wantRes, ok)
			tc.after(limiter)
		})
	}
}

func TestTokenBucketLimiter_Close(t *testing.T) {
	limiter := NewTokenBucketLimiter(10, 1*time.Millisecond)
	time.Sleep(10 * time.Millisecond)
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(time.Second))
	defer cancel()
	res, err := limiter.Allow(ctx)
	require.NoError(t, err)
	require.True(t, res)
	limiter.Close()
	limiter.Close()
}

func TestTokenBucketLimiter_channelBlock(t *testing.T) {
	limiter := NewTokenBucketLimiter(5, 1*time.Millisecond)
	time.Sleep(10 * time.Millisecond)
	limiter.Close()
	limiter.Close()
}
