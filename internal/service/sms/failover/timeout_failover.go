package failover

import (
	"context"
	"github.com/mrhelloboy/wehook/internal/service/sms"
	"sync/atomic"
)

type TimeoutFailoverSMSService struct {
	svcs []sms.Service
	idx  int32
	// 连续超时的个数
	cnt int32
	// 阈值
	// 如果连续超时的个数超过阈值，则切换到下一个服务
	threshold int32
}

func NewTimeoutFailoverSMSService(svcs ...sms.Service) sms.Service {
	return &TimeoutFailoverSMSService{
		svcs:      svcs,
		idx:       0,
		cnt:       0,
		threshold: 3,
	}
}

func (t *TimeoutFailoverSMSService) Send(ctx context.Context, tpl string, args []string, numbers ...string) error {
	idx := atomic.LoadInt32(&t.idx)
	cnt := atomic.LoadInt32(&t.cnt)
	if cnt > t.threshold {
		// 切换到下一个服务，下标后挪
		newIdx := (idx + 1) % int32(len(t.svcs))
		if atomic.CompareAndSwapInt32(&t.idx, idx, newIdx) {
			atomic.StoreInt32(&t.cnt, 0)
		}
		// else 就是出现并发，别人先切换了
		// idx = newIdx
		idx = atomic.LoadInt32(&t.idx)
	}
	svc := t.svcs[idx]
	err := svc.Send(ctx, tpl, args, numbers...)
	switch err {
	case context.DeadlineExceeded:
		atomic.AddInt32(&t.cnt, 1)
	case nil:
		// 连续状态被打断了
		atomic.StoreInt32(&t.cnt, 0)
	default:
		// 不知道什么错误
		// 可以考虑换下一个
	}
	return err
}
