package service

import (
	"context"
	"fmt"
	"github.com/mrhelloboy/wehook/internal/repository"
	"github.com/mrhelloboy/wehook/internal/service/sms"
	"math/rand"
)

const codeTplId = "1877556"

var (
	ErrCodeSendTooMany        = repository.ErrCodeSendTooMany
	ErrCodeVerifyTooManyTimes = repository.ErrCodeVerifyTooManyTimes
)

type CodeService struct {
	repo *repository.CodeRepository
	sms  sms.Service
}

func NewCodeService(repo *repository.CodeRepository, sms sms.Service) *CodeService {
	return &CodeService{
		repo: repo,
		sms:  sms,
	}
}

// Send 发验证码
func (svc *CodeService) Send(ctx context.Context, biz string, phone string) error {
	// 存在check-do something并发安全问题场景

	// 生成一个验证码
	code := svc.generateCode()
	// 存入 Redis
	err := svc.repo.Store(ctx, biz, phone, code)
	if err != nil {
		return err
	}
	// 发送验证码
	err = svc.sms.Send(ctx, codeTplId, []string{code}, phone)
	//if err != nil {
	//	// Redis 有这个验证码，但没发送成功，用户收不到
	//	// 可能存在超时
	//	// 如果要重试，初始化的时候，传入一个可以重试的smsSvc，但不能去删掉存入 Redis 的 code
	//}
	return err
}

// Verify 验证验证码
func (svc *CodeService) Verify(ctx context.Context, biz string, phone string, inputCode string) (bool, error) {
	return svc.repo.Verify(ctx, biz, phone, inputCode)
}

func (svc *CodeService) generateCode() string {
	// 随机生成 6 位数的数字
	num := rand.Intn(1000000)
	// 不够6位的，加上前导 0
	return fmt.Sprintf("%06d", num)
}
