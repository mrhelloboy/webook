// Code generated by Wire. DO NOT EDIT.

//go:generate go run github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package startup

import (
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
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

// Injectors from wire.go:

func InitWebServer() *gin.Engine {
	cmdable := InitRedis()
	limiter := ioc.InitRateLimiterOfMiddleware(cmdable)
	handler := jwt.NewRedisJWTHandler(cmdable)
	logger := InitLog()
	v := ioc.InitMiddleware(limiter, handler, logger)
	gormDB := InitTestDB()
	userDAO := dao.NewUserDAO(gormDB)
	userCache := cache.NewUserCache(cmdable)
	userRepository := repository.NewUserRepository(userDAO, userCache)
	userService := service.NewUserSvc(userRepository, logger)
	codeCache := cache.NewCodeCache(cmdable)
	codeRepository := repository.NewCachedCodeRepository(codeCache)
	smsService := ioc.InitSMSService()
	codeService := service.NewCodeSvc(codeRepository, smsService)
	userHandler := web.NewUserHandler(userService, codeService, cmdable, handler)
	wechatService := InitPhantomWechatService(logger)
	oAuth2WechatHandler := web.NewOAuth2WechatHandler(wechatService, userService, handler)
	authorDAO := article.NewGormArticleDAO(gormDB)
	authorRepository := article2.NewCachedAuthorRepo(authorDAO)
	articleService := service.NewArticleSvc(authorRepository, logger)
	articleHandler := web.NewArticleHandler(articleService, logger)
	engine := ioc.InitGin(v, userHandler, oAuth2WechatHandler, articleHandler)
	return engine
}

func InitArticleHandler(dao2 article.AuthorDAO) *web.ArticleHandler {
	authorRepository := article2.NewCachedAuthorRepo(dao2)
	logger := InitLog()
	articleService := service.NewArticleSvc(authorRepository, logger)
	articleHandler := web.NewArticleHandler(articleService, logger)
	return articleHandler
}

func InitUserSvc() service.UserService {
	gormDB := InitTestDB()
	userDAO := dao.NewUserDAO(gormDB)
	cmdable := InitRedis()
	userCache := cache.NewUserCache(cmdable)
	userRepository := repository.NewUserRepository(userDAO, userCache)
	logger := InitLog()
	userService := service.NewUserSvc(userRepository, logger)
	return userService
}

func InitJwtHdl() jwt.Handler {
	cmdable := InitRedis()
	handler := jwt.NewRedisJWTHandler(cmdable)
	return handler
}

func InitInteractiveService() service.InteractiveService {
	gormDB := InitTestDB()
	interactiveDAO := dao.NewGormInteractiveDAO(gormDB)
	cmdable := InitRedis()
	interactiveCache := cache.NewRedisInteractiveCache(cmdable)
	logger := InitLog()
	interactiveRepository := repository.NewCachedInteractiveRepo(interactiveDAO, interactiveCache, logger)
	interactiveService := service.NewInteractiveService(interactiveRepository, logger)
	return interactiveService
}

// wire.go:

var (
	thirdProvider   = wire.NewSet(InitRedis, InitTestDB, InitLog)
	userSvcProvider = wire.NewSet(dao.NewUserDAO, cache.NewUserCache, repository.NewUserRepository, service.NewUserSvc)

	articlSvcProvider = wire.NewSet(article.NewGormArticleDAO, article2.NewCachedAuthorRepo, service.NewArticleSvc)

	interactiveSvcProvider = wire.NewSet(service.NewInteractiveService, repository.NewCachedInteractiveRepo, dao.NewGormInteractiveDAO, cache.NewRedisInteractiveCache)
)
