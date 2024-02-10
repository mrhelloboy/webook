//go:build wireinject

package main

import (
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	"github.com/mrhelloboy/wehook/internal/repository"
	"github.com/mrhelloboy/wehook/internal/repository/cache"
	"github.com/mrhelloboy/wehook/internal/repository/dao"
	"github.com/mrhelloboy/wehook/internal/service"
	"github.com/mrhelloboy/wehook/internal/web"
	myjwt "github.com/mrhelloboy/wehook/internal/web/jwt"
	"github.com/mrhelloboy/wehook/ioc"
)

func InitWebServer() *gin.Engine {
	wire.Build(
		ioc.InitDB, ioc.InitRedis,
		ioc.InitLogger,
		dao.NewUserDAO, cache.NewUserCache, cache.NewCodeCache,
		repository.NewUserRepository, repository.NewCachedCodeRepository,
		service.NewUserSvc, service.NewCodeSvc,
		ioc.InitOAuth2WechatService,
		ioc.InitSMSService,
		web.NewUserHandler,
		web.NewOAuth2WechatHandler,
		ioc.InitGin,
		myjwt.NewRedisJWTHandler,
		ioc.InitMiddleware,
		ioc.InitRateLimiterOfMiddleware,
	)
	return new(gin.Engine)
}
