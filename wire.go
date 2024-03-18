//go:build wireinject

package main

import (
	"github.com/google/wire"
	eventsArt "github.com/mrhelloboy/wehook/internal/events/article"
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

func InitWebServer() *App {
	wire.Build(
		ioc.InitDB,
		ioc.InitRedis,
		ioc.InitLogger,
		ioc.InitKafka,
		ioc.NewSyncProducer,
		ioc.NewConsumers,

		// consumer
		eventsArt.NewInteractiveReadEventConsumer,
		// producer
		eventsArt.NewKafkaProducer,

		dao.NewUserDAO, cache.NewUserCache, cache.NewCodeCache,
		daoArt.NewGormArticleDAO,
		// daoArt.NewGormReaderDAO,
		dao.NewGormInteractiveDAO,

		repository.NewUserRepository, repository.NewCachedCodeRepository,
		repository.NewCachedInteractiveRepo,
		article.NewCachedAuthorRepo,
		// article.NewCachedReaderRepo,
		cache.NewRedisInteractiveCache,
		cache.NewRedisArticleCache,
		service.NewUserSvc, service.NewCodeSvc,
		service.NewArticleSvc,
		service.NewInteractiveService,
		ioc.InitOAuth2WechatService,
		ioc.InitSMSService,
		web.NewUserHandler,
		web.NewOAuth2WechatHandler,
		web.NewArticleHandler,
		ioc.InitGin,
		myjwt.NewRedisJWTHandler,
		ioc.InitMiddleware,
		ioc.InitRateLimiterOfMiddleware,

		// 组装 App 这个结构体的所有字段
		wire.Struct(new(App), "*"),
	)
	return new(App)
}
