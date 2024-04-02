// Code generated by Wire. DO NOT EDIT.

//go:generate go run github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package main

import (
	"github.com/google/wire"
	"github.com/mrhelloboy/wehook/interactive/events"
	repository2 "github.com/mrhelloboy/wehook/interactive/repository"
	cache2 "github.com/mrhelloboy/wehook/interactive/repository/cache"
	dao2 "github.com/mrhelloboy/wehook/interactive/repository/dao"
	service2 "github.com/mrhelloboy/wehook/interactive/service"
	article3 "github.com/mrhelloboy/wehook/internal/events/article"
	"github.com/mrhelloboy/wehook/internal/repository"
	article2 "github.com/mrhelloboy/wehook/internal/repository/article"
	"github.com/mrhelloboy/wehook/internal/repository/cache"
	"github.com/mrhelloboy/wehook/internal/repository/dao"
	"github.com/mrhelloboy/wehook/internal/repository/dao/article"
	"github.com/mrhelloboy/wehook/internal/service"
	"github.com/mrhelloboy/wehook/internal/web"
	"github.com/mrhelloboy/wehook/internal/web/jwt"
	"github.com/mrhelloboy/wehook/ioc"
)

import (
	_ "github.com/spf13/viper/remote"
)

// Injectors from wire.go:

func InitWebServer() *App {
	cmdable := ioc.InitRedis()
	limiter := ioc.InitRateLimiterOfMiddleware(cmdable)
	handler := jwt.NewRedisJWTHandler(cmdable)
	logger := ioc.InitLogger()
	v := ioc.InitMiddleware(limiter, handler, logger)
	db := ioc.InitDB(logger)
	userDAO := dao.NewUserDAO(db)
	userCache := cache.NewUserCache(cmdable)
	userRepository := repository.NewUserRepository(userDAO, userCache)
	userService := service.NewUserSvc(userRepository, logger)
	codeCache := cache.NewCodeCache(cmdable)
	codeRepository := repository.NewCachedCodeRepository(codeCache)
	smsService := ioc.InitSMSService()
	codeService := service.NewCodeSvc(codeRepository, smsService)
	userHandler := web.NewUserHandler(userService, codeService, cmdable, handler)
	wechatService := ioc.InitOAuth2WechatService(logger)
	oAuth2WechatHandler := web.NewOAuth2WechatHandler(wechatService, userService, handler)
	authorDAO := article.NewGormArticleDAO(db)
	articleCache := cache.NewRedisArticleCache(cmdable)
	authorRepository := article2.NewCachedAuthorRepo(authorDAO, userRepository, articleCache, logger)
	client := ioc.InitKafka()
	syncProducer := ioc.NewSyncProducer(client)
	producer := article3.NewKafkaProducer(syncProducer)
	articleService := service.NewArticleSvc(authorRepository, logger, producer)
	interactiveDAO := dao2.NewGormInteractiveDAO(db)
	interactiveCache := cache2.NewRedisInteractiveCache(cmdable)
	interactiveRepository := repository2.NewCachedInteractiveRepo(interactiveDAO, interactiveCache, logger)
	interactiveService := service2.NewInteractiveService(interactiveRepository, logger)
	articleHandler := web.NewArticleHandler(articleService, interactiveService, logger)
	engine := ioc.InitGin(v, userHandler, oAuth2WechatHandler, articleHandler)
	interactiveReadEventBatchConsumer := events.NewInteractiveReadEventBatchConsumer(client, interactiveRepository, logger)
	v2 := ioc.NewConsumers(interactiveReadEventBatchConsumer)
	rankingRedisCache := cache.NewRankingRedisCache(cmdable)
	rankingLocalCache := cache.NewRankingLocalCache()
	rankingRepository := repository.NewCachedRankingRepo(rankingRedisCache, rankingLocalCache)
	rankingService := service.NewBatchRankingSrv(articleService, interactiveService, rankingRepository)
	rlockClient := ioc.InitRLockClient(cmdable)
	rankingJob := ioc.InitRankingJob(rankingService, rlockClient, logger)
	cron := ioc.InitJobs(logger, rankingJob)
	app := &App{
		web:       engine,
		consumers: v2,
		cron:      cron,
	}
	return app
}

// wire.go:

var interactiveSvcProvider = wire.NewSet(service2.NewInteractiveService, repository2.NewCachedInteractiveRepo, dao2.NewGormInteractiveDAO, cache2.NewRedisInteractiveCache)

var rankingSvcProvider = wire.NewSet(service.NewBatchRankingSrv, repository.NewCachedRankingRepo, cache.NewRankingRedisCache, cache.NewRankingLocalCache)
