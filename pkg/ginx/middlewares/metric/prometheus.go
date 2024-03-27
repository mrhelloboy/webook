package metric

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
)

type Builder struct {
	Namespace  string
	Subsystem  string
	Name       string
	Help       string
	InstanceID string
}

func NewBuilder(namespace, subsystem, name, help, instanceID string) *Builder {
	return &Builder{
		Namespace:  namespace,
		Subsystem:  subsystem,
		Name:       name,
		Help:       help,
		InstanceID: instanceID,
	}
}

func (b *Builder) Build() gin.HandlerFunc {
	// pattern 是指命中的路由
	labels := []string{"method", "pattern", "status"}
	summary := prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Namespace: b.Namespace,
		Subsystem: b.Subsystem,
		Name:      b.Name + "_resp_time",
		Help:      b.Help,
		ConstLabels: map[string]string{
			"instance_id": b.InstanceID,
		},
		Objectives: map[float64]float64{
			0.5:   0.05,
			0.9:   0.01,
			0.99:  0.005,
			0.999: 0.0001,
		},
	}, labels)
	prometheus.MustRegister(summary)
	gauge := prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: b.Namespace,
		Subsystem: b.Subsystem,
		Name:      b.Name + "_active_req",
		Help:      b.Help,
		ConstLabels: map[string]string{
			"instance_id": b.InstanceID,
		},
	})
	prometheus.MustRegister(gauge)
	return func(ctx *gin.Context) {
		start := time.Now()
		gauge.Inc()
		defer func() {
			duration := time.Since(start)
			gauge.Dec()
			pattern := ctx.FullPath()
			if pattern == "" {
				pattern = "unknown"
			}
			summary.WithLabelValues(
				ctx.Request.Method,
				pattern,
				strconv.Itoa(ctx.Writer.Status()),
			).Observe(float64(duration.Milliseconds()))
		}()
		ctx.Next()
	}
}
