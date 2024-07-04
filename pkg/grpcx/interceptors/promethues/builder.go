package promethues

import (
	"context"
	"strings"
	"time"

	"github.com/mrhelloboy/wehook/pkg/grpcx/interceptors"
	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

type InterceptorBuilder struct {
	Namespace string
	Subsystem string
	interceptors.Builder
}

func (b *InterceptorBuilder) BuildServer() grpc.UnaryServerInterceptor {
	summary := prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: b.Namespace,
			Subsystem: b.Subsystem,
			Name:      "server_handle_seconds",
			Objectives: map[float64]float64{
				0.5:   0.01,
				0.9:   0.01,
				0.95:  0.01,
				0.99:  0.001,
				0.999: 0.0001,
			},
		}, []string{"type", "service", "method", "peer", "code"})
	prometheus.MustRegister(summary)
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		start := time.Now()
		defer func() {
			s, m := b.splitMethodName(info.FullMethod)
			duration := float64(time.Since(start).Milliseconds())
			if err == nil {
				summary.WithLabelValues("unary", s, m, b.PeerName(ctx), "OK").Observe(duration)
			} else {
				st, _ := status.FromError(err)
				summary.WithLabelValues("unary", s, m, b.PeerName(ctx), st.Code().String()).Observe(duration)
			}
		}()
		resp, err = handler(ctx, req)
		return
	}
}

func (b *InterceptorBuilder) splitMethodName(fullMethdName string) (string, string) {
	fullMethdName = strings.TrimPrefix(fullMethdName, "/")
	if i := strings.Index(fullMethdName, "/"); i >= 0 {
		return fullMethdName[:i], fullMethdName[i+1:]
	}
	return "unknown", "unknown"
}
