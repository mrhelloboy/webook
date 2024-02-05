package ioc

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/mrhelloboy/wehook/internal/web"
	"github.com/mrhelloboy/wehook/internal/web/middleware"
	"github.com/mrhelloboy/wehook/pkg/ginx/middlewares/ratelimit"
	ratelimit2 "github.com/mrhelloboy/wehook/pkg/ratelimit"
	"strings"
	"time"
)

func InitGin(mws []gin.HandlerFunc, userhdr *web.UserHandler, oauth2WechatHdl *web.OAuth2WechatHandler) *gin.Engine {
	server := gin.Default()
	server.Use(mws...)
	userhdr.RegisterRouters(server)
	oauth2WechatHdl.RegisterRouters(server)
	return server
}

func InitMiddleware(limiter ratelimit2.Limiter) []gin.HandlerFunc {
	return []gin.HandlerFunc{
		// 限流
		rateLimitMiddleware(limiter),
		// 跨域
		corsMiddleware(),
		// JWT
		jwtMiddleware(),
	}
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

func jwtMiddleware() gin.HandlerFunc {
	return middleware.NewLoginJWTMiddlewareBuilder().
		IgnorePath("/user/signup").
		IgnorePath("/user/loginJWT").
		IgnorePath("/user/login_sms").
		IgnorePath("/user/login_sms/code/send").
		IgnorePath("/user/refresh_token").
		IgnorePath("/oauth2/wechat/authurl").
		IgnorePath("/oauth2/wechat/callback").
		Build()
}
