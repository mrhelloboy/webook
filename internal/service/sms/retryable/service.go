package retryable

import (
	"context"
	"errors"
	"github.com/mrhelloboy/wehook/internal/service/sms"
)

type Service struct {
	svc      sms.Service
	retryMax int // 重试次数
}

func (s *Service) Send(ctx context.Context, biz string, args []string, numbers ...string) error {
	err := s.svc.Send(ctx, biz, args, numbers...)
	cnt := 1
	if err != nil && cnt < s.retryMax {
		err = s.svc.Send(ctx, biz, args, numbers...)
		if err == nil {
			return nil
		}
		cnt++
	}
	return errors.New("重试都失败了")
}
