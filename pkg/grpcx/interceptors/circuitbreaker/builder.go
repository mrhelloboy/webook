package circuitbreaker

import (
	"context"
	"math/rand"

	"github.com/go-kratos/aegis/circuitbreaker"
	"google.golang.org/grpc"
)

type InterceptorBuilder struct {
	breaker circuitbreaker.CircuitBreaker
	// 设置标记位
	// 假如我们考虑使用随机数 + 阈值的回复方式
	// 触发熔断时，直接将 threshold 设置为 0
	// 后续等待一段时间，将 threshold 调整为 1，判定请求有没有问题
	threshold int
}

func (b *InterceptorBuilder) BuildServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		if b.breaker.Allow() == nil {
			resp, err = handler(ctx, req)
			if err != nil {
				// 可以再进一步区别是不是系统错误
				// 这里没有区别业务错误和系统错误
				b.breaker.MarkFailed()
			} else {
				b.breaker.MarkSuccess()
			}
		}
		b.breaker.MarkFailed()
		// 触发了熔断
		return nil, err
	}
}

func (b *InterceptorBuilder) BuildServerInterceptorV1() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		if !b.allow() {
			b.threshold = b.threshold / 2
			// 这里就是触发了熔断
			//b.threshold = 0
			//time.AfterFunc(time.Minute, func() {
			//	b.threshold = 1
			//})
		}
		// 下面随机数判定
		r := rand.Intn(100)
		if r <= b.threshold {
			resp, err = handler(ctx, req)
			if err == nil && b.threshold != 0 {
				// 你要考虑调大 threshold
			} else if b.threshold != 0 {
				// 你要考虑调小 threshold
			}
		}
		return
	}
}

func (b *InterceptorBuilder) allow() bool {
	// 判定节点是否健康的各种做法
	// 从 prometheus 里面拿数据判定
	// prometheus.DefaultGatherer.Gather()
	return false
}
