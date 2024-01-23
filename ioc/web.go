package ioc

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/mrhelloboy/wehook/internal/web"
	"github.com/mrhelloboy/wehook/internal/web/middleware"
	"github.com/mrhelloboy/wehook/pkg/ginx/middlewares/ratelimit"
	"github.com/redis/go-redis/v9"
	"strings"
	"time"
)

func InitGin(mws []gin.HandlerFunc, userhdr web.Handler) *gin.Engine {
	server := gin.Default()
	server.Use(mws...)
	userhdr.RegisterRouters(server)
	return server
}

func InitMiddleware(redisClient redis.Cmdable) []gin.HandlerFunc {
	return []gin.HandlerFunc{
		// 限流
		rateLimitMiddleware(redisClient),
		// 跨域
		corsMiddleware(),
		// JWT
		jwtMiddleware(),
	}
}

func rateLimitMiddleware(redisClient redis.Cmdable) gin.HandlerFunc {
	return ratelimit.NewBuilder(redisClient, time.Second, 100).Build()
}

func corsMiddleware() gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowHeaders: []string{"Content-Type", "Authorization"},
		// 允许跨域使用的 header，否则前端无法读取 x-jwt-token
		// 前端读取 x-jwt-token 的值来配置 Authorization 头
		ExposeHeaders:    []string{"x-jwt-token"},
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
		Build()
}
