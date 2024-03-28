package job

import (
	"context"
	"sync"
	"time"

	"github.com/mrhelloboy/wehook/pkg/logger"

	rlock "github.com/gotomicro/redis-lock"
	"github.com/mrhelloboy/wehook/internal/service"
)

type RankingJob struct {
	svc       service.RankingService
	timeout   time.Duration
	client    *rlock.Client
	key       string
	l         logger.Logger
	lock      *rlock.Lock
	localLock *sync.Mutex
}

func (r *RankingJob) Name() string {
	return "ranking"
}

func (r *RankingJob) Run() error {
	r.localLock.Lock()
	defer r.localLock.Unlock()
	if r.lock == nil {
		// 说明没有拿到锁，得尝试拿锁
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		// 设置一个比较短的过期时间
		lock, err := r.client.Lock(ctx, r.key, r.timeout, &rlock.FixIntervalRetry{
			Interval: time.Millisecond * 100,
			Max:      0,
		}, time.Second)
		if err != nil {
			// 这边没有拿到锁，极大概率是别其他人持有了锁
			return nil
		}
		r.lock = lock
		// 保证自己一直拿着锁
		go func() {
			r.localLock.Lock()
			defer r.localLock.Unlock()
			// 自动续约机制
			err1 := lock.AutoRefresh(r.timeout/2, time.Second)
			// 注意，执行到这里说明退出了续约机制
			// 续约失败了怎么办？
			if err1 != nil {
				// 续约失败了，就让它续约失败，争取下一次，继续抢锁
				r.l.Error("续约失败", logger.Error(err1))
			}
			r.lock = nil
		}()
	}
	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()
	return r.svc.TopN(ctx)
}

func (r *RankingJob) Close() error {
	r.localLock.Lock()
	lock := r.lock
	r.lock = nil
	r.localLock.Unlock()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	return lock.Unlock(ctx)
}

func NewRankingJob(svc service.RankingService, timeout time.Duration, client *rlock.Client, l logger.Logger) *RankingJob {
	// 根据数据量来设置，如果要是七天内的帖子数量很多，就需要设置长一些
	return &RankingJob{
		svc:       svc,
		timeout:   timeout,
		client:    client,
		key:       "rlock:cron_job:ranking",
		l:         l,
		localLock: &sync.Mutex{},
	}
}
