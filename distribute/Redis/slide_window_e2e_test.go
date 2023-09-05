package Redis

import (
	"context"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestSlideWindowLimiter_Allow(t *testing.T) {
	client := redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6379",
		Password: "123456",
	})

	limit := NewSlideWindowLimiter(client, 100, time.Minute)

	for i := 0; i < 110; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		res, err := limit.Allow(ctx, "test")
		cancel()
		require.NoError(t, err)
		require.True(t, res)
	}
}
