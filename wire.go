//go:build wireinject

package main

import (
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	"github.com/mrhelloboy/wehook/internal/repository"
	"github.com/mrhelloboy/wehook/internal/repository/article"
	"github.com/mrhelloboy/wehook/internal/repository/cache"
	"github.com/mrhelloboy/wehook/internal/repository/dao"
	daoArt "github.com/mrhelloboy/wehook/internal/repository/dao/article"
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
		daoArt.NewGormArticleDAO, daoArt.NewGormReaderDAO,
		repository.NewUserRepository, repository.NewCachedCodeRepository,
		article.NewCachedAuthorRepo, article.NewCachedReaderRepo,
		service.NewUserSvc, service.NewCodeSvc,
		service.NewArticleSvc,
		ioc.InitOAuth2WechatService,
		ioc.InitSMSService,
		web.NewUserHandler,
		web.NewOAuth2WechatHandler,
		web.NewArticleHandler,
		ioc.InitGin,
		myjwt.NewRedisJWTHandler,
		ioc.InitMiddleware,
		ioc.InitRateLimiterOfMiddleware,
	)
	return new(gin.Engine)
}
