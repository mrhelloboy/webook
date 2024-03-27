package ioc

import (
	"context"
	"strings"
	"time"

	"github.com/mrhelloboy/wehook/pkg/ginx/middlewares/metric"

	"github.com/fsnotify/fsnotify"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/mrhelloboy/wehook/internal/web"
	myjwt "github.com/mrhelloboy/wehook/internal/web/jwt"
	"github.com/mrhelloboy/wehook/internal/web/middleware"
	loggermw "github.com/mrhelloboy/wehook/pkg/ginx/middlewares/logger"
	"github.com/mrhelloboy/wehook/pkg/ginx/middlewares/ratelimit"
	"github.com/mrhelloboy/wehook/pkg/logger"
	ratelimit2 "github.com/mrhelloboy/wehook/pkg/ratelimit"
	"github.com/spf13/viper"
)

func InitGin(mws []gin.HandlerFunc, userhdr *web.UserHandler, oauth2WechatHdl *web.OAuth2WechatHandler, articleHdl *web.ArticleHandler) *gin.Engine {
	server := gin.Default()
	server.Use(mws...)
	userhdr.RegisterRouters(server)
	oauth2WechatHdl.RegisterRouters(server)
	articleHdl.RegisterRouters(server)
	(&web.ObservabilityHandler{}).RegisterRouters(server)
	return server
}

func InitMiddleware(limiter ratelimit2.Limiter, jwtHdl myjwt.Handler, logger logger.Logger) []gin.HandlerFunc {
	return []gin.HandlerFunc{
		// 日志
		// loggerMiddleware(logger),
		// 限流
		rateLimitMiddleware(limiter),
		// 跨域
		corsMiddleware(),
		// JWT
		jwtMiddleware(jwtHdl),
		// prometheus 监控
		prometheusMiddleware(),
	}
}

func prometheusMiddleware() gin.HandlerFunc {
	bd := metric.NewBuilder(
		"geekbang_daming",
		"webook",
		"gin_http",
		"统计 GIN 的 HTTP 接口",
		"my-instance-1")
	return bd.Build()
}

func loggerMiddleware(l logger.Logger) gin.HandlerFunc {
	bd := loggermw.NewBuilder(func(ctx context.Context, al *loggermw.AccessLog) {
		l.Debug("HTTP请求", logger.Field{Key: "al", Value: al})
	}).AllowReqBody(true).AllowRespBody(true)

	viper.OnConfigChange(func(in fsnotify.Event) {
		ok := viper.GetBool("web.logreq")
		bd.AllowReqBody(ok)
	})

	viper.OnConfigChange(func(in fsnotify.Event) {
		ok := viper.GetBool("web.logresp")
		bd.AllowRespBody(ok)
	})

	return bd.Build()
}

func rateLimitMiddleware(limiter ratelimit2.Limiter) gin.HandlerFunc {
	return ratelimit.NewBuilder(limiter).Build()
}

func corsMiddleware() gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowHeaders: []string{"Content-Type", "Authorization"},
		// 允许跨域使用的 header，否则前端无法读取 x-jwt-token
		// 前端读取 x-jwt-token 的值来配置 Authorization 头
		ExposeHeaders:    []string{"x-jwt-token", "x-refresh-token"},
		AllowCredentials: true,
		AllowOriginFunc: func(origin string) bool {
			// 开发环境允许跨域
			if strings.HasPrefix(origin, "http://localhost") {
				return true
			}
			return strings.Contains(origin, "your-company.com")
		},
		MaxAge: 1 * time.Hour,
	})
}

func jwtMiddleware(jwtHdl myjwt.Handler) gin.HandlerFunc {
	return middleware.NewLoginJWTMiddlewareBuilder(jwtHdl).
		IgnorePath("/user/signup").
		IgnorePath("/user/loginJWT").
		IgnorePath("/user/login_sms").
		IgnorePath("/user/login_sms/code/send").
		IgnorePath("/user/refresh_token").
		IgnorePath("/oauth2/wechat/authurl").
		IgnorePath("/oauth2/wechat/callback").
		IgnorePath("/test/metric").
		Build()
}
