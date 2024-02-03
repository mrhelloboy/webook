package sms

import "context"

type Service interface {
	// Send biz 很含糊的业务
	Send(ctx context.Context, biz string, args []string, numbers ...string) error
}
