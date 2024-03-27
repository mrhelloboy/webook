package redisx

import (
	"context"
	"errors"
	"net"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/redis/go-redis/v9"
)

type PrometheusHook struct {
	vector *prometheus.SummaryVec
}

func (p *PrometheusHook) DialHook(next redis.DialHook) redis.DialHook {
	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		// 什么也不干
		return next(ctx, network, addr)
	}
}

func (p *PrometheusHook) ProcessHook(next redis.ProcessHook) redis.ProcessHook {
	return func(ctx context.Context, cmd redis.Cmder) error {
		// 在 Redis 执行之前
		startTime := time.Now()
		var err error
		defer func() {
			duration := time.Since(startTime).Milliseconds()
			keyExist := errors.Is(redis.Nil, err)
			p.vector.WithLabelValues(cmd.Name(), strconv.FormatBool(keyExist)).Observe(float64(duration))
		}()
		// 这个会最终发送命令到 redis 上
		err = next(ctx, cmd)
		return err
	}
}

func (p *PrometheusHook) ProcessPipelineHook(next redis.ProcessPipelineHook) redis.ProcessPipelineHook {
	// TODO implement me
	panic("implement me")
}

func NewPrometheusHook(opt prometheus.SummaryOpts) *PrometheusHook {
	vector := prometheus.NewSummaryVec(opt, []string{"cmd", "key_exist"}) // key_exist 是否命中缓存
	prometheus.MustRegister(vector)
	return &PrometheusHook{
		vector: vector,
	}
}
