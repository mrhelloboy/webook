package ioc

import (
	"time"

	"github.com/robfig/cron/v3"

	rlock "github.com/gotomicro/redis-lock"
	"github.com/mrhelloboy/wehook/internal/job"
	"github.com/mrhelloboy/wehook/internal/service"
	"github.com/mrhelloboy/wehook/pkg/logger"
)

func InitRankingJob(svc service.RankingService, rlockClient *rlock.Client, l logger.Logger) *job.RankingJob {
	return job.NewRankingJob(svc, time.Second*30, rlockClient, l)
}

func InitJobs(l logger.Logger, rankingJob *job.RankingJob) *cron.Cron {
	res := cron.New(cron.WithSeconds())
	cbd := job.NewCronJobBuilder(l)
	// 这里每三分钟一次
	_, err := res.AddJob("0 */3 * * * ?", cbd.Build(rankingJob))
	if err != nil {
		panic(err)
	}
	return res
}
