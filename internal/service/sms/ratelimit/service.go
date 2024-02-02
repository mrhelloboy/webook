package ratelimit

import (
	"context"
	"fmt"
	"github.com/mrhelloboy/wehook/internal/service/sms"
	"github.com/mrhelloboy/wehook/pkg/ratelimit"
)

var errLimited = fmt.Errorf("短信服务限流")

// RateLimitSMSService 添加了限流的短信服务
type RateLimitSMSService struct {
	svc     sms.Service
	limiter ratelimit.Limiter
}

func NewRateLimitSMSService(svc sms.Service, limiter ratelimit.Limiter) sms.Service {
	return &RateLimitSMSService{
		svc:     svc,
		limiter: limiter,
	}
}

func (s *RateLimitSMSService) Send(ctx context.Context, tpl string, args []string, numbers ...string) error {
	limited, err := s.limiter.Limit(ctx, "sms:limit")
	if err != nil {
		// 可以限流：保守策略（下游比较坑的时候用）
		// 不限流：下游很强，业务可用性要求很高
		return fmt.Errorf("短信服务判断是否限流出现问题，%w", err)
	}
	if limited {
		return errLimited
	}

	err = s.svc.Send(ctx, tpl, args, numbers...)
	return err
}
