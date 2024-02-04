package ioc

import (
	"github.com/mrhelloboy/wehook/internal/service/oauth2/wechat"
	"os"
)

func InitOAuth2WechatService() wechat.Service {
	appId, ok := os.LookupEnv("WECHAT_APP_ID")
	if !ok {
		panic("WECHAT_APP_ID is not set")
	}

	appKey, ok := os.LookupEnv("WECHAT_APP_KEY")
	if !ok {
		panic("WECHAT_APP_KEY is not set")
	}

	return wechat.NewService(appId, appKey)
}
