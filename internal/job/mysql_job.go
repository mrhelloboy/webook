package job

import (
	"context"
	"fmt"
	"time"

	"github.com/mrhelloboy/wehook/internal/service"
	"github.com/mrhelloboy/wehook/pkg/logger"
	"golang.org/x/sync/semaphore"

	"github.com/mrhelloboy/wehook/internal/domain"
)

type Executor interface {
	Name() string
	// Exec ctx 是整个任务调度的上下文
	// 当从 ctx.Done 有信号的时候，就需要考虑结束执行
	Exec(ctx context.Context, j domain.Job) error
}

type LocalFuncExecutor struct {
	funcs map[string]func(ctx context.Context, j domain.Job) error
}

func (l *LocalFuncExecutor) Name() string {
	return "local"
}

func (l *LocalFuncExecutor) Exec(ctx context.Context, j domain.Job) error {
	fn, ok := l.funcs[j.Name]
	if !ok {
		return fmt.Errorf("未知任务，是否注册了？ %s", j.Name)
	}
	return fn(ctx, j)
}

func (l *LocalFuncExecutor) RegisterFunc(name string, fn func(ctx context.Context, j domain.Job) error) {
	l.funcs[name] = fn
}

func NewLocalFuncExecutor() *LocalFuncExecutor {
	return &LocalFuncExecutor{
		funcs: make(map[string]func(ctx context.Context, j domain.Job) error),
	}
}

// Scheduler 调度器
type Scheduler struct {
	execs   map[string]Executor
	svc     service.JobService
	l       logger.Logger
	limiter *semaphore.Weighted
}

func NewScheduler(svc service.JobService, l logger.Logger) *Scheduler {
	return &Scheduler{
		svc:     svc,
		l:       l,
		execs:   make(map[string]Executor),
		limiter: semaphore.NewWeighted(200),
	}
}

func (s *Scheduler) RegisterExecutor(e Executor) {
	s.execs[e.Name()] = e
}

func (s *Scheduler) Schedule(ctx context.Context) error {
	for {
		if ctx.Err() != nil {
			// 退出循环
			return ctx.Err()
		}
		err := s.limiter.Acquire(ctx, 1)
		if err != nil {
			return err
		}
		// 一次调度的数据库查询时间
		dbCtx, cancel := context.WithTimeout(ctx, time.Second)
		j, err := s.svc.Preempt(dbCtx)
		cancel()
		if err != nil {
			// 继续下一轮抢占
			s.l.Error("抢占任务失败", logger.Error(err))
		}
		exec, ok := s.execs[j.Executor]
		if !ok {
			s.l.Error("未找到对应的执行器", logger.String("executor", j.Executor))
			continue
		}
		// 执行
		go func() {
			defer func() {
				s.limiter.Release(1)
				err1 := j.CancelFunc()
				if err1 != nil {
					s.l.Error("释放任务失败", logger.Error(err1), logger.Int64("job_id", j.Id))
				}
			}()

			// 异步执行，不要阻塞主调度循环
			err1 := exec.Exec(ctx, j)
			if err1 != nil {
				// todo：可以重试
				s.l.Error("任务执行失败", logger.Error(err1))
			}

			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			err1 = s.svc.ResetNextTime(ctx, j)
			if err1 != nil {
				s.l.Error("设置下一次执行时间失败", logger.Error(err1))
			}
		}()
	}
}
