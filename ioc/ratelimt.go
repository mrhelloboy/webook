package ioc

import (
	"github.com/mrhelloboy/wehook/pkg/ratelimit"
	"github.com/redis/go-redis/v9"
	"time"
)

// InitRateLimiterOfMiddleware 初始化中间件上的限流器
func InitRateLimiterOfMiddleware(redisClient redis.Cmdable) ratelimit.Limiter {
	return ratelimit.NewRedisSliceWindowLimiter(redisClient, time.Second, 100)
}
