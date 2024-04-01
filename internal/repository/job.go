package repository

import (
	"context"
	"time"

	"github.com/mrhelloboy/wehook/internal/repository/dao"

	"github.com/mrhelloboy/wehook/internal/domain"
)

type JobRepository interface {
	Preempt(ctx context.Context) (domain.Job, error)
	Release(ctx context.Context, id int64) error
	UpdateUtime(ctx context.Context, id int64) error
	UpdateNextTime(ctx context.Context, id int64, next time.Time) error
	Stop(ctx context.Context, id int64) error
}

type PreemptCronJobRepo struct {
	dao dao.JobDAO
}

func (p *PreemptCronJobRepo) Preempt(ctx context.Context) (domain.Job, error) {
	j, err := p.dao.Preempt(ctx)
	if err != nil {
		return domain.Job{}, err
	}
	return domain.Job{
		Id:       j.Id,
		Name:     j.Name,
		Executor: j.Executor,
		Cfg:      j.Cfg,
	}, nil
}

func (p *PreemptCronJobRepo) Release(ctx context.Context, id int64) error {
	return p.dao.Release(ctx, id)
}

func (p *PreemptCronJobRepo) UpdateUtime(ctx context.Context, id int64) error {
	return p.dao.UpdateUtime(ctx, id)
}

func (p *PreemptCronJobRepo) UpdateNextTime(ctx context.Context, id int64, next time.Time) error {
	return p.dao.UpdateNextTime(ctx, id, next)
}

func (p *PreemptCronJobRepo) Stop(ctx context.Context, id int64) error {
	return p.dao.Stop(ctx, id)
}
