package startup

import (
	"github.com/mrhelloboy/wehook/internal/service/oauth2/wechat"
	"github.com/mrhelloboy/wehook/pkg/logger"
)

// InitPhantomWechatService 没啥用的虚拟的 wechatService
func InitPhantomWechatService(l logger.Logger) wechat.Service {
	return wechat.NewService("", "", l)
}
