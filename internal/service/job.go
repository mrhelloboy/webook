package service

import (
	"context"
	"time"

	"github.com/mrhelloboy/wehook/internal/repository"
	"github.com/mrhelloboy/wehook/pkg/logger"

	"github.com/mrhelloboy/wehook/internal/domain"
)

type JobService interface {
	Preempt(ctx context.Context) (domain.Job, error)
	ResetNextTime(ctx context.Context, j domain.Job) error
}

type cronJobSvc struct {
	repo            repository.JobRepository
	refreshInterval time.Duration
	l               logger.Logger
}

func (c *cronJobSvc) Preempt(ctx context.Context) (domain.Job, error) {
	j, err := c.repo.Preempt(ctx)

	// 续约
	ticker := time.NewTicker(c.refreshInterval)
	go func() {
		for range ticker.C {
			c.refresh(j.Id)
		}
	}()

	// 注意：考虑释放问题
	j.CancelFunc = func() error {
		ticker.Stop()
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		return c.repo.Release(ctx, j.Id)
	}
	return j, err
}

func (c *cronJobSvc) ResetNextTime(ctx context.Context, j domain.Job) error {
	next := j.NextTime()
	if next.IsZero() {
		return c.repo.Stop(ctx, j.Id)
	}
	return c.repo.UpdateNextTime(ctx, j.Id, next)
}

func (c *cronJobSvc) refresh(id int64) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	// 续约方式：
	// 更新一下 utime 即可
	err := c.repo.UpdateUtime(ctx, id)
	if err != nil {
		// todo: 可以考虑重试
		c.l.Error("续约失败", logger.Error(err), logger.Int64("job_id", id))
	}
}
