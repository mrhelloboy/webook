package ratelimit

import (
	"context"
	_ "embed"
	"github.com/redis/go-redis/v9"
	"time"
)

//go:embed slide_window.lua
var luaSlideWindow string

// RedisSliceWindowLimiter Redis 上的滑动窗口限流器
type RedisSliceWindowLimiter struct {
	cmd      redis.Cmdable
	interval time.Duration // 窗口大小
	rate     int           // 阈值
	// interval 内允许 rate 个请求 eg: 1s 内允许 3000 个请求
}

func NewRedisSliceWindowLimiter(cmd redis.Cmdable, interval time.Duration, rate int) *RedisSliceWindowLimiter {
	return &RedisSliceWindowLimiter{
		cmd:      cmd,
		interval: interval,
		rate:     rate,
	}
}

func (r *RedisSliceWindowLimiter) Limit(ctx context.Context, key string) (bool, error) {
	return r.cmd.Eval(ctx, luaSlideWindow, []string{key},
		r.interval.Milliseconds(), r.rate, time.Now().UnixMilli()).Bool()
}
