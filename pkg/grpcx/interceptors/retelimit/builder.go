package retelimit

import (
	"context"
	"strings"

	"github.com/mrhelloboy/wehook/pkg/logger"
	"github.com/mrhelloboy/wehook/pkg/ratelimit"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type InterceptorBuilder struct {
	limiter ratelimit.Limiter
	key     string
	l       logger.Logger
}

func NewInterceptorBuilder(limiter ratelimit.Limiter, key string, l logger.Logger) *InterceptorBuilder {
	return &InterceptorBuilder{limiter: limiter, key: key, l: l}
}

func (b *InterceptorBuilder) BuildServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		limited, err := b.limiter.Limit(ctx, b.key)
		if err != nil {
			// 保守策略，限流失败时，直接返回错误
			b.l.Error("判定限流出现问题", logger.Error(err))
			return nil, status.Errorf(codes.ResourceExhausted, "触发限流")

			// 激进策略
			// return handler(ctx, req)
		}
		if limited {
			return nil, status.Errorf(codes.ResourceExhausted, "触发限流")
		}
		return handler(ctx, req)
	}
}

// BuildServerInterceptorV1 用来配合后面业务的
func (b *InterceptorBuilder) BuildServerInterceptorV1() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		limited, err := b.limiter.Limit(ctx, b.key)
		if err != nil || limited {
			ctx = context.WithValue(ctx, "limited", "true")
		}
		return handler(ctx, req)
	}
}

func (b *InterceptorBuilder) BuildClientInterceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		limited, err := b.limiter.Limit(ctx, b.key)
		if err != nil {
			b.l.Error("判定限流出现问题", logger.Error(err))
			return status.Errorf(codes.ResourceExhausted, "触发限流")
		}
		if limited {
			return status.Errorf(codes.ResourceExhausted, "触发限流")
		}
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

func (b *InterceptorBuilder) BuildServerInterceptorService() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		if strings.HasPrefix(info.FullMethod, "/UserService") {
			limited, err := b.limiter.Limit(ctx, "limiter:service:user:UserService")
			if err != nil {
				b.l.Error("判定限流出现问题", logger.Error(err))
				return nil, status.Errorf(codes.ResourceExhausted, "触发限流")
			}
			if limited {
				return nil, status.Errorf(codes.ResourceExhausted, "触发限流")
			}
		}
		return handler(ctx, req)
	}
}
