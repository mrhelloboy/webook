// Code generated by Wire. DO NOT EDIT.

//go:generate go run github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package main

import (
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
	interactiveDAO := dao.NewGormInteractiveDAO(db)
	interactiveCache := cache.NewRedisInteractiveCache(cmdable)
	interactiveRepository := repository.NewCachedInteractiveRepo(interactiveDAO, interactiveCache, logger)
	interactiveService := service.NewInteractiveService(interactiveRepository, logger)
	articleHandler := web.NewArticleHandler(articleService, interactiveService, logger)
	engine := ioc.InitGin(v, userHandler, oAuth2WechatHandler, articleHandler)
	interactiveReadEventBatchConsumer := article3.NewInteractiveReadEventBatchConsumer(client, interactiveRepository, logger)
	v2 := ioc.NewConsumers(interactiveReadEventBatchConsumer)
	app := &App{
		web:       engine,
		consumers: v2,
	}
	return app
}
