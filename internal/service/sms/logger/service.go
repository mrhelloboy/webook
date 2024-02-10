package logger

import (
	"context"
	"github.com/mrhelloboy/wehook/internal/service/sms"
	"go.uber.org/zap"
)

// 通过 AOP 方式来记录发送短信的日志

type Service struct {
	svc sms.Service
}

func NewService(svc sms.Service) sms.Service {
	return &Service{
		svc: svc,
	}
}

func (s *Service) Send(ctx context.Context, biz string, args []string, numbers ...string) error {
	zap.L().Debug("发送短信", zap.String("biz", biz), zap.Any("args", args))
	err := s.svc.Send(ctx, biz, args, numbers...)
	if err != nil {
		zap.L().Debug("发送短信出现异常", zap.Error(err))
	}
	return err
}
