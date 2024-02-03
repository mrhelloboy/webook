package auth

import (
	"context"
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"github.com/mrhelloboy/wehook/internal/service/sms"
)

type TokenSMSService struct {
	svc sms.Service
	key string
}

//func (s *TokenSMSService) GenerateSMSToken(ctx context.Context, tpl string) (string, error) {
//
//}

// Send 其中 biz 必须是线下申请的一个代表业务方的 token
func (s *TokenSMSService) Send(ctx context.Context, biz string, args []string, numbers ...string) error {

	var sc SmsClaims
	// 这里能解析成功，说明就是对应的业务方
	token, err := jwt.ParseWithClaims(biz, &sc, func(token *jwt.Token) (interface{}, error) {
		return s.key, nil
	})
	if err != nil {
		return err
	}
	if !token.Valid {
		return errors.New("token 不合法")
	}

	return s.svc.Send(ctx, sc.Tpl, args, numbers...)
}

type SmsClaims struct {
	Tpl string
	jwt.RegisteredClaims
}
