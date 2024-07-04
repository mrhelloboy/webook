package trace

import (
	"context"

	"google.golang.org/grpc/metadata"

	"go.opentelemetry.io/otel/codes"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/mrhelloboy/wehook/pkg/grpcx/interceptors"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
)

type InterceptorBuilder struct {
	tracer     trace.Tracer
	propagator propagation.TextMapPropagator
	interceptors.Builder
}

func NewInterceptorBuilder(tracer trace.Tracer, propagator propagation.TextMapPropagator) *InterceptorBuilder {
	return &InterceptorBuilder{tracer: tracer, propagator: propagator}
}

func (b *InterceptorBuilder) BuildClient() grpc.UnaryClientInterceptor {
	propagator := b.propagator
	if propagator == nil {
		propagator = otel.GetTextMapPropagator()
	}
	tracer := b.tracer
	if tracer == nil {
		tracer = otel.Tracer("github.com/mrhelloboy/webook/pkg/grpcx/interceptors/trace")
	}
	attrs := []attribute.KeyValue{
		semconv.RPCSystemKey.String("grpc"),
		attribute.Key("rpc.grpc.kind").String("unary"),
		attribute.Key("rpc.component").String("client"),
	}
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) (err error) {
		ctx, span := tracer.Start(
			ctx,
			method,
			trace.WithAttributes(attrs...),
			trace.WithSpanKind(trace.SpanKindClient),
		)
		defer span.End()
		defer func() {
			if err != nil {
				span.RecordError(err)
				if e := errors.FromError(err); e != nil {
					span.SetAttributes(semconv.RPCGRPCStatusCodeKey.Int64(int64(e.Code)))
				}
				span.SetStatus(codes.Error, err.Error())
			} else {
				span.SetStatus(codes.Ok, "OK")
			}
			span.End()
		}()
		// inject 过程
		// 要 trace 有关的链路元数据，传递到服务端
		ctx = inject(ctx, propagator)
		err = invoker(ctx, method, req, reply, cc, opts...)
		return
	}
}

func (b *InterceptorBuilder) BuildServer() grpc.UnaryServerInterceptor {
	propagator := b.propagator
	if propagator == nil {
		// 全局
		propagator = otel.GetTextMapPropagator()
	}
	tracer := b.tracer
	if tracer == nil {
		tracer = otel.Tracer("github.com/mrhelloboy/webook/pkg/grpcx/interceptors/trace")
	}
	attrs := []attribute.KeyValue{
		semconv.RPCSystemKey.String("grpc"),
		attribute.Key("rpc.grpc.kind").String("unary"),
		attribute.Key("rpc.component").String("server"),
	}
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		ctx = extract(ctx, propagator)
		ctx, span := tracer.Start(ctx, info.FullMethod, trace.WithSpanKind(trace.SpanKindServer), trace.WithAttributes(attrs...))
		defer span.End()
		span.SetAttributes(semconv.RPCMethodKey.String(info.FullMethod), semconv.NetPeerNameKey.String(b.PeerName(ctx)), attribute.Key("net.peer.ip").String(b.PeerIP(ctx)))
		defer func() {
			// 就要结束了
			if err != nil {
				span.RecordError(err)
			} else {
				span.SetStatus(codes.Ok, "OK")
			}
		}()
		resp, err = handler(ctx, req)
		return
	}
}

func inject(ctx context.Context, propagators propagation.TextMapPropagator) context.Context {
	md, ok := metadata.FromOutgoingContext(ctx)
	if !ok {
		md = metadata.New(map[string]string{})
	}
	propagators.Inject(ctx, GrpcHeaderCarrier(md))
	return metadata.NewOutgoingContext(ctx, md)
}

func extract(ctx context.Context, p propagation.TextMapPropagator) context.Context {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		md = metadata.New(map[string]string{})
	}
	return p.Extract(ctx, GrpcHeaderCarrier(md))
}

type GrpcHeaderCarrier metadata.MD

func (g GrpcHeaderCarrier) Get(key string) string {
	vals := metadata.MD(g).Get(key)
	if len(vals) > 0 {
		return vals[0]
	}
	return ""
}

func (g GrpcHeaderCarrier) Set(key string, value string) {
	metadata.MD(g).Set(key, value)
}

func (g GrpcHeaderCarrier) Keys() []string {
	keys := make([]string, 0, len(g))
	for k := range metadata.MD(g) {
		keys = append(keys, k)
	}
	return keys
}
