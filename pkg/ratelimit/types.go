package ratelimit

import "context"

type Limiter interface {
	// Limit 是否触发限流
	// key 限流的对象
	// 返回值为 true 表示触发限流，false 表示未触发限流
	// 返回值为 error 表示限流器本身有错误
	Limit(ctx context.Context, key string) (bool, error)
}
