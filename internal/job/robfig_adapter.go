package job

import (
	"time"

	"github.com/mrhelloboy/wehook/pkg/logger"
	"github.com/prometheus/client_golang/prometheus"
)

type RankingJobAdapter struct {
	j Job
	l logger.Logger
	p prometheus.Summary
}

func NewRankingJobAdapter(j Job, l logger.Logger) *RankingJobAdapter {
	p := prometheus.NewSummary(prometheus.SummaryOpts{
		Name:        "cron_job",
		ConstLabels: map[string]string{"name": j.Name()},
	})
	prometheus.MustRegister(p)
	return &RankingJobAdapter{
		j: j,
		l: l,
		p: p,
	}
}

func (a *RankingJobAdapter) Run() {
	start := time.Now()
	defer func() {
		duration := time.Since(start).Milliseconds()
		a.p.Observe(float64(duration))
	}()
	err := a.j.Run()
	if err != nil {
		a.l.Error("运行任务失败", logger.Error(err), logger.String("job", a.j.Name()))
	}
}
