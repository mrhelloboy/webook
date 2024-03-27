package metrics

import (
	"context"
	"time"

	"github.com/mrhelloboy/wehook/internal/service/sms"
	"github.com/prometheus/client_golang/prometheus"
)

// 给 SMS 服务添加 metrics 功能的装饰器

type PrometheusDecorator struct {
	svc    sms.Service
	vector *prometheus.SummaryVec
}

func (p *PrometheusDecorator) Send(ctx context.Context, biz string, args []string, numbers ...string) error {
	startTime := time.Now()
	defer func() {
		duration := time.Since(startTime)
		p.vector.WithLabelValues(biz).Observe(float64(duration))
	}()
	return p.svc.Send(ctx, biz, args, numbers...)
}

func NewPrometheusDecorator(svc sms.Service) *PrometheusDecorator {
	vector := prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Namespace: "geekbang_daming",
		Subsystem: "webook",
		Name:      "sms_resp_time",
		Help:      "统计 SMS 服务的性能数据",
	}, []string{"biz"})
	prometheus.MustRegister(vector)
	return &PrometheusDecorator{
		svc:    svc,
		vector: vector,
	}
}
