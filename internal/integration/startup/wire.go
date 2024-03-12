//go:build wireinject

package startup

import (
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	"github.com/mrhelloboy/wehook/internal/repository"
	"github.com/mrhelloboy/wehook/internal/repository/cache"
	"github.com/mrhelloboy/wehook/internal/repository/dao"
	"github.com/mrhelloboy/wehook/internal/service"
	"github.com/mrhelloboy/wehook/internal/web"
	ijwt "github.com/mrhelloboy/wehook/internal/web/jwt"
	"github.com/mrhelloboy/wehook/ioc"
)

var (
	thirdProvider   = wire.NewSet(InitRedis, InitTestDB, InitLog)
	userSvcProvider = wire.NewSet(
		dao.NewUserDAO,
		cache.NewUserCache,
		repository.NewUserRepository,
		service.NewUserSvc)

	articlSvcProvider = wire.NewSet(
		dao.NewGormArticleDAO,
		repository.NewCachedArticleRepo,
		service.NewArticleSvc,
	)
)

func InitWebServer() *gin.Engine {
	wire.Build(
		thirdProvider,
		userSvcProvider,
		articlSvcProvider,
		cache.NewCodeCache,
		repository.NewCachedCodeRepository,
		// service 部分
		// 集成测试我们显式指定使用内存实现
		ioc.InitSMSService,

		// 指定啥也不干的 wechat service
		InitPhantomWechatService,
		service.NewCodeSvc,

		// handler 部分
		web.NewUserHandler,
		web.NewOAuth2WechatHandler,
		web.NewArticleHandler,

		ijwt.NewRedisJWTHandler,

		// gin 的中间件
		ioc.InitRateLimiterOfMiddleware,
		ioc.InitMiddleware,

		// Web 服务器
		ioc.InitGin,
	)
	// 随便返回一个
	return gin.Default()
}

func InitArticleHandler() *web.ArticleHandler {
	wire.Build(
		thirdProvider,
		dao.NewGormArticleDAO,
		repository.NewCachedArticleRepo,
		service.NewArticleSvc,
		web.NewArticleHandler,
	)
	return &web.ArticleHandler{}
}

func InitUserSvc() service.UserService {
	wire.Build(thirdProvider, userSvcProvider)
	return service.NewUserSvc(nil, nil)
}

func InitJwtHdl() ijwt.Handler {
	wire.Build(thirdProvider, ijwt.NewRedisJWTHandler)
	return ijwt.NewRedisJWTHandler(nil)
}
