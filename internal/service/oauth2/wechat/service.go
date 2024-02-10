package wechat

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/mrhelloboy/wehook/internal/domain"
	"github.com/mrhelloboy/wehook/pkg/logger"
	"net/http"
	"net/url"
)

var redirectURI = url.PathEscape("https://meoying.com/oauth2/wechat/callback")

type Service interface {
	AuthURL(ctx context.Context, state string) (string, error)
	VerifyCode(ctx context.Context, code string) (domain.WechatInfo, error)
}

type svc struct {
	appId     string
	appSecret string
	logger    logger.Logger
}

func NewService(appId string, appSecret string, logger logger.Logger) Service {
	return &svc{
		appId:     appId,
		appSecret: appSecret,
		logger:    logger,
	}
}

func (s *svc) AuthURL(ctx context.Context, state string) (string, error) {
	const urlPattern = "https://open.weixin.qq.com/connect/qrconnect?appid=%s&redirect_uri=%s&response_type=code&scope=snsapi_login&state=%s#wechat_redirect"
	return fmt.Sprintf(urlPattern, s.appId, redirectURI, state), nil
}

func (s *svc) VerifyCode(ctx context.Context, code string) (domain.WechatInfo, error) {
	const urlPattern = "https://api.weixin.qq.com/sns/oauth2/access_token?appid=%s&secret=%s&code=%s&grant_type=authorization_code"
	resp, err := http.Get(fmt.Sprintf(urlPattern, s.appId, s.appSecret, code))
	if err != nil {
		return domain.WechatInfo{}, err
	}

	decoder := json.NewDecoder(resp.Body)
	var res AccessTokenResult
	err = decoder.Decode(&res)
	if err != nil {
		return domain.WechatInfo{}, err
	}
	if res.ErrCode != 0 {
		return domain.WechatInfo{}, fmt.Errorf("微信返回错误信息，errcode: %d, errmsg: %s", res.ErrCode, res.ErrMsg)
	}

	s.logger.Info("调用微信，拿到用户信息", logger.String("unionID", res.UnionId), logger.String("openID", res.OpenId))
	return domain.WechatInfo{
		OpenID:  res.OpenId,
		UnionID: res.UnionId,
	}, nil
}

type AccessTokenResult struct {
	ErrCode      int64  `json:"errcode"`
	ErrMsg       string `json:"errmsg"`
	AccessToken  string `json:"access_token"`
	ExpiresIn    int64  `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	OpenId       string `json:"openid"`
	Scope        string `json:"scope"`
	UnionId      string `json:"unionid"`
}
