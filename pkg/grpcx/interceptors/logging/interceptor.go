package logging

import (
	"context"
	"fmt"
	"runtime"
	"time"

	"github.com/mrhelloboy/wehook/pkg/grpcx/interceptors"
	"github.com/mrhelloboy/wehook/pkg/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type InterceptorBuilder struct {
	l logger.Logger
	// fn func(msg string, fields ...logger.Field)
	interceptors.Builder
	reqBody  bool
	respBody bool
}

func (i *InterceptorBuilder) BuildClient() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

func (i *InterceptorBuilder) Build() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		start := time.Now()
		event := "normal"
		defer func() {
			// 执行时间
			duration := time.Since(start)
			if rec := recover(); rec != nil {
				switch recType := rec.(type) {
				case error:
					err = recType
				default:
					err = fmt.Errorf("%v", rec)
				}
				stack := make([]byte, 4096)
				stack = stack[:runtime.Stack(stack, true)]
				event = "recover"
				err = status.New(codes.Internal, "panic, err"+err.Error()).Err()
			}
			fields := []logger.Field{
				logger.Int64("cost", duration.Milliseconds()),
				logger.String("type", "unary"),
				logger.String("method", info.FullMethod),
				logger.String("event", event),
				// 需要客户端配合
				// 需要知道是哪一个业务调用过来的
				// 是哪个业务的哪个节点过来的
				logger.String("peer", i.PeerName(ctx)),
				logger.String("peer_ip", i.PeerIP(ctx)),
			}
			if err != nil {
				st, _ := status.FromError(err)
				fields = append(fields, logger.String("code", st.Code().String()), logger.String("code_msg", st.Message()))
			}
			i.l.Info("RPC请求", fields...)
		}()
		resp, err = handler(ctx, req)
		return
	}
}
